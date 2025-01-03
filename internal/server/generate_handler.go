package server

import (
	"net/http"

	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/templates"
)

func generateHandler(generator *namemyserver.Generator) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		res, err := generator.Generate(r.Context(), namemyserver.GenerateOptions{})
		if err != nil {
			return err
		}

		c := templates.GeneratePartial(templates.GenerateViewModel{
			Name: res.Name,
		})
		return component(w, r, http.StatusOK, c)
	}
}
