package dns

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

type Question struct {
	name       string
	recordType uint16
	class      uint16
}

func (q Question) Print() string {
	return fmt.Sprintf(" ; %s,\t \ttype: %s, \tclass: %s", q.name, GetResourceRecordType(q.recordType), GetResourceRecordClass(q.class))
}

func (q Question) String() string {
	return fmt.Sprintf("%s%04x%04x", EncodeDomainName(q.name), q.recordType, q.class)
}

func (q Question) GetName() string {
	return q.name
}

func (q Question) GetType() string {
	return GetResourceRecordType(q.recordType)
}

func ParseQuestion(b []byte, qcount, offset int) ([]Question, int) {
	var questions []Question
	for range qcount {
		name, newOffset := DecodeDomainName(b, offset)
		offset = newOffset
		recordType := binary.BigEndian.Uint16(b[offset : offset+2])
		class := binary.BigEndian.Uint16(b[offset+2 : offset+4])
		questions = append(questions, Question{name, recordType, class})
		offset += 4
	}
	return questions, offset
}

func EncodeDomainName(name string) string {
	var encoded strings.Builder
	labels := strings.SplitSeq(name, ".")
	for l := range labels {
		fmt.Fprintf(&encoded, "%02x", len(l))
		encoded .WriteString(hex.EncodeToString([]byte(l)))
	}
	fmt.Fprintf(&encoded, "%02x", 0)
	return encoded.String()
}

func DecodeDomainName(b []byte, offset int) (string, int) {
	var name strings.Builder
	isCompressed := false
	cOffset := offset
	for {
		if b[cOffset] == 0xc0 {
			if !isCompressed {
				isCompressed = true
				offset += 2
			}
			cOffset = int(b[cOffset+1])
		}
		length := int(b[cOffset])
		if !isCompressed {
			offset += length + 1
		}
		if length == 0 {
			break
		}
		name .WriteString(string(b[cOffset+1 : cOffset+1+length]))
		cOffset += length + 1
		if b[cOffset] != 0 {
			name .WriteString(".")
		}
	}
	return name.String(), offset
}