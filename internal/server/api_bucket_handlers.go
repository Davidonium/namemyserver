package server

import (
	"fmt"
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

func apiPopBucketNameHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	type response struct {
		Name string `json:"name"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		bucketName := r.PathValue("name")

		b, err := bucketStore.OneByName(ctx, bucketName)
		if err != nil {
			return fmt.Errorf("failed to retrieve bucket by name: %w", err)
		}

		name, err := bucketStore.PopName(ctx, b)
		if err != nil {
			return fmt.Errorf("failed to pop a name from the bucket: %w", err)
		}

		return encode(w, http.StatusOK, response{
			Name: name,
		})
	}
}
