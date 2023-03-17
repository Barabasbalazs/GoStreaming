package main

import (
	"log"
	"time"
	"os"
	"fmt"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"

	"github.com/gorilla/mux"
	"net/http"
	"github.com/rs/cors"

	"github.com/pion/webrtc/v3/pkg/media/ivfreader"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/ice/v2"
	"github.com/pion/randutil"
)

type offerJson struct {
	Offer string `json:"offer"`
}

const videoFileName = "output.ivf"
const compress = false
var api *webrtc.API 

// Encode encodes the input in base64
// It can optionally zip the input before encoding
func zip(in []byte) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err := gz.Write(in)
	if err != nil {
		panic(err)
	}
	err = gz.Flush()
	if err != nil {
		panic(err)
	}
	err = gz.Close()
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func encode(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	if compress {
		b = zip(b)
	}

	return base64.StdEncoding.EncodeToString(b)
}

// Decode decodes the input from base64
// It can optionally unzip the input after decoding
func decode(in string, obj interface{}) {

	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	if compress {
		b = unzip(b)
	}

	err = json.Unmarshal(b, obj)
	if err != nil {
		panic(err)
	}
}

func unzip(in []byte) []byte {
	var b bytes.Buffer
	_, err := b.Write(in)
	if err != nil {
		panic(err)
	}
	r, err := gzip.NewReader(&b)
	if err != nil {
		panic(err)
	}
	res, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return res
}

func readRTCP(rtpSender *webrtc.RTPSender) {
	rtcpBuf := make([]byte, 1500)
	for {
		// problem here!!!!
		if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
			return
		}
	}
}

func writeVideoToTrack(t *webrtc.TrackLocalStaticSample) {
	// Open a IVF file and start reading using our IVFReader
	file, err := os.Open(videoFileName)
	if err != nil {
		panic(err)
	}

	ivf, header, err := ivfreader.NewWith(file)
	if err != nil {
		panic(err)
	}
	// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
	// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
	//
	// It is important to use a time.Ticker instead of time.Sleep because
	// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
	// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
	ticker := time.NewTicker(time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1000))
	for ; true; <-ticker.C {
		frame, _, err := ivf.ParseNextFrame()
		if err != nil {
			fmt.Printf("Finish writing video track: %s ", err)
			return
		}

		if err = t.WriteSample(media.Sample{Data: frame, Duration: time.Second}); err != nil {
			fmt.Printf("Finish writing video track: %s ", err)
			return
		}
	}
}

func doSignaling (response http.ResponseWriter, request *http.Request) {

	log.Println("GET stream")

	var body offerJson

    // Try to decode the request body into the struct. If there is an error,
    // respond to the client with the error message and a 400 status code.
    err := json.NewDecoder(request.Body).Decode(&body)
    if err != nil {
        http.Error(response, err.Error(), http.StatusBadRequest)
        return
    }

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
		if err != nil {
			panic(err)
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8},
		fmt.Sprintf("video-%d", randutil.NewMathRandomGenerator().Uint32()),
		fmt.Sprintf("video-%d", randutil.NewMathRandomGenerator().Uint32()),
	)

	if err != nil {
		log.Println("Videotrack error")
		panic(err)
	}

	rtpSender, err := peerConnection.AddTrack(videoTrack)
		if err != nil {
			panic(err)
	}

	go readRTCP(rtpSender)

	go writeVideoToTrack(videoTrack)

	// get the offer from the request body
	offer := webrtc.SessionDescription{}
	decode(body.Offer, &offer)
	
	// Set the remote SessionDescription
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		panic(err)
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// Output the answer in base64 so we can paste it in browser
	// log.Println(encode(*peerConnection.LocalDescription()))

	responseOffer := offerJson{Offer: encode(*peerConnection.LocalDescription())}

	jsonResponse, jsonError := json.Marshal(responseOffer)

	if jsonError != nil {
		log.Println("Unable to encode JSON")
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(jsonResponse)
}

func main() {
	// setupWebRTC()
	settingEngine := webrtc.SettingEngine{}

	udpMux, err := ice.NewMultiUDPMuxFromPort(8443)
	if err != nil {
		panic(err)
	}

	settingEngine.SetICEUDPMux(udpMux)

	api = webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))

	router := mux.NewRouter()

	router.HandleFunc("/api/stream", doSignaling).Methods(http.MethodPost)
	
	c := cors.New(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowCredentials: true,
    })

    handler := c.Handler(router)

	http.ListenAndServe(":8080", handler)
}