package internal

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateImagePullSecretsValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    ImagePullSecrets
		expected string
	}{
		{
			name: "All fields provided",
			input: ImagePullSecrets{
				Registry: "my-registry.com",
				Username: "user",
				Password: "password123",
				Email:    "user@example.com",
			},
			expected: fmt.Sprintf(imagePullSecretsTemplate, "my-registry.com", "user", "password123", "email: user@example.com"),
		},
		{
			name: "Default registry used when registry is empty",
			input: ImagePullSecrets{
				Username: "user",
				Password: "password123",
				Email:    "user@example.com",
			},
			expected: fmt.Sprintf(imagePullSecretsTemplate, "https://index.docker.io/v1/", "user", "password123", "email: user@example.com"),
		},
		{
			name: "Email is optional",
			input: ImagePullSecrets{
				Registry: "my-registry.com",
				Username: "user",
				Password: "password123",
			},
			expected: fmt.Sprintf(imagePullSecretsTemplate, "my-registry.com", "user", "password123", ""),
		},
		{
			name: "Returns empty string if username is missing",
			input: ImagePullSecrets{
				Password: "password123",
			},
			expected: "",
		},
		{
			name: "Returns empty string if password is missing",
			input: ImagePullSecrets{
				Username: "user",
			},
			expected: "",
		},
		{
			name:     "Returns empty string for empty input struct",
			input:    ImagePullSecrets{},
			expected: "",
		},
	}

	for _, tc := range tests {
		// Create a new variable for each iteration to avoid issues with parallel tests.
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := generateImagePullSecretsValue(tc.input)

			// Trim whitespace to make the comparison robust.
			gotTrimmed := strings.TrimSpace(got)
			expectedTrimmed := strings.TrimSpace(tc.expected)

			if gotTrimmed != expectedTrimmed {
				t.Errorf("generateImagePullSecretsValue() output mismatch:\nwant: %q\ngot:  %q", expectedTrimmed, gotTrimmed)
			}
		})
	}
}
