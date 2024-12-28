package server

import (
	"net/http"

	"github.com/davidonium/namemyserver/internal/namemyserver"
)

func apiGenerateHandler(generator *namemyserver.Generator) appHandlerFunc {
	type response struct {
		Name string `json:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		res, err := generator.Generate(r.Context(), namemyserver.GenerateOptions{})
		if err != nil {
			return err
		}

		return encode(w, http.StatusOK, response{
			Name: res.Name,
		})
	}
}
