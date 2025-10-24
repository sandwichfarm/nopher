package aggregates

import (
	"testing"
)

func TestParseInvoiceAmount(t *testing.T) {
	zp := &ZapProcessor{}

	tests := []struct {
		name     string
		invoice  string
		expected int64
		wantErr  bool
	}{
		{
			name:     "millibitcoin",
			invoice:  "lnbc10m1...",
			expected: 1000000, // 10m * 100,000 = 1,000,000 sats
			wantErr:  false,
		},
		{
			name:     "microbitcoin",
			invoice:  "lnbc100u1...",
			expected: 10000, // 100u * 100 = 10,000 sats
			wantErr:  false,
		},
		{
			name:     "nanobitcoin",
			invoice:  "lnbc1000n1...",
			expected: 100, // 1000n / 10 = 100 sats
			wantErr:  false,
		},
		{
			name:     "simple amount (full bitcoin)",
			invoice:  "lnbc21001...",
			expected: 2100100000000, // 21001 * 100,000,000 (no multiplier = full bitcoin)
			wantErr:  false,
		},
		{
			name:     "invalid format",
			invoice:  "invalid",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := zp.parseInvoiceAmount(tt.invoice)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInvoiceAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && amount != tt.expected {
				t.Errorf("parseInvoiceAmount() = %d, expected %d", amount, tt.expected)
			}
		})
	}
}

func TestFormatSats(t *testing.T) {
	tests := []struct {
		sats     int64
		expected string
	}{
		{0, "0 sats"},
		{100, "100 sats"},
		{999, "999 sats"},
		{1000, "1.0K sats"},
		{1500, "1.5K sats"},
		{999999, "1000.0K sats"},
		{1000000, "1.00M sats"},
		{2100000, "2.10M sats"},
	}

	for _, tt := range tests {
		result := FormatSats(tt.sats)
		if result != tt.expected {
			t.Errorf("FormatSats(%d) = %s, expected %s", tt.sats, result, tt.expected)
		}
	}
}
