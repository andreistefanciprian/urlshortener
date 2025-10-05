package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateExpiration_ValidExpiration(t *testing.T) {

	// Test case: Valid expiration value
	req := CreateShortURLRequest{
		LongUrl:   "https://example.com",
		ExpiresIn: 30,
	}

	err := req.validateExpiration()

	// Use testify assert for cleaner assertions
	assert.NoError(t, err, "Expected no error for valid expiration")
	assert.Equal(t, 30, req.ExpiresIn, "ExpiresIn should remain unchanged")
}

func TestValidateExpiration_ZeroExpiration(t *testing.T) {

	// Test case: Valid expiration value
	req := CreateShortURLRequest{
		LongUrl:   "https://example.com",
		ExpiresIn: 0,
	}

	err := req.validateExpiration()

	// Use testify assert for cleaner assertions
	assert.NoError(t, err, "Expected no error for zero expiration")
	assert.Equal(t, 7, req.ExpiresIn, "ExpiresIn should default to 7 days")
}

func TestValidateExpiration_NegativeExpiration(t *testing.T) {

	// Test case: Valid expiration value
	req := CreateShortURLRequest{
		LongUrl:   "https://example.com",
		ExpiresIn: -10,
	}

	err := req.validateExpiration()

	// Use testify assert for cleaner assertions
	assert.ErrorContains(t, err, "expiration time cannot be negative")
}

func TestValidateURL_ValidURL(t *testing.T) {
	req := CreateShortURLRequest{
		LongUrl: "https://example.com",
	}

	err := req.validateURL()

	assert.NoError(t, err, "Expected no error for valid URL")
}
func TestValidateURL_EmptyURL(t *testing.T) {

	req := CreateShortURLRequest{
		LongUrl: "",
	}

	err := req.validateURL()

	assert.ErrorContains(t, err, "empty url")
}

func TestValidateURL_NoHostURL(t *testing.T) {
	testCases := []struct {
		name    string
		longUrl string
		errMsg  string
	}{
		{
			name:    "No host with single slash",
			longUrl: "https:/",
			errMsg:  "URL must have a valid host",
		},
		{
			name:    "No host with double slash",
			longUrl: "https://",
			errMsg:  "URL must have a valid host",
		},
		{
			name:    "No host with triple slash",
			longUrl: "https:///",
			errMsg:  "URL must have a valid host",
		},
		{
			name:    "No host with domain after triple slash",
			longUrl: "https:///example.com",
			errMsg:  "URL must have a valid host",
		},
		{
			name:    "No host with domain after single slash",
			longUrl: "https:/example.com",
			errMsg:  "URL must have a valid host",
		},
		{
			name:    "No host with domain after scheme only",
			longUrl: "https:example.com",
			errMsg:  "URL must have a valid host",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateShortURLRequest{
				LongUrl: tc.longUrl,
			}

			err := req.validateURL()

			assert.ErrorContains(t, err, tc.errMsg)
		})
	}
}

func TestValidateURL_InvalidURL(t *testing.T) {
	testCases := []struct {
		name    string
		longUrl string
		errMsg  string
	}{
		{
			name:    "No HTTP/HTTPS scheme: example.com",
			longUrl: "example.com",
			errMsg:  "invalid URL format",
		},
		{
			name:    "Invalid URL: invalidurlcom",
			longUrl: "invalidurlcom",
			errMsg:  "invalid URL format",
		},
		{
			name:    "Invalid URL: invalidurlcom//",
			longUrl: "invalidurlcom//",
			errMsg:  "invalid URL format",
		},
		{
			name:    "Invalid URL: whitespace in host https://example .com",
			longUrl: "https://example .com",
			errMsg:  "invalid URL format",
		},
		{
			name:    "Invalid URL: whitespace after host https://example.com ",
			longUrl: "https://example.com ",
			errMsg:  "invalid URL format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateShortURLRequest{
				LongUrl: tc.longUrl,
			}

			err := req.validateURL()

			assert.ErrorContains(t, err, tc.errMsg)
		})
	}
}

func TestValidateURL_NoHttpScheme(t *testing.T) {
	testCases := []struct {
		name    string
		longUrl string
		errMsg  string
	}{
		{
			name:    "FTP scheme: ftp://example.com",
			longUrl: "ftp://example.com",
			errMsg:  "URL must use http or https scheme",
		},
		{
			name:    "JavaScript scheme: javascript:alert('xss')",
			longUrl: "javascript:alert('xss')",
			errMsg:  "URL must use http or https scheme",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateShortURLRequest{
				LongUrl: tc.longUrl,
			}

			err := req.validateURL()

			assert.ErrorContains(t, err, tc.errMsg)
		})
	}
}

func TestValidateURL_LowercaseSchemaAndHost(t *testing.T) {
	testCases := []struct {
		name            string
		longUrl         string
		expectedLongUrl string
	}{
		{
			name:            "URL Contains Schema and Host Uppercase chars: HTTPS://Example.COM/path",
			longUrl:         "HTTPS://Example.COM/path",
			expectedLongUrl: "https://example.com/path",
		},
		{
			name:            "URL Contains Host Uppercase chars: https://SUBDOMAIN.EXAMPLE.COM:8080/api",
			longUrl:         "https://SUBDOMAIN.EXAMPLE.COM:8080/api",
			expectedLongUrl: "https://subdomain.example.com:8080/api",
		},
		{
			name:            "URL Contains Host Uppercase chars. Keep case sensitive path: https://SUBDOMAIN.EXAMPLE.COM:8080/Api",
			longUrl:         "https://SUBDOMAIN.EXAMPLE.COM:8080/Api",
			expectedLongUrl: "https://subdomain.example.com:8080/Api",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateShortURLRequest{
				LongUrl: tc.longUrl,
			}

			err := req.validateURL()

			assert.NoError(t, err, "Expected no error for URL: "+tc.longUrl)
			assert.Equal(t, tc.expectedLongUrl, req.LongUrl, "LongUrl should have schema and host lowercased")
		})
	}
}

func TestValidateURL_RemoveDefaultPorts(t *testing.T) {
	testCases := []struct {
		name            string
		longUrl         string
		expectedLongUrl string
	}{
		{
			name:            "Remove default HTTPS port: https://example.com:443/path",
			longUrl:         "https://example.com:443/path",
			expectedLongUrl: "https://example.com/path",
		},
		{
			name:            "Remove default HTTP port: http://example.com:80/path",
			longUrl:         "http://example.com:80/path",
			expectedLongUrl: "http://example.com/path",
		},
		{
			name:            "Keeps non default ports: http://example.com:9090/path",
			longUrl:         "http://example.com:9090/path",
			expectedLongUrl: "http://example.com:9090/path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateShortURLRequest{
				LongUrl: tc.longUrl,
			}

			err := req.validateURL()

			assert.NoError(t, err, "Expected no error for URL: "+tc.longUrl)
			assert.Equal(t, tc.expectedLongUrl, req.LongUrl, "LongUrl should have default ports removed")
		})
	}
}
