package main

import (
	"net/http"
	"testing"
)

// TestGetUserByIDHandler tests the functionality of the GetUserByIDHandler endpoint.
// It covers two scenarios: unauthenticated requests and authenticated requests.
func TestGetUserByIDHandler(t *testing.T) {

	//Setup test application
	app := newTestApplication(t)
	mux := app.mount()

	testToken, err := app.authenticator.GenerateToken(nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("should not allow unauthenticated requests", func(t *testing.T) {
		//Set up: mock user data
		req, err := http.NewRequest(http.MethodGet, "/users/ad35e351-b639-4c06-b75g-cf19a1965e5c", nil)
		if err != nil {
			t.Fatal(err)
		}

		//Hit the endpoint
		rr := executeRequest(req, mux)

		//Expectations
		checkResponseCode(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allow authenticated requests", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/users/7831ef38-724e-4543-b3bd-51e980f88541", nil)
		if err != nil {
			t.Fatal(err)
		}

		//We need a token for testing
		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)
	})

}
