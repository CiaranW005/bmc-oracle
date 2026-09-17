package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/coder/websocket"
	"github.com/joho/godotenv"
)

var (
	esp32Conn *websocket.Conn
	connMu    sync.RWMutex

	deviceToken  string
	wakeAPIToken string
)

func deviceHandler(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")

	if auth != "Bearer "+deviceToken {
		http.Error(w, "Unauthorised", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Println("websocket accept failed", err)
		return
	}

	connMu.Lock()
	esp32Conn = conn
	connMu.Unlock()

	fmt.Println("ESP32 connected")

	// Keeps connection alive by reading messages
	for {
		_, msg, err := conn.Read(context.Background())
		if err != nil {
			fmt.Println("ESP32 disconnected", err)

			connMu.Lock()
			if esp32Conn == conn {
				esp32Conn = nil
			}
			connMu.Unlock()
			return
		}

		fmt.Println("ESP32:", string(msg))
	}
}

// Function that handles a post command to send command to ESP32 to wake PC
func wakeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	auth := r.Header.Get("Authorization")

	if auth != "Bearer "+wakeAPIToken {
		http.Error(w, "Unauthorised", http.StatusUnauthorized)
		return
	}

	connMu.RLock()
	conn := esp32Conn
	connMu.RUnlock()

	if conn == nil {
		http.Error(w, "ESP32 not connected", http.StatusServiceUnavailable)
		return
	}

	err := conn.Write(
		context.Background(),
		websocket.MessageText,
		[]byte("wake"),
	)

	if err != nil {
		http.Error(w, "Failed to send wake command", http.StatusInternalServerError)
		return
	}

	fmt.Println("wake sent")
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found using enviroment variables")
	}

	deviceToken = os.Getenv("ESP32_DEVICE_TOKEN")

	if deviceToken == "" {
		log.Fatal("ESP32_DEVICE_TOKEN is not set")
	}

	wakeAPIToken = os.Getenv("WAKE_API_TOKEN")

	if wakeAPIToken == "" {
		log.Fatal("WAKE_API_TOKEN is not set")
	}

	http.HandleFunc("/device", deviceHandler)
	http.HandleFunc("/wake", wakeHandler)

	fmt.Println("BMC Oracle listening on :8080")

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
