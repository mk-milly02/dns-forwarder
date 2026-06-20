package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"north-polaris/internal/forwarder"
	"strconv"
)

const DEFAULT_LOCAL_NAME_SERVER = "127.0.0.1"
const DEFAULT_PORT = 53

func main() {
	port := flag.Int("p", DEFAULT_PORT, "specified port for incomming requests")
	flag.Parse()

	if *port != 53 && *port <= 1023 {
		log.Fatal("use ephimeral ports")
	}

	serverAddr, err := net.ResolveUDPAddr("udp", DEFAULT_LOCAL_NAME_SERVER+":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatal("could to resolve ip to the name server", err)
	}

	connection, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatal("listen failed", err)
	}
	defer connection.Close()
	log.Printf("listening on <<>> @%s -p %s\n\n", DEFAULT_LOCAL_NAME_SERVER, strconv.Itoa(*port))

	message := make([]byte, 1024)
	for {
		n, clientAddr, err := connection.ReadFromUDP(message)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		log.Printf("message from <<>> %s\n\n", clientAddr)

		var query forwarder.Message
		query.ParseQueryMessage(message)
		fmt.Printf("%s", query.Print())

		request := query.BuildQuery()
		hex_req, err := hex.DecodeString(request)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("forwarding %s to %s\n", query.Question[0].GetName(), forwarder.GOOGLE_DNS_SERVER)

		res, err := forwarder.SendRequestTo(hex_req, forwarder.GOOGLE_DNS_SERVER)
		if err != nil {
			log.Fatal(err)
		}

		if !query.ValidateResponse(res) {
			log.Fatal("invalid response")
		}

		var response forwarder.Message
		response.ParseQueryMessage(res)
		fmt.Printf("%s", response.Print())

		_, err = connection.WriteToUDP(message[:n], clientAddr)
		if err != nil {
			log.Printf("write error: %v", err)
		}
	}
}
