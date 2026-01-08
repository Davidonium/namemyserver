package server

import (
	"net/http"
	"strconv"

	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/templates"
)

func homeHandler(pairStore namemyserver.PairStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		stats, err := pairStore.Stats(ctx, namemyserver.RandomPairFilters{})
		if err != nil {
			return err
		}

		c := templates.HomePage(templates.HomeViewModel{
			PossiblePairCount: stats.PairCount,
		})
		return component(w, r, http.StatusOK, c)
	}
}

func configStatsHandler(pairStore namemyserver.PairStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		lengthEnabled := r.FormValue("length_enabled") == "on"
		lengthMode := r.FormValue("length_mode")
		lengthValue, _ := strconv.Atoi(r.FormValue("length_value"))

		filters := namemyserver.RandomPairFilters{}
		if lengthEnabled {
			filters.Length = lengthValue
			filters.LengthMode = namemyserver.LengthMode(lengthMode)
		}

		stats, err := pairStore.Stats(ctx, filters)
		if err != nil {
			return err
		}

		c := templates.ConfigurationStatsPartial(templates.ConfigurationStatsPartialViewModel{
			PossiblePairCount: stats.PairCount,
		})
		return component(w, r, http.StatusOK, c)
	}
}
