package main

import (

	"context"

	"fmt"

	"log"

	"os"

	"os/signal"

	"syscall"



	"github.com/nathfavour/anyisland/internal/cli"

	"github.com/nathfavour/anyisland/internal/pal"

	"github.com/nathfavour/anyisland/internal/registry"

	"github.com/nathfavour/anyisland/pkg/discovery"

)



func main() {

	sys, err := pal.New()

	if err != nil {

		log.Fatalf("failed to init pal: %v", err)

	}



	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()



	// Setup Signal Handling

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	go func() {

		for sig := range sigChan {

			switch sig {

			case syscall.SIGHUP:

				fmt.Println("Received SIGHUP, performing hot-swap...")

				lm := cli.NewLifecycleManager(sys)

				if err := lm.HotSwap(); err != nil {

					fmt.Printf("Hot-swap failed: %v\n", err)

				}

			case syscall.SIGINT, syscall.SIGTERM:

				fmt.Println("Shutting down daemon...")

				cancel()

			}

		}

	}()



	reg, err := registry.Open(sys.GetIslandDir())

	if err != nil {

		log.Fatalf("failed to open registry: %v", err)

	}

	defer reg.Close()



	// Start Local Update Broker (UDS)

	broker := cli.NewUpdateBroker(sys, reg)

	go func() {

		if err := broker.Start(ctx); err != nil {

			log.Printf("Broker error: %v", err)

		}

	}()



	srv, err := discovery.NewServer(1995)

	if err != nil {

		log.Fatalf("failed to start discovery server: %v", err)

	}

	defer srv.Close()



	fmt.Println("Anyisland Daemon listening on UDP :1995 and UDS :anyisland.sock")



	go func() {

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

			log.Printf("server error: %v", err)

		}

	}()



	<-ctx.Done()

}


