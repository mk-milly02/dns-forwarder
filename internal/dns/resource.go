package dns

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net/netip"
)

const (
	A     = 0x01
	NS    = 0x02
	CNAME = 0x05
	SOA   = 0x06
	PTR   = 0x0c
	MX    = 0x0f
	TXT   = 0x10
	AAAA  = 0x1c
	OPT   = 0x29
)

const (
	IN = 0x01
	CS = 0x02
	CH = 0x03
	HS = 0x04
)

func GetResourceRecordType(recordType uint16) string {
	switch recordType {
	case A:
		return "A"
	case NS:
		return "NS"
	case CNAME:
		return "CNAME"
	case SOA:
		return "SOA"
	case PTR:
		return "PTR"
	case MX:
		return "MX"
	case TXT:
		return "TXT"
	case AAAA:
		return "AAAA"
	case OPT:
		return "OPT"
	default:
		return "Unknown"
	}
}

func GetResourceRecordClass(class uint16) string {
	switch class {
	case IN:
		return "IN"
	case CS:
		return "CS"
	case CH:
		return "CH"
	case HS:
		return "HS"
	default:
		return "Unknown"
	}
}

type ResourceRecord struct {
	Name       string
	RecordType uint16
	Class      uint16
	TTL        uint32
	DataLength uint16
	Data       string
}

func ParseResourceRecord(b []byte, count, offset int) (rr []ResourceRecord, newOffset int) {
	for range count {
		name, nOffset := DecodeDomainName(b, offset)
		recordType := binary.BigEndian.Uint16(b[nOffset : nOffset+2])
		class := binary.BigEndian.Uint16(b[nOffset+2 : nOffset+4])
		ttl := binary.BigEndian.Uint32(b[nOffset+4 : nOffset+8])
		dataLength := binary.BigEndian.Uint16(b[nOffset+8 : nOffset+10])
		var data string
		switch recordType {
		case A:
			ip, ok := netip.AddrFromSlice(b[nOffset+10 : nOffset+14])
			if ok {
				data = ip.String()
			}
		case NS, CNAME:
			data, _ = DecodeDomainName(b, nOffset+10)
		case AAAA:
			ip, ok := netip.AddrFromSlice(b[nOffset+10 : nOffset+26])
			if ok {
				data = ip.String()
			}
		default:
			data = hex.EncodeToString(b[nOffset+10 : nOffset+10+int(dataLength)])
		}
		rr = append(rr, ResourceRecord{name, recordType, class, ttl, dataLength, data})
		nOffset += 10 + int(dataLength)
		offset = nOffset
	}
	return rr, offset
}

func (rr ResourceRecord) Print() string {
	switch rr.RecordType {
	case A:
		return fmt.Sprintf(" ; %s\t %d\t %s\t %s\t %s\n", rr.Name, int(rr.TTL), GetResourceRecordType(rr.RecordType), GetResourceRecordClass(rr.Class), rr.Data)
	case OPT:
		return fmt.Sprintf(" ; %s\t %d\t %s\t udp: %d\t\n", rr.Name, int(rr.TTL), GetResourceRecordType(rr.RecordType), int(rr.Class))
	default:
		return fmt.Sprintf(" ; %s\t %d\t %s\t %s\n", rr.Name, int(rr.TTL), GetResourceRecordType(rr.RecordType), GetResourceRecordClass(rr.Class))
	}
}

func (rr ResourceRecord) String() string {
	recordType := fmt.Sprintf("%04x", rr.RecordType)
	class := fmt.Sprintf("%04x", rr.Class)
	ttl := fmt.Sprintf("%08x", rr.TTL)
	dataLength := fmt.Sprintf("%04x", rr.DataLength)
	data := ""
	switch rr.RecordType {
	case A:
		ip, err := netip.ParseAddr(rr.Data)
		if err != nil {
			log.Fatalf("invalid ipv4 address : %v", err)
		}
		data = hex.EncodeToString(ip.AsSlice())
	case NS, CNAME:
		data = EncodeDomainName(rr.Data)
	case AAAA:
		ip, err := netip.ParseAddr(rr.Data)
		if err != nil {
			log.Fatalf("invalid ipv6 address : %v", err)
		}
		data = hex.EncodeToString(ip.AsSlice())
	default:
		data = rr.Data
	}
	return EncodeDomainName(rr.Name) + recordType + class + ttl + dataLength + data
}

func (rr ResourceRecord) GetType() string {
	return GetResourceRecordType(rr.RecordType)
}

func NewOPTRecord(udpSize uint16) ResourceRecord {
	return ResourceRecord{
		Name:       "",
		RecordType: 41,
		Class:      udpSize,
		TTL:        0,
		DataLength: 0,
		Data:       "",
	}
}
