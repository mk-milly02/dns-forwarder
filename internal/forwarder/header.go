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

type Header struct {
	id          uint16
	flags       Flags
	queries     uint16
	answers     uint16
	authorities uint16
	additions   uint16
}

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

func (f Flags) String() string {
	return fmt.Sprintf("%016b", (
		uint16(f.qr)<<15)|(uint16(f.opcode)<<11)|(uint16(f.aa)<<10)|(uint16(f.tc)<<9)|
		(uint16(f.rd)<<8)|(uint16(f.ra)<<7)|(uint16(f.z)<<4)|uint16(f.rcode))
}

func parseFlags(f uint16) Flags {
	return Flags {
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

func (f *Flags) print() string {
	var str strings.Builder
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

func (h Header) String() string {
	f, err := strconv.ParseUint(h.flags.String(), 2, 16)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%04x%04x%04x%04x%04x%04x", h.id, f, h.queries, h.answers, h.authorities, h.additions)
}

func (h Header) Print() string {
	return fmt.Sprintf(";; -->>HEADER<<-- opcode: %s, id: %d\n;; flags: %s, QUERY: %d, ANSWER: %d, AUTHORIRY: %d, ADDITIONAL: %d",
		GetOpcode(h.flags.opcode), h.id, h.flags.print(), h.queries, int(h.answers), int(h.authorities), int(h.additions))
}

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

func newTransactionID() uint16 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return uint16(r.Intn(16))
}

func NewHeader() Header {
	return Header{
		id:      newTransactionID(),
		flags:   Flags{},
		queries: 0x01,
	}
}

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
