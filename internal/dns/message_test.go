package dns_test

import (
	"north-polaris/internal/dns"
	"testing"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name     string
		header   dns.Header
		question dns.Question
		want     string
	}{
		{
			name:   "baidu.com",
			header: dns.NewHeaderWithParams(0x0002, dns.NewFlags(0, 0, 0, 0, 0, 1, 0, 0), 1, 0, 0, 0),
			question: dns.NewQuestion("baidu.com", 1, 1),
			want:   "00020080000100000000000005626169647503636f6d0000010001",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := dns.NewMessage(tt.question, tt.header)
			got := m.BuildMessage()
			if got != tt.want {
				t.Errorf("NewMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
