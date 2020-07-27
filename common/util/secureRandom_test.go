package util

import (
	"strings"
	"testing"
)

func TestGetSecureRandom(t *testing.T) {
	tests := []struct {
		name string
	}{
		// Add test cases.
		{
			name: "TestGetSecureRandom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newRandom := GetSecureRandom()
			if got := GetSecureRandom(); got == newRandom {
				t.Errorf("GetSecureRandom() want different value between %v and %v", got, newRandom)
			}
		})
	}
}

func TestGetSecureRandomSeed(t *testing.T) {
	t.Run("GetSecureRandomSeed - Length 24 words", func(t *testing.T) {
		seed := GetSecureRandomSeed()
		if len(strings.Split(seed, " ")) != 24 {
			t.Errorf("Expect 24 word random seed, but got: %d", len(strings.Split(seed, " ")))
		}
	})
}
