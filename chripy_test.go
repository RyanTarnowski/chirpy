package main

import (
	"testing"
)

func TestProfanityScrubber(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "This is a kerfuffle opinion I need to share with the world",
			expected: "This is a **** opinion I need to share with the world",
		},
		{
			input:    "This is a Kerfuffle opinion I need to share with the world",
			expected: "This is a **** opinion I need to share with the world",
		},
		{
			input:    "This is a FORNAX opinion I need to share with the world",
			expected: "This is a **** opinion I need to share with the world",
		},
		{
			input:    "kerfuffle sharbert fornax",
			expected: "**** **** ****",
		},
		{
			input:    "kerfuffle kerfuffle kerfuffle",
			expected: "**** **** ****",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, c := range cases {
		actual := ProfanityScrubber(c.input)

		if actual != c.expected {
			t.Errorf("Actual and expected do not match: \n'%v' \nvs\n'%v'", actual, c.expected)
		}
	}
}
