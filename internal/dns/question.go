package dns

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// Question represents a DNS question section, which contains the domain name being queried, 
// the type of record being requested, and the class of the query.
type Question struct {
	name       string
	recordType uint16
	class      uint16
}

// Print returns a human-readable string representation of the Question struct, showing the domain name,
// record type, and class of the query.
func (q Question) Print() string {
	return fmt.Sprintf(" ; %s,\t \ttype: %s, \tclass: %s", q.name, GetResourceRecordType(q.recordType), GetResourceRecordClass(q.class))
}

// String returns a string representation of the Question struct, showing the encoded domain name,
// record type, and class in hexadecimal format.
func (q Question) String() string {
	return fmt.Sprintf("%s%04x%04x", EncodeDomainName(q.name), q.recordType, q.class)
}

// GetName returns the domain name of the Question struct.
func (q Question) GetName() string {
	return q.name
}

// GetType returns the record type of the Question struct as a string representation, 
// using the GetResourceRecordType function to convert the numeric type to a human-readable format.
func (q Question) GetType() string {
	return GetResourceRecordType(q.recordType)
}

// ParseQuestion parses the question section of a DNS message from the provided byte slice, starting at the given offset,
// and returns a slice of Question structs along with the new offset after parsing.
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

// EncodeDomainName encodes a domain name into the DNS message format, where each label is prefixed with its length in bytes,
// and the entire name is terminated with a zero-length label.
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

// DecodeDomainName decodes a domain name from the DNS message format, handling both uncompressed and compressed names.
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
		name.WriteString(string(b[cOffset+1 : cOffset+1+length]))
		cOffset += length + 1
		if b[cOffset] != 0 {
			name.WriteString(".")
		}
	}
	return name.String(), offset
}

// NewQuestion creates a new Question struct with the specified domain name, record type, and class.
func NewQuestion(name string, recordType, class uint16) Question {
	return Question{name, recordType, class}
}