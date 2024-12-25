package server

import (
	"net/http"

	"github.com/davidonium/namemyserver/internal/store/sqlitestore"
	"github.com/davidonium/namemyserver/internal/templates"
)

func generateHandler(pairStore *sqlitestore.PairStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		p, err := pairStore.FindSinglePair(r.Context())
		if err != nil {
			return err
		}

		c := templates.GeneratePartial(templates.GenerateViewModel{Name: p})
		return component(w, r, http.StatusOK, c)
	}
}
