package server

import (
	"net/http"

	"github.com/davidonium/serverplate/internal/templates"
)

func apiDocsHandler() appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return component(w, r, http.StatusOK, templates.DocsPage())
	}
}
