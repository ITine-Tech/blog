package main

import (
	"database/sql"
	"testing"
)

func Test_application_handlers(t *testing.T) {
	/*	var args = []struct {
			name               string
			method             string
			url                string
			expectedStatusCode int
		}{
			{name: "Comments handler", method: "POST", url: "/comments/{postID}", expectedStatusCode: http.StatusCreated},
		}

		db, err := setupDatabase()
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()

		var app = &application{}
		routes := app.mount()

		testServer := httptest.NewServer(routes)
		defer testServer.Close()

		for _, tt := range args {
			t.Run(tt.name, func(t *testing.T) {
				client := &http.Client{}
				req, err := http.NewRequest(tt.method, testServer.URL+tt.url, nil)
				if err != nil {
					t.Fatal(err)
				}

				resp, err := client.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()

				if resp.StatusCode != tt.expectedStatusCode {
					t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, resp.StatusCode)
				}

			})
		}*/
}

func setupDatabase() (*sql.DB, error) {
	return nil, nil
}
