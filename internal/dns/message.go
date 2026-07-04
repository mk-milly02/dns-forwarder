package dns

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type Message struct {
	Header        Header
	Questions     []Question
	AnswerRRs     []ResourceRecord
	AuthorityRRs  []ResourceRecord
	AdditionalRRs []ResourceRecord
}

func NewMessage(name string, header Header) Message {
	return Message{
		Header:    NewHeaderWithParams(header.id, header.flags, 1, 0, 0, 0),
		Questions: []Question{{EncodeDomainName(name), 0x01, 0x01}},
	}
}

func (m Message) BuildMessage() string {
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(m.Header.String())
	for _, q := range m.Questions {
		queryBuilder.WriteString(q.String())
	}
	for _, an := range m.AnswerRRs {
		queryBuilder.WriteString(an.String())
	}
	for _, au := range m.AuthorityRRs {
		queryBuilder.WriteString(au.String())
	}
	for _, ad := range m.AdditionalRRs {
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

func (m *Message) ParseMessage(content []byte) {
	header, offset := ParseHeader(content)
	m.Header = header
	questions, newOffset := ParseQuestion(content, int(header.queries), offset)
	m.Questions = questions
	answers, newOffset := ParseResourceRecord(content, int(header.answers), newOffset)
	m.AnswerRRs = answers
	authority, newOffset := ParseResourceRecord(content, int(header.authorities), newOffset)
	m.AuthorityRRs = authority
	additional, _ := ParseResourceRecord(content, int(header.additions), newOffset)
	m.AdditionalRRs = additional
}

func (m Message) Print() string {
	var str strings.Builder
	str.WriteString(m.Header.Print() + "\n\n")
	if len(m.Questions) == 0 {
		str.WriteString(";; QUESTIONS: 0\n")
	} else {
		str.WriteString(";; QUESTION SECTION:\n")
		for _, q := range m.Questions {
			str.WriteString(q.Print())
		}
		str.WriteString("\n\n")
	}

	if len(m.AnswerRRs) == 0 {
		str.WriteString(";; ANSWERS: 0\n")
	} else {
		str.WriteString(";; ANSWER SECTION:\n")
		for _, a := range m.AnswerRRs {
			str.WriteString(a.Print())
		}
		str.WriteString("\n")
	}

	if len(m.AuthorityRRs) == 0 {
		str.WriteString(";; AUTHORITY: 0\n")
	} else {
		str.WriteString(";; AUTHORITY SECTION:\n")
		for _, a := range m.AuthorityRRs {
			str.WriteString(a.Print())
		}
		str.WriteString("\n")
	}

	if len(m.AdditionalRRs) == 0 {
		str.WriteString(";; ADDITIONAL: 0\n\n")
	} else {
		str.WriteString(";; ADDITIONAL SECTION:\n")
		for _, a := range m.AdditionalRRs {
			str.WriteString(a.Print())
		}
		str.WriteString("\n")
	}

	return str.String()
}

func (m Message) AddAnswers(answers []ResourceRecord) {
	m.Header.answers += uint16(len(answers))
	m.AnswerRRs = append(m.AnswerRRs, answers...)
}

func (m Message) Answers(request Message) {
	m.Header.id = request.Header.id
	m.Header.flags.qr = 0x1
	m.Header.flags.opcode = request.Header.flags.opcode
	m.Header.flags.rd = request.Header.flags.rd
	m.Header.flags.ra = 0x1
	m.Questions = append(m.Questions, request.Questions...)
	m.AuthorityRRs = append(m.AuthorityRRs, request.AuthorityRRs...)
	m.AdditionalRRs = append(m.AdditionalRRs, request.AdditionalRRs...)
	m.Header.queries += uint16(len(m.Questions))
	m.Header.answers += uint16(len(m.AnswerRRs))
	m.Header.authorities += uint16(len(m.AuthorityRRs))
	m.Header.additions += uint16(len(m.AdditionalRRs))
}
