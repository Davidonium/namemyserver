package server

import (
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/templates"
)

func generateHandler(generator *namemyserver.Generator) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		lengthEnabled := r.FormValue("length_enabled")
		lengthMode := r.FormValue("length_mode")
		lengthValue, _ := strconv.Atoi(r.FormValue("length_value"))
		componentType := r.URL.Query().Get("component")

		res, err := generator.Generate(r.Context(), namemyserver.GenerateOptions{
			LengthEnabled: lengthEnabled == "on",
			LengthMode:    namemyserver.LengthMode(lengthMode),
			LengthValue:   lengthValue,
		})
		// Handle errors based on component type
		if err != nil {
			if componentType == "bucket-input" {
				// Silent failure: return empty input
				return component(w, r, http.StatusOK, templates.BucketNameInput(""))
			}
			// Default: return error for other components
			return err
		}

		// Render different components based on parameter
		var c templ.Component
		switch componentType {
		case "bucket-input":
			c = templates.BucketNameInput(res.Name)
		default:
			// Default behavior: render generate partial (home page)
			c = templates.GeneratePartial(templates.GenerateViewModel{
				Name: res.Name,
			})
		}

		return component(w, r, http.StatusOK, c)
	}
}
