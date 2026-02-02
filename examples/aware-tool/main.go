package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Packet struct {
	Op      string `json:"op"`
	Name    string `json:"name"`
	Source  string `json:"source"`
	Version string `json:"version"`
	Type    string `json:"type"`
}

func main() {
	fmt.Println("Anyisland-aware tool starting...")

	p := Packet{
		Op:      "REGISTER",
		Name:    "aware-tool",
		Source:  "github.com/nathfavour/anyisland-example",
		Version: "v1.0.0",
		Type:    "binary",
	}

	data, _ := json.Marshal(p)

	conn, err := net.Dial("udp", "localhost:1995")
	if err != nil {
		fmt.Printf("failed to connect to anyisland daemon: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		fmt.Printf("failed to send heartbeat: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Registration heartbeat sent to anyisland daemon.")
}
