package main

import (
	"net/http"

	"github.com/gorilla/mux"
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

	router.HandleFunc("/api/stream", DoVideoSignaling).Methods(http.MethodPost)
	// connect a client to the respective data channel
	router.HandleFunc("/api/connectToChat", DoTextSignaling).Methods(http.MethodPost)
	// with the proper id it can listen to

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	http.ListenAndServe(":8080", handler)
}
