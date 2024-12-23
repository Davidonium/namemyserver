package server

import (
	"net/http"

	"github.com/davidonium/namemyserver/internal/templates"
)

func homeHandler() appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		c := templates.HomePage(templates.HomeViewModel{})
		return component(w, r, http.StatusOK, c)
	}
}
