package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"north-polaris/forwarder"
	"strconv"
)

const DEFAULT_NAME_SERVER = "127.0.0.1"
const DEFAULT_PORT = 53

func main() {
	port := flag.Int("p", DEFAULT_PORT, "specified port for incomming requests")
	flag.Parse()

	if *port != 53 && *port <= 1023 {
		log.Fatal("use ephimeral ports")
	}

	serverAddr, err := net.ResolveUDPAddr("udp", DEFAULT_NAME_SERVER+":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatal("could to resolve ip to the name server", err)
	}

	connection, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatal("listen failed", err)
	}
	defer connection.Close()
	log.Printf("listening on <<>> @%s -p %s\n\n", DEFAULT_NAME_SERVER, strconv.Itoa(*port))

	message := make([]byte, 1024)
    for {
        n, clientAddr, err := connection.ReadFromUDP(message)
        if err != nil {
            log.Printf("read error: %v", err)
            continue
        }

        log.Printf("got message from <<>> %s\n\n", clientAddr)
 
		var query forwarder.Message
		query.ParseQueryMessage(message)
		fmt.Printf("%s", query.Print())

        _, err = connection.WriteToUDP(message[:n], clientAddr)
        if err != nil {
            log.Printf("write error: %v", err)
        }
    }
}