package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/coder/websocket"
)

var esp32Conn *websocket.Conn

func deviceHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println("websocket accept failed", err)
		return
	}

	esp32Conn = conn

	fmt.Println("ESP32 connected")

	// Keeps connection alive by reading messages
	for {
		_, msg, err := conn.Read(context.Background())
		if err != nil {
			fmt.Println("ESP32 disconnected", err)
			esp32Conn = nil
			return
		}

		fmt.Println("ESP32:", string(msg))
	}
}

func wakeHandler(w http.ResponseWriter, r *http.Request) {
	if esp32Conn == nil {
		http.Error(w, "ESP32 not connected", http.StatusServiceUnavailable)
		return
	}

	err := esp32Conn.Write(
		context.Background(),
		websocket.MessageText,
		[]byte("wake"),
	)

	if err != nil {
		http.Error(w, "Failed to send wake command", http.StatusInternalServerError)
		return
	}

	fmt.Println(w, "wake sent")
}
func main() {
	http.HandleFunc("/device", deviceHandler)
	http.HandleFunc("/wake", wakeHandler)

	fmt.Println("BMC Oracle listening on :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
