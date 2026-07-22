package forwarder

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"north-polaris/internal/dns"
	"north-polaris/internal/server"
	"slices"

	"github.com/google/uuid"
)

const DEFAULT_NAME_SERVER = "198.41.0.4"
const GOOGLE_DNS_SERVER = "8.8.8.8"
const DEFAULT_PORT = "53"

// send sends a DNS request to the specified server and returns the response.
func send(req []byte, server string) ([]byte, error) {
	conn, err := net.Dial("udp", server+":"+DEFAULT_PORT)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		return nil, err
	}
	return response[:n], nil
}

// forward forwards a DNS question to the specified server and returns the answer resource records.
func forward(question dns.Question, header dns.Header) []dns.ResourceRecord {
	log.Printf("forwarding [%s - %s] to <<>> %s:%s\n", question.GetType(), question.GetName(), GOOGLE_DNS_SERVER, DEFAULT_PORT)
	request := dns.NewMessage(question, header)
	req, err := hex.DecodeString(request.BuildMessage())
	if err != nil {
		log.Printf("error decoding request: %v\n", err)
		return nil
	}
	res, err := send(req, GOOGLE_DNS_SERVER)
	if err != nil {
		log.Printf("error forwarding request: %v\n", err)
		return nil
	}
	var response dns.Message
	response.ParseMessage(res)
	if !response.IsAResponse() {
		log.Printf("invalid response for domain %s\n", question.GetName())
		return nil
	}
	for len(response.AnswerRRs) == 0 && len(response.AuthorityRRs) != 0 {
		recursive(&response, req, question.GetName())
	}
	return response.AnswerRRs
}

// recursive handles the case where the DNS response does not contain any answer resource records but contains authority resource records. 
// It attempts to find a name server from the additional resource records and forwards the query to that name server.
func recursive(m *dns.Message, query []byte, domain string) {
	if len(m.AnswerRRs) == 0 && len(m.AuthorityRRs) != 0 {
		var nServer string
		for _, rr := range m.AdditionalRRs {
			if rr.RecordType == dns.A && slices.ContainsFunc(m.AuthorityRRs, func(a dns.ResourceRecord) bool { return a.Data == rr.Name }) {
				nServer = rr.Data
				break
			}
		}
		if nServer != "" {
			log.Printf("forwarding %s to <<>> %s:%s\n", domain, nServer, DEFAULT_PORT)
			response, err := send(query, nServer)
			if err != nil {
				log.Printf("error forwarding request: %v\n", err)
			}
			if !m.ValidateResponse(response) {
				log.Printf("invalid response for domain %s\n", domain)
			}
			m.ParseMessage(response)
		}
	}
}

// HandleRequest handles a DNS request, checks the cache for answers, and forwards the request if necessary. It returns the response to the client.
func HandleRequest(req []byte, c *server.Cache, clientAddress net.Addr) (res []byte) {
	var request, response dns.Message
	request.ParseMessage(req)
	response.Answers(request)
	if len(request.AdditionalRRs) == 1 && request.AdditionalRRs[0].RecordType == 41 {
		response.AdditionalRRs = []dns.ResourceRecord{dns.NewOPTRecord(request.AdditionalRRs[0].Class)}
	}
	for _, q := range request.Questions {
		log.Printf("query [%s - %s] from <<>> %s\n", q.GetType(), q.GetName(), clientAddress.String())
		qTag := fmt.Sprintf("%s-%s", q.GetType(), q.GetName())
		if cached, err := c.GetAll(qTag); err == nil && cached != nil {
			log.Printf("found in cache: tag [%s]\n", qTag)
			response.AddAnswers(cached)
		} else {
			answers := forward(q, request.Header)
			if answers == nil {
				log.Printf("no answers found for %s\n", q.GetName())
				continue
			}
			for _, a := range answers {
				c.Put(uuid.New().String(), a)
			}
			response.AddAnswers(answers)
		}
	}
	res, err := hex.DecodeString(response.BuildMessage())
	if err != nil {
		log.Printf("error decoding response: %v\n", err)
	}
	if request.ValidateResponse(res) {
		return res
	}
	return res
}
