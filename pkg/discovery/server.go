package discovery

import (
	"encoding/json"
	"fmt"
	"net"
)

type Packet struct {
        Op        string `json:"op"`
        Name      string `json:"name"`
        Source    string `json:"source"`
        SourceDir string `json:"source_dir"`
        Version   string `json:"version"`
        Type      string `json:"type"`
}
type Server struct {
	conn *net.UDPConn
}

func NewServer(port int) (*Server, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{conn: conn}, nil
}

func (s *Server) Listen(handler func(Packet)) error {
	buf := make([]byte, 1024)
	for {
		n, _, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			return err
		}

		var p Packet
		if err := json.Unmarshal(buf[:n], &p); err != nil {
			fmt.Printf("failed to unmarshal packet: %v\n", err)
			continue
		}

		handler(p)
	}
}

func (s *Server) Close() error {
	return s.conn.Close()
}
