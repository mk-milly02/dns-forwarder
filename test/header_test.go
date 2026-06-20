package forwarder_test

import (
	"north-polaris/internal/forwarder"
	"testing"
)

func TestFlags_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		qr     uint8
		opcode uint8
		aa     uint8
		tc     uint8
		rd     uint8
		ra     uint8
		z      uint8
		rcode  uint8
		want   string
	}{
		{
			name:   "H",
			qr:     0,
			opcode: 0,
			aa:     0,
			tc:     0,
			rd:     1,
			ra:     1,
			z:      0,
			rcode:  0,
			want:   "0000000110000000",
		},
		{
			name:   "He",
			qr:     1,
			opcode: 0,
			aa:     0,
			tc:     0,
			rd:     0,
			ra:     0,
			z:      0,
			rcode:  0,
			want:   "1000000000000000",
		},
		{
			name:   "Li",
			qr:     1,
			opcode: 0,
			aa:     0,
			tc:     0,
			rd:     1,
			ra:     1,
			z:      0,
			rcode:  0,
			want:   "1000000110000000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := forwarder.NewFlags(tt.qr, tt.opcode, tt.aa, tt.tc, tt.rd, tt.ra, tt.z, tt.rcode)
			got := f.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeader_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		header forwarder.Header
		want   string
	}{
		{
			name:   "Be",
			header: forwarder.NewHeaderWithParams(uint16(33395), forwarder.NewFlags(0, 0, 0, 0, 1, 1, 0, 0), 1, 1, 0, 0),
			want:   "827301800001000100000000",
		},
		{
			name:   "B",
			header: forwarder.NewHeaderWithParams(uint16(63506), forwarder.NewFlags(1, 0, 0, 0, 1, 1, 0, 0), 1, 10, 0, 0),
			want:   "f81281800001000a00000000",
		},
		{
			name:   "C",
			header: forwarder.NewHeaderWithParams(uint16(33395), forwarder.NewFlags(0, 0, 0, 0, 1, 0, 0, 0), 1, 10, 0, 1),
			want:   "827301000001000a00000001",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.header
			got := h.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
