package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/nathfavour/anyisland/internal/agent"
	"github.com/nathfavour/anyisland/internal/pal"
	"github.com/nathfavour/anyisland/internal/registry"
)

type BrokerRequest struct {
	Op   string `json:"op"`   // "QUERY", "SUBSCRIBE"
	Tool string `json:"tool"` // Tool name
}

type BrokerResponse struct {
	Status  string `json:"status"`  // "STABLE", "UPDATE_AVAIL", "ERROR"
	Version string `json:"version,omitempty"`
	Message string `json:"message,omitempty"`
}

type UpdateBroker struct {
	sys         pal.System
	reg         *registry.Registry
	subscribers map[string][]net.Conn
	mu          sync.Mutex
}

func NewUpdateBroker(sys pal.System, reg *registry.Registry) *UpdateBroker {
	return &UpdateBroker{
		sys:         sys,
		reg:         reg,
		subscribers: make(map[string][]net.Conn),
	}
}

func (b *UpdateBroker) Start(ctx context.Context) error {
	socketPath := b.sys.GetSocketPath()
	_ = os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}
	defer l.Close()
	os.Chmod(socketPath, 0600)

	go b.pollForUpdates(ctx)

	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				continue
			}
		}
		go b.handleConnection(conn)
	}
}

func (b *UpdateBroker) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	var req BrokerRequest
	if err := decoder.Decode(&req); err != nil {
		return
	}

	switch req.Op {
	case "QUERY":
		resp := b.checkTool(req.Tool)
		json.NewEncoder(conn).Encode(resp)
	case "SUBSCRIBE":
		b.mu.Lock()
		b.subscribers[req.Tool] = append(b.subscribers[req.Tool], conn)
		b.mu.Unlock()
		// Keep connection open for push notifications
		for {
			// Stay alive until client closes
			buf := make([]byte, 1)
			if _, err := conn.Read(buf); err != nil {
				break
			}
		}
		b.removeSubscriber(req.Tool, conn)
	}
}

func (b *UpdateBroker) checkTool(name string) BrokerResponse {
	tools, err := b.reg.ListTools()
	if err != nil {
		return BrokerResponse{Status: "ERROR", Message: err.Error()}
	}

	for _, t := range tools {
		if t.Name == name {
			// Here we would check the actual latest version
			// For now, we simulate "STABLE"
			return BrokerResponse{Status: "STABLE", Version: t.Version}
		}
	}
	return BrokerResponse{Status: "ERROR", Message: "Tool not found"}
}

func (b *UpdateBroker) pollForUpdates(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.doPoll()
		}
	}
}

func (b *UpdateBroker) doPoll() {

	tools, err := b.reg.ListTools()

	if err != nil {

		return

	}



		lm := NewLifecycleManager(b.sys)



		ag := &agent.HeuristicSynthesizer{} // Use local heuristics for background polling



	



		// Check Anyisland itself



	

	latest, available, err := lm.CheckAnyislandUpdate(context.Background(), ag)

	if err == nil && available {

		fmt.Printf("[Broker] Pulse: Anyisland update available (%s)\n", ShortCommit(latest))

		b.notifySubscribers("anyisland", latest)

	}



	for _, t := range tools {

		if t.Name == "anyisland" {

			continue

		}

		// Logic to check remote version for other tools...

		fmt.Printf("[Broker] Polling for %s updates...\n", t.Name)

	}

}


		func (b *UpdateBroker) notifySubscribers(toolName, newVersion string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	conns := b.subscribers[toolName]
	resp := BrokerResponse{Status: "UPDATE_AVAIL", Version: newVersion}
	
	for _, conn := range conns {
		json.NewEncoder(conn).Encode(resp)
	}
}

func (b *UpdateBroker) removeSubscriber(tool string, conn net.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	conns := b.subscribers[tool]
	for i, c := range conns {
		if c == conn {
			b.subscribers[tool] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}
