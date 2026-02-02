package main

import (
	"fmt"
	"log"

	"github.com/nathfavour/anyisland/internal/pal"
	"github.com/nathfavour/anyisland/internal/registry"
	"github.com/nathfavour/anyisland/pkg/discovery"
)

func main() {
	sys, err := pal.New()
	if err != nil {
		log.Fatalf("failed to init pal: %v", err)
	}

	reg, err := registry.Open(sys.GetIslandDir())
	if err != nil {
		log.Fatalf("failed to open registry: %v", err)
	}
	defer reg.Close()

	srv, err := discovery.NewServer(1995)
	if err != nil {
		log.Fatalf("failed to start discovery server: %v", err)
	}
	defer srv.Close()

	fmt.Println("Anyisland Daemon listening on UDP :1995")

	err = srv.Listen(func(p discovery.Packet) {
		fmt.Printf("Received packet: %+v\n", p)
		if p.Op == "REGISTER" {
			err := reg.RegisterTool(registry.Tool{
				Name:    p.Name,
				Source:  p.Source,
				Version: p.Version,
				Type:    p.Type,
			})
			if err != nil {
				fmt.Printf("failed to register tool: %v\n", err)
			} else {
				fmt.Printf("Registered tool: %s\n", p.Name)
			}
		}
	})

	if err != nil {
		log.Fatalf("server error: %v", err)
	}
}
