package main

import (
	"berta2/internal/auth"
	"berta2/internal/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestApplication creates a new instance of the application for testing purposes.
// It initializes a mock store and returns a pointer to an application struct with the mock store.
//
// Parameters:
// - t: A pointer to a testing.T instance, used for logging errors and marking tests as helper functions.
//
// Return:
// - A pointer to an application instance with a mock store.
func newTestApplication(t *testing.T) *application {
	t.Helper()

	mockStore := store.NewMockStore()
	testAuth := &auth.TestAuthenticator{} // Mock authentication for testing purposes

	return &application{
		store:         mockStore,
		authenticator: testAuth,
	}
}

// executeRequest sends an HTTP request to the provided handler and returns a ResponseRecorder
// to record the response. This function is useful for testing HTTP handlers.
//
// Parameters:
// - req: A pointer to an http.Request instance representing the HTTP request to be sent.
// - mux: An http.Handler instance that will handle the request.
//
// Return:
// - A pointer to an httptest.ResponseRecorder instance, which can be used to inspect the response.
func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

// checkResponseCode compares the expected HTTP response code with the actual one.
// If the expected and actual codes do not match, it logs an error message using the testing.T interface.
//
// Parameters:
// - t: A pointer to a testing.T instance, used for logging errors.
// - expected: The expected HTTP response code.
// - actual: The actual HTTP response code received.
//
// Return:
// This function does not return any value.
func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
