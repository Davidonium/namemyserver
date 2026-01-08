package server

import (
	"net/http"

	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/templates"
)

func statsHandler(pairStore namemyserver.PairStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		stats, err := pairStore.Stats(ctx, namemyserver.RandomPairFilters{})
		if err != nil {
			return err
		}

		vm := templates.StatsViewModel{
			DatabaseSizeBytes: stats.DatabaseSizeBytes,
			PairCount:         stats.PairCount,
			AdjectiveCount:    stats.AdjectiveCount,
			NounCount:         stats.NounCount,
		}
		c := templates.StatsPage(vm)
		return component(w, r, http.StatusOK, c)
	}
}
