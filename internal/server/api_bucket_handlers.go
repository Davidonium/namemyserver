package server

import (
	"net/http"

	"github.com/davidonium/namemyserver/internal/namemyserver"
)

func apiCreateBucketHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	type request struct {
		Name string `json:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		req, err := decode[request](r)
		if err != nil {
			return err
		}

		ctx := r.Context()

		b := namemyserver.Bucket{
			Name: req.Name,
		}
		if err := bucketStore.Create(ctx, &b); err != nil {
			return err
		}


		if err := bucketStore.FillBucketValues(ctx, b, namemyserver.RandomPairFilters{}); err != nil {
			return err
		}


		w.WriteHeader(http.StatusCreated)
		return nil
	}
}
