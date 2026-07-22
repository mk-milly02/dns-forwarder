package dns_test

import (
	"north-polaris/internal/dns"
	"testing"
)

func TestEncodeDomainName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "example.com", want: "076578616d706c6503636f6d00"},
		{name: "www.example.com", want: "03777777076578616d706c6503636f6d00"},
		{name: "subdomain.example.com", want: "09737562646f6d61696e076578616d706c6503636f6d00"},
		{name: "meta.com", want: "046d65746103636f6d00"},
		{name: "baidu.com", want: "05626169647503636f6d00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dns.EncodeDomainName(tt.name)
			if got != tt.want {
				t.Errorf("EncodeDomainName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuestion_String(t *testing.T) {
	tests := []struct {
		name string
		question dns.Question
		want string
	}{
		{
			name: "baidu.com", 
			question: dns.NewQuestion("baidu.com", 1, 1),
			want: "05626169647503636f6d0000010001",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.question.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
