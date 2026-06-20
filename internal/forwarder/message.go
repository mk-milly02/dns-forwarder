package forwarder

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type Message struct {
	Header     Header
	Question   []Question
	Answer     []ResourceRecord
	Authority  []ResourceRecord
	Additional []ResourceRecord
}

func NewMessage(name string) Message {
	return Message{
		Header:   NewHeader(),
		Question: []Question{{EncodeDomainName(name), 0x01, 0x01}},
	}
}

func (m Message) BuildQuery() string {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(m.Header.String())
	for _, q := range m.Question {
		queryBuilder.WriteString(q.String())
	}
	for _, an := range m.Answer {
		queryBuilder.WriteString(an.String())
	}
	for _, au := range m.Authority {
		queryBuilder.WriteString(au.String())
	}
	for _, ad := range m.Additional {
		queryBuilder.WriteString(ad.String())
	}
	return queryBuilder.String()
}

func (m Message) ValidateResponse(response []byte) bool {
	return hex.EncodeToString(response[:2]) == fmt.Sprintf("%04x", m.Header.id)
}

func (m Message) IsAResponse() bool {
	return m.Header.flags.qr == 0x1
}

func (m *Message) ParseQueryMessage(content []byte) {
	header, offset := ParseHeader(content)
	m.Header = header
	questions, newOffset := ParseQuestion(content, int(header.queries), offset)
	m.Question = questions
	answers, newOffset := ParseResourceRecord(content, int(header.answers), newOffset)
	m.Answer = answers
	authority, newOffset := ParseResourceRecord(content, int(header.authorities), newOffset)
	m.Authority = authority
	additional, _ := ParseResourceRecord(content, int(header.additions), newOffset)
	m.Additional = additional
}

func (m Message) Print() string {
	var str strings.Builder
	str.WriteString(m.Header.Print() + "\n")
	for _, q := range m.Question {
		str.WriteString(q.Print() + "\n")
	}
	if len(m.Answer) == 0 {
		str.WriteString(";; ANSWERS: 0\n")
	} else {
		for _, a := range m.Answer {
		str.WriteString(";; ANSWER SECTION:\n" + a.Print() + "\n")
		}
	}
	if len(m.Authority) == 0 {
		str.WriteString(";; AUTHORITY: 0\n")
	} else {
		for _, a := range m.Authority {
		str.WriteString(";; AUTHORITY SECTION:\n" + a.Print() + "\n")
		}
	}
	if len(m.Additional) == 0 {
		str.WriteString(";; ADDITIONAL: 0\n")
	} else {
		for _, a := range m.Additional {
		str.WriteString(";; ADDITIONAL SECTION:\n" + a.Print() + "\n")
		}
	}
	return str.String()
}
