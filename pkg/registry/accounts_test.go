package registry

import "testing"

func TestContainsOnlyDigits(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"all digits", "0123456789", true},
		{"single digit", "5", true},
		{"leading zeros", "000123", true},

		{"empty string", "", false},
		{"letters inside", "123a45", false},
		{"space inside", "12 34", false},
		{"plus sign", "+123", false},
		{"dash", "123-456", false},
		{"decimal point", "12.34", false},
		{"newline at end", "123\n", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ContainsOnlyDigits(tc.in); got != tc.want {
				t.Errorf("ContainsOnlyDigits(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func FuzzContainsOnlyDigits(f *testing.F) {
	for _, seed := range []string{"", "42", "abc", "123abc"} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		_ = ContainsOnlyDigits(s)
	})
}
