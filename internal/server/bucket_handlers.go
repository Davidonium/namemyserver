package server

import (
	"net/http"
	"strconv"

	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/templates"
)

func bucketListHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		buckets, err := bucketStore.All(ctx)
		if err != nil {
			return err
		}

		c := templates.BucketListPage(templates.BucketListPageViewModel{
			Buckets: buckets,
		})
		return component(w, r, http.StatusOK, c)
	}
}

func bucketCreateHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		c := templates.BucketCreatePage()
		return component(w, r, http.StatusOK, c)
	}
}

func bucketCreateSubmitHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		name := r.FormValue("name")

		b := namemyserver.Bucket{
			Name: name,
		}
		if err := bucketStore.Create(ctx, &b); err != nil {
			return err
		}

		// TODO handle creating buckets with filters
		f := namemyserver.RandomPairFilters{}
		if err := bucketStore.FillBucketValues(ctx, b, f); err != nil {
			return err
		}
		http.Redirect(w, r, "/buckets/"+b.Name, http.StatusFound)
		return nil
	}
}

func bucketDetailsHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		rawID := r.PathValue("id")
		id, _ := strconv.Atoi(rawID)
		b, err := bucketStore.OneByID(ctx, int32(id))
		if err != nil {
			return err
		}

		c := templates.BucketDetailsPage(templates.BucketDetailsPageViewModel{Bucket: b})
		return component(w, r, http.StatusOK, c)
	}
}
