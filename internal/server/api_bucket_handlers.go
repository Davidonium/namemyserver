package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
		rawID := r.PathValue("id")

		id, err := strconv.Atoi(rawID)
		if err != nil {
			return err
		}

		b, err := bucketStore.OneByID(ctx, int32(id))
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
	type response struct {
		Buckets []bucketListItem `json:"buckets"`
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

		var items []bucketListItem
		for _, b := range buckets {
			items = append(items, bucketToListItem(b))
		}

		return encode(w, http.StatusOK, response{
			Buckets: items,
		})
	}
}

func apiBucketDetailsHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		rawID := r.PathValue("id")

		id, err := strconv.Atoi(rawID)
		if err != nil {
			return err
		}

		b, err := bucketStore.OneByID(ctx, int32(id))
		if err != nil {
			return err
		}

		remaining, err := bucketStore.RemainingValuesTotal(ctx, b)
		if err != nil {
			return err
		}

		return encode(w, http.StatusOK, bucketToDetails(b, remaining))
	}
}

type bucketListItem struct {
	ID          int32      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	ArchivedAt  *time.Time `json:"archivedAt"`
}

func bucketToListItem(b namemyserver.Bucket) bucketListItem {
	return bucketListItem{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
		ArchivedAt:  b.ArchivedAt,
	}
}

type bucketDetails struct {
	ID             int32      `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt"`
	ArchivedAt     *time.Time `json:"archivedAt"`
	RemainingPairs int64      `json:"remainingPairs"`
}

func bucketToDetails(b namemyserver.Bucket, remainingPairs int64) bucketDetails {
	return bucketDetails{
		ID:             b.ID,
		Name:           b.Name,
		Description:    b.Description,
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
		ArchivedAt:     b.ArchivedAt,
		RemainingPairs: remainingPairs,
	}
}
