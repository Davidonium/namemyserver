package server

import (
	"net/http"

	"github.com/davidonium/namemyserver/internal/templates"
)

func generateHandler() appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// TODO all logic
		c := templates.GeneratePartial(templates.GenerateViewModel{Name: "happy-dog"})
		return component(w, r, http.StatusOK, c)
	}
}
