package normalize

import "testing"

func TestWord(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"istanbul", "istanbul"},
		{"İSTANBUL", "istanbul"},
		{"IŞIK", "ışık"},
		{"Iğdır", "ığdır"},
		{"ÇİÇEK", "çiçek"},
		{"ğü", "ğü"}, // 'g' + combining breve should become a single 'ğ'
		// TODO: add cases you run into in tr_full.txt
	}
	for _, tt := range tests {
		if got := Word(tt.in); got != tt.want {
			t.Errorf("Word(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
