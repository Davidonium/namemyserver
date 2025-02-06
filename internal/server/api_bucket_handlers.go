package server

import (
	"fmt"
	"net/http"

	"github.com/davidonium/namemyserver/internal/namemyserver"
)

func apiCreateBucketHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	type filters struct {
		Length     int    `json:"length"`
		LengthMode string `json:"lengthMode"`
	}

	type request struct {
		Name    string  `json:"name"`
		Filters filters `json:"filters"`
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

		f := namemyserver.RandomPairFilters{
			Length:     req.Filters.Length,
			LengthMode: namemyserver.LengthMode(req.Filters.LengthMode),
		}
		if err := bucketStore.FillBucketValues(ctx, b, f); err != nil {
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

func apiListBucketsHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	type bucketItem struct {
		ID   int32  `json:"id"`
		Name string `json:"name"`
	}
	type response struct {
		Buckets []bucketItem `json:"buckets"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		buckets, err := bucketStore.All(ctx)
		if err != nil {
			return err
		}

		var items []bucketItem
		for _, b := range buckets {
			items = append(items, bucketItem{
				ID:   b.ID,
				Name: b.Name,
			})
		}

		return encode(w, http.StatusOK, response{
			Buckets: items,
		})
	}
}
