package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"north-polaris/internal/forwarder"
	"north-polaris/internal/server"
	"strconv"
	"time"
)

const DEFAULT_LOCAL_NAME_SERVER = "127.0.0.100"
const DEFAULT_PORT = 53

var ipAddress = flag.String("ip", DEFAULT_LOCAL_NAME_SERVER, "specified ip address for incomming requests")
var port = flag.Int("p", DEFAULT_PORT, "specified port for incomming requests")
var debug = flag.Bool("d", false, "print hex dumps")

func main() {
	flag.Parse()

	if *ipAddress == "" {
		log.Fatal("ip address is required")
	}
	if net.ParseIP(*ipAddress) == nil {
		log.Fatal("invalid ip address")
	}
	if *port < 1 || *port > 65535 {
		log.Fatal("invalid port number")
	}

	c := server.NewCache(5*time.Minute, 1*time.Minute)
	s := server.NewServer(net.IPAddr{IP: net.ParseIP(*ipAddress)}, *port, "udp")
	connection, err := s.Run()
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()
	log.Printf("listening on <<>> @%s:%s\n", *ipAddress, strconv.Itoa(*port))

	message := make([]byte, 1024)
	for {
		n, clientAddress, err := connection.ReadFromUDP(message)
		if err != nil {
			log.Printf("read error: %v\n", err)
			continue
		}
		req := message[:n]
		if *debug {
			log.Println("request message")
			fmt.Printf("%s", hex.Dump(req))
		}
		res := forwarder.HandleRequest(req, c, clientAddress)
		if *debug {
			log.Println("response message")
			fmt.Printf("%s", hex.Dump(res))
		}
		_, err = connection.WriteToUDP(res, clientAddress)
		if err != nil {
			log.Printf("write error: %v\n", err)
			continue
		}
	}
}
