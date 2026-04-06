package parser

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantSHK int
		wantErr error
	}{
		{name: "number only", input: "100", wantSHK: 100},
		{name: "mention number only", input: "@bot_name 42", wantSHK: 42},
		{name: "plain command", input: "куса 100", wantSHK: 100},
		{name: "mention prefix", input: "@bot_name куса 42", wantSHK: 42},
		{name: "empty", input: "", wantErr: ErrEmpty},
		{name: "unknown", input: "курс 100", wantErr: ErrUnknownCommand},
		{name: "invalid number", input: "nope", wantErr: ErrInvalidNumber},
		{name: "too large", input: "101", wantErr: ErrTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input, 100)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Parse() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got.SHK != tt.wantSHK {
				t.Fatalf("SHK = %d, want %d", got.SHK, tt.wantSHK)
			}
		})
	}
}
