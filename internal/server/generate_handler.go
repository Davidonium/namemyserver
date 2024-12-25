package server

import (
	"fmt"
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

		c := templates.GeneratePartial(templates.GenerateViewModel{Name: fmt.Sprintf("%s-%s", p.Adjective, p.Noun)})
		return component(w, r, http.StatusOK, c)
	}
}
