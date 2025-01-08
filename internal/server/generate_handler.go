package server

import (
	"net/http"
	"strconv"

	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/templates"
)

func generateHandler(generator *namemyserver.Generator) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		lengthEnabled := r.FormValue("lengthEnabled")
		lengthMode := r.FormValue("lengthMode")
		lengthValue, _ := strconv.Atoi(r.FormValue("lengthValue"))

		res, err := generator.Generate(r.Context(), namemyserver.GenerateOptions{
			LengthEnabled: lengthEnabled == "on",
			LengthMode:    namemyserver.LengthMode(lengthMode),
			LengthValue:   lengthValue,
		})
		if err != nil {
			return err
		}

		c := templates.GeneratePartial(templates.GenerateViewModel{
			Name: res.Name,
		})
		return component(w, r, http.StatusOK, c)
	}
}
