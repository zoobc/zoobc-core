package util

import "testing"

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
