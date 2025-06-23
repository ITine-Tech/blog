package main

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Test_basicAuthMiddleware(t *testing.T) {
	var app application
	mux := app.mount()

	tests := []struct {
		name            string
		username        string
		password        string
		expectedStatus  int
		expectedError   error
		requestMethod   string
		requestEndpoint string
		authHeader      string
	}{
		{
			name:            "valid Authorization header, correct user data",
			username:        app.config.auth.basic.username,
			password:        app.config.auth.basic.pass,
			expectedStatus:  http.StatusOK,
			expectedError:   nil,
			requestMethod:   http.MethodGet,
			requestEndpoint: "/healthcheck",
			authHeader:      "Basic " + base64.StdEncoding.EncodeToString([]byte(app.config.auth.basic.username+":"+app.config.auth.basic.pass)),
		},
		{
			name:            "missing Authorization header",
			username:        "",
			password:        "",
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   errors.New("missing Authorization header"),
			requestMethod:   http.MethodGet,
			requestEndpoint: "/healthcheck",
		},
		{
			name:            "invalid Authorization header format",
			username:        "",
			password:        "",
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   errors.New("invalid Authorization header"),
			requestMethod:   http.MethodGet,
			requestEndpoint: "/healthcheck",
			authHeader:      "Bearer invalid_token",
		},
		{
			name:            "incorrect username",
			username:        "incorrect_username",
			password:        app.config.auth.basic.pass,
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   errors.New("invalid Authorization header"),
			requestMethod:   http.MethodGet,
			requestEndpoint: "/healthcheck",
		},
		{
			name:            "incorrect password",
			username:        app.config.auth.basic.username,
			password:        "invalid_password",
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   errors.New("invalid Authorization header"),
			requestMethod:   http.MethodGet,
			requestEndpoint: "/healthcheck",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(mux)
			defer testServer.Close()

			authHeader := ""
			if tt.username != "" && tt.password != "" {
				authHeader = "Basic " + base64.StdEncoding.EncodeToString([]byte(tt.username+":"+tt.password))
			} else if tt.authHeader != "" {
				authHeader = tt.authHeader
			}

			req, err := http.NewRequest(tt.requestMethod, testServer.URL+tt.requestEndpoint, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", authHeader)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
			if tt.expectedError != nil && resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected error: %v, got none", tt.expectedError)
			}
		})
	}
}

func Test_MY_authTokenMiddleware(t *testing.T) {
	app := newTestApplication(t)
	mux := app.mount()
	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		token           string
		expectedStatus  int
		expectedError   error
		requestMethod   string
		requestEndpoint string
		authHeader      string
	}{
		{
			name:            "Correct authentication",
			token:           testToken,
			expectedStatus:  http.StatusOK,
			expectedError:   nil,
			requestMethod:   "GET",
			requestEndpoint: "/users",
			authHeader:      "Bearer " + testToken,
		},
		{
			name:            "Missing token",
			token:           "",
			expectedStatus:  http.StatusUnauthorized,
			expectedError:   errors.New("missing Authorization header"),
			requestMethod:   "GET",
			requestEndpoint: "/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(mux)
			defer testServer.Close()

			authHeader := ""
			if tt.token != "" {
				authHeader = "Bearer " + tt.token
			} else if tt.authHeader != "" {
				authHeader = tt.authHeader
			}

			req, err := http.NewRequest(tt.requestMethod, testServer.URL+tt.requestEndpoint, nil)
			if err != nil {
				t.Fatal(err)
			}
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedError != nil && resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected error: %v, got none", tt.expectedError)
			}
		})
	}
}

func Test_AuthTokenMiddleware_ExpiredToken(t *testing.T) {
	app := newTestApplication(t)
	mux := app.mount()

	expiredToken, err := app.authenticator.GenerateToken(jwt.MapClaims{
		"exp": time.Now().Add(-360 * time.Hour).Unix(),
	})

	if err != nil {
		t.Fatal(err)
	}

	testServer := httptest.NewServer(mux)
	defer testServer.Close()

	req, err := http.NewRequest(http.MethodGet, testServer.URL+"/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}

}
