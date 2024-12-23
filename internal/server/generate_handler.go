package server

import (
	"math/rand/v2"
	"net/http"

	"github.com/davidonium/namemyserver/internal/templates"
)

func generateHandler() appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// TODO all logic
		names := []string{
			"happy-dog",
			"happy-cat",
			"big-bird",
			"small-cow",
		}

		name := names[rand.UintN(uint(len(names)))]

		c := templates.GeneratePartial(templates.GenerateViewModel{Name: name})
		return component(w, r, http.StatusOK, c)
	}
}
