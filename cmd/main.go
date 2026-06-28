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

var port = flag.Int("p", DEFAULT_PORT, "specified port for incomming requests")
var debug = flag.Bool("d", false, "enable debug mode")

func main() {
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
	log.Printf("listening on <<>> @%s:%s\n\n", DEFAULT_LOCAL_NAME_SERVER, strconv.Itoa(*port))

	message := make([]byte, 1024)
	for {
		_, clientAddr, err := connection.ReadFromUDP(message)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		var query forwarder.Message
		query.ParseQueryMessage(message)
		if debug != nil && *debug {
			fmt.Printf("%s", query.Print())
		}
		for _, q := range query.Question {
			log.Printf("query[%s] %s from <<>> %s\n\n", q.GetType(), q.GetName(), clientAddr)
		}

		request := query.BuildQuery()
		hex_req, err := hex.DecodeString(request)
		if err != nil {
			log.Fatal(err)
		}
		for _, q := range query.Question {
			log.Printf("forwarded %s to <<>> %s:%s\n\n", q.GetName(), forwarder.GOOGLE_DNS_SERVER, forwarder.DEFAULT_PORT)
		}

		res, err := forwarder.SendRequestTo(hex_req, forwarder.GOOGLE_DNS_SERVER)
		if err != nil {
			log.Fatal(err)
		}

		if !query.ValidateResponse(res) {
			log.Fatal("invalid response")
		}

		var response forwarder.Message
		response.ParseQueryMessage(res)
		if debug != nil && *debug {
			fmt.Printf("%s", response.Print())
		}
		for _, r := range response.Answer {
			log.Printf("response[%s] %s is %s\n\n", r.GetType(), r.Name, r.Data)
		}

		_, err = connection.WriteToUDP(res, clientAddr)
		if err != nil {
			log.Printf("write error: %v", err)
		}
	}
}
