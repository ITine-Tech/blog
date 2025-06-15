package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strings"
	"testing"
)

func Test_application_routes(t *testing.T) {
	var registeredRoutes = []struct {
		name           string
		route          string
		expectedMethod string
	}{
		{name: "Health route", route: "/healthcheck", expectedMethod: "GET"},
		{name: "Get feed", route: "/feed", expectedMethod: "GET"},
		{name: "Get post by ID", route: "/feed/{postID}", expectedMethod: "GET"},
		{name: "Activates user", route: "/users/activate/{token}", expectedMethod: "PUT"},
		{name: "Authenticate user", route: "/authentication/user", expectedMethod: "POST"},
		{name: "Authentication token", route: "/authentication/token", expectedMethod: "POST"},
		{name: "Activation of user accounts", route: "/users/activate/{token}", expectedMethod: "PUT"},
	}

	var app application
	mux := app.mount()

	chiRoutes := mux.(chi.Routes)

	for _, route := range registeredRoutes {
		t.Run(route.name, func(t *testing.T) {
			if !routeExists(route.route, route.expectedMethod, chiRoutes) {
				t.Errorf("Route %s not found", route.route)
			}
		})
	}
}

func routeExists(testRoute, testMethod string, chiRoutes chi.Routes) bool {
	found := false

	err := chi.Walk(chiRoutes, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if strings.EqualFold(method, testMethod) && strings.EqualFold(route, testRoute) {
			found = true
			return nil
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return found
}
