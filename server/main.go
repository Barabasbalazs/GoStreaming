package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/rs/cors"

	"github.com/pion/ice/v2"
	"github.com/pion/webrtc/v3"
)

type offerJson struct {
	Offer string `json:"offer"`
}

const videoFileName = "output.ivf"
const compress = false
var api *webrtc.API 

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

	router.HandleFunc("/api/stream", DoSignaling).Methods(http.MethodPost)
	
	c := cors.New(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowCredentials: true,
    })

    handler := c.Handler(router)

	http.ListenAndServe(":8080", handler)
}