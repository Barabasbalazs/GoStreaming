package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/pion/webrtc/v3"
)

func messageStream(peerConnection *webrtc.PeerConnection) {
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			for range time.Tick(time.Second * 3) {
				if err := d.SendText(time.Now().String()); err != nil {
					panic(err)
				}
			}
		})
	})
}

func DoTextSignaling(response http.ResponseWriter, request *http.Request) {
	log.Println("GET textSignaling")

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

	go messageStream(peerConnection)

	// get the offer from the request body
	offer := webrtc.SessionDescription{}
	Decode(body.Offer, &offer)

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

	responseOffer := offerJson{Offer: Encode(*peerConnection.LocalDescription())}

	jsonResponse, jsonError := json.Marshal(responseOffer)

	if jsonError != nil {
		log.Println("Unable to encode JSON")
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	response.Write(jsonResponse)
}
