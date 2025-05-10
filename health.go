package main

import (
	_ "berta2/docs"
	"net/http"
	"time"
)

// healthcheck godoc
//
//	@Summary		Healthcheck
//	@Description	Healthcheck endpoint
//	@Tags			Ops
//	@Produce		json
//	@Success		200	{object}	string	"ok"
//	@Router			/healthcheck [get]
func (app *application) healthCheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "healthy",
		"time":    time.Now().Format(time.RFC3339),
		"version": version,
		"message": "API is running and ready to accept requests",
	}

	err := writeJSON(w, http.StatusOK, data)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
