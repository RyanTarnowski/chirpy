package auth

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	cases := []struct {
		input    string
		input2   string
		expected bool
	}{
		{
			input:    "TestPassWord1",
			input2:   "TestPassWord1",
			expected: true,
		},
		{
			input:    "TestPassWord1",
			input2:   "WrongPassWord1",
			expected: false,
		},
	}

	for _, c := range cases {
		hash, err := HashPassword(c.input)
		if err != nil {
			t.Errorf("Failed during password hashing %v", err)
		}

		result, err := CheckPasswordHash(c.input2, hash)
		if err != nil {
			t.Errorf("Failed during password compare %v", err)
		}

		if result != c.expected {
			t.Errorf("Actual and expected do not match: \n'%v' \nvs\n'%v'", result, c.expected)
		}
	}
}
