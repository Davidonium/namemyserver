package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/davidonium/namemyserver/internal/namemyserver"
)

var ErrArchived = errors.New("the bucket is archived")

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

		if b.Archived() {
			return ErrArchived
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

func apiBucketListHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	type bucketItem struct {
		ID   int32  `json:"id"`
		Name string `json:"name"`
	}
	type response struct {
		Buckets []bucketItem `json:"buckets"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		_, archived := r.URL.Query()["archived"]

		buckets, err := bucketStore.List(ctx, namemyserver.ListOptions{
			ArchivedOnly: archived,
		})
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
