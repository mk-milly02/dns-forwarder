package server

import (
	"log"
	"net"
	"strconv"
)

// server struct represents a DNS server with its IP address, port, and protocol (e.g., UDP).
type server struct {
	ipAddress net.IPAddr
	port      int
	protocol  string
}

// NewServer creates a new server instance with the specified IP address, port, and protocol.
func NewServer(ipAddress net.IPAddr, port int, protocol string) *server {
	return &server{
		ipAddress: ipAddress,
		port:      port,
		protocol:  protocol,
	}
}

// Run starts the server, listens for incoming connections, and initializes the cache.
func (s *server) Run() (*net.UDPConn, error) {
	serverAddress, err := net.ResolveUDPAddr(s.protocol, s.ipAddress.IP.String()+":"+strconv.Itoa(s.port))
	if err != nil {
		log.Println("could not resolve ip to the name server", err)
		return nil, err
	}
	connection, err := net.ListenUDP(s.protocol, serverAddress)
	if err != nil {
		log.Println("listen failed", err)
		return nil, err
	}
	return connection, nil
}