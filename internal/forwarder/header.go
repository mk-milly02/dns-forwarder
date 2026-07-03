package forwarder

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)
 // Header represents the DNS message header, which contains metadata about the DNS message, 
 // such as the transaction ID, flags, and counts of various sections (queries, answers, authorities, and additional records).
type Header struct {
	id          uint16
	flags       Flags
	queries     uint16
	answers     uint16
	authorities uint16
	additions   uint16
}

// Flags represents the flags in the DNS message header, which control various aspects of the DNS protocol, 
// such as query/response type, operation code, and response codes.
type Flags struct {
	qr     uint8
	opcode uint8
	aa     uint8
	tc     uint8
	rd     uint8
	ra     uint8
	z      uint8
	rcode  uint8
}

// String returns a string representation of the Flags struct, showing the binary representation of the flags.
func (f Flags) String() string {
	return fmt.Sprintf("%016b", (uint16(f.qr)<<15)|(uint16(f.opcode)<<11)|(uint16(f.aa)<<10)|(uint16(f.tc)<<9)|
		(uint16(f.rd)<<8)|(uint16(f.ra)<<7)|(uint16(f.z)<<4)|uint16(f.rcode))
}

// Print returns a human-readable string representation of the Header struct, showing the opcode, 
// transaction ID, flags, and counts of various sections.
func parseFlags(f uint16) Flags {
	return Flags{
		qr:     uint8((f >> 15) & 0x1),
		opcode: uint8((f >> 11) & 0xF),
		aa:     uint8((f >> 10) & 0x1),
		tc:     uint8((f >> 9) & 0x1),
		rd:     uint8((f >> 8) & 0x1),
		ra:     uint8((f >> 7) & 0x1),
		z:      uint8((f >> 4) & 0x7),
		rcode:  uint8(f & 0xF),
	}
}

// print returns a string representation of the Flags struct, showing which flags are set (aa, tc, rd, ra).
func (f *Flags) print() string {
	var str strings.Builder
	if f.qr == 0x0 {
		str.WriteString("qr ")
	}
	if f.aa != 0x0 {
		str.WriteString("aa ")
	}
	if f.tc != 0x0 {
		str.WriteString("tc ")
	}
	if f.rd != 0x0 {
		str.WriteString("rd ")
	}
	if f.ra != 0x0 {
		str.WriteString("ra ")
	}
	return str.String()
}

// NewFlags creates a new Flags struct with the specified values for each flag.
func NewFlags(qr, opcode, aa, tc, rd, ra, z, rcode uint8) Flags {
	return Flags{
		qr:     qr,
		opcode: opcode,
		aa:     aa,
		tc:     tc,
		rd:     rd,
		ra:     ra,
		z:      z,
		rcode:  rcode,
	}
}

// GetOpcode returns a string representation of the opcode value, which indicates the 
// type of DNS operation (QUERY, IQUERY, STATUS, or EXPERIMENTAL).
func GetOpcode(opcode uint8) string {
	switch opcode {
	case 0x0:
		return "QUERY"
	case 0x1:
		return "IQUERY"
	case 0x2:
		return "STATUS"
	default:
		return "EXPERIMENTAL"
	}
}

// String returns a string representation of the Header struct, showing the transaction ID, flags, and counts of various sections in hexadecimal format.
func (h Header) String() string {
	f, err := strconv.ParseUint(h.flags.String(), 2, 16)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%04x%04x%04x%04x%04x%04x", h.id, f, h.queries, h.answers, h.authorities, h.additions)
}

// Print returns a human-readable string representation of the Header struct, showing the opcode, transaction ID, flags, and counts of various sections.
func (h Header) Print() string {
	return fmt.Sprintf(";; -->>HEADER<<-- opcode: %s, id: %d\n;; flags: %s, QUERY: %d, ANSWER: %d, AUTHORIRY: %d, ADDITIONAL: %d",
		GetOpcode(h.flags.opcode), h.id, h.flags.print(), h.queries, int(h.answers), int(h.authorities), int(h.additions))
}

// ParseHeader parses a byte slice representing a DNS message header and returns a Header struct 
// along with the number of bytes consumed (12 bytes for the header).
func ParseHeader(b []byte) (Header, int) {
	if len(b) < 12 {
		log.Fatal("invalid DNS header")
	}
	return Header{
		id:          binary.BigEndian.Uint16(b[:2]),
		flags:       parseFlags(binary.BigEndian.Uint16(b[2:4])),
		queries:     binary.BigEndian.Uint16(b[4:6]),
		answers:     binary.BigEndian.Uint16(b[6:8]),
		authorities: binary.BigEndian.Uint16(b[8:10]),
		additions:   binary.BigEndian.Uint16(b[10:12]),
	}, 12
}

// ValidateResponse checks if the response message has the same transaction ID as the original query,
func newTransactionID() uint16 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return uint16(r.Intn(16))
}

// NewHeader creates a new Header struct with a random transaction ID, default flags, and one query.
func NewHeader() Header {
	return Header{
		id:      newTransactionID(),
		flags:   Flags{},
		queries: 0x01,
	}
}

// NewHeaderWithParams creates a new Header struct with the specified parameters, allowing for custom transaction ID, flags, and counts of various sections.
func NewHeaderWithParams(id uint16, flags Flags, queries, answers, authorities, additions uint16) Header {
	return Header{
		id:          id,
		flags:       flags,
		queries:     queries,
		answers:     answers,
		authorities: authorities,
		additions:   additions,
	}
}

// AddQuery increments the number of queries in the header by 1. This is used when adding a new question to the DNS message.
func (h *Header) AddQuery() {
	h.queries++
}

// RemoveQuery decrements the number of queries in the header by 1. This is used when removing a question from the DNS message.
func (h *Header) RemoveQuery() {
	if h.queries > 0 {
		h.queries--
	}
}

// AddAnswer increments the number of answers in the header by 1. This is used when adding a new answer to the DNS message.
func (h *Header) AddAnswer() {
	h.answers++
}