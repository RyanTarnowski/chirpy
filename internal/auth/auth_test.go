package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestJWTs(t *testing.T) {
	userId, _ := uuid.Parse("84ad67f7-4da1-4fcc-a91e-68af9261d45a")

	cases := []struct {
		userId          uuid.UUID
		secret          string
		validate_secret string
		expires_in      time.Duration
		expected        bool
	}{
		{
			userId:          userId,
			secret:          "mySecret",
			validate_secret: "mySecret",
			expires_in:      2 * time.Minute,
			expected:        true,
		},
		{
			userId:          userId,
			secret:          "mySecret",
			validate_secret: "mySecret",
			expires_in:      1,
			expected:        false,
		},
		{
			userId:          userId,
			secret:          "mySecret",
			validate_secret: "wrongSecret",
			expires_in:      2 * time.Minute,
			expected:        false,
		},
	}

	for _, c := range cases {
		token, err := MakeJWT(c.userId, c.secret, c.expires_in)
		if err != nil {
			t.Errorf("Failed to make JWT: %v", err)
			continue
		}

		token_userId, err := ValidateJWT(token, c.validate_secret)
		if err != nil {
			if c.expected == true {
				t.Errorf("Failed to validate JWT: %v", err)
				continue
			}
		}

		if c.userId != token_userId && c.expected == true {
			t.Errorf("Actual and expected userIDs do not match: \n%v \vvs\n %v", c.userId, token_userId)
		}
	}
}

func TestGetBearerToken(t *testing.T) {
	token := "Ea/+QMPauUxzmhi+ZwusDlFTAl9/17dAMgWw723hibWt5YknmpvZh7PL5wHPvCAbEYp8bADRZ2B9fHt4QG+hCQ=="
	req, _ := http.NewRequest("Get", "", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	cases := []struct {
		token    string
		expected bool
	}{
		{
			token:    "Ea/+QMPauUxzmhi+ZwusDlFTAl9/17dAMgWw723hibWt5YknmpvZh7PL5wHPvCAbEYp8bADRZ2B9fHt4QG+hCQ==",
			expected: true,
		},
		{
			token:    "IncorrectToken",
			expected: false,
		},
	}

	for _, c := range cases {
		token, err := GetBearerToken(req.Header)
		if err != nil {
			t.Errorf("Failed to get bearer token: %v", err)
			continue
		}

		if c.token != token && c.expected == true {
			t.Errorf("Actual and expected tokens do not match: \n%v \vvs\n %v", c.token, token)
		}

	}
}
