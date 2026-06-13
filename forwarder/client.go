package forwarder

import (
	"net"
)

const DEFAULT_NAME_SERVER = "198.41.0.4"
const GOOGLE_DNS_SERVER = "8.8.8.8"
const DEFAULT_PORT = "53"

func SendRequest(query []byte) ([]byte, error) {
	conn, err := net.Dial("udp", DEFAULT_NAME_SERVER+":"+DEFAULT_PORT)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = conn.Write(query)
	if err != nil {
		return nil, err
	}
	response := make([]byte, 512)
	_, err = conn.Read(response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func SendRequestTo(query []byte, nameServer string) ([]byte, error) {
	conn, err := net.Dial("udp", nameServer+":"+DEFAULT_PORT)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = conn.Write(query)
	if err != nil {
		return nil, err
	}
	response := make([]byte, 512)
	_, err = conn.Read(response)
	if err != nil {
		return nil, err
	}
	return response, nil
}