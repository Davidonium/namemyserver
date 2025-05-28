package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/templates"
)

func bucketListHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		_, archived := r.URL.Query()["archived"]

		buckets, err := bucketStore.List(ctx, namemyserver.ListOptions{
			ArchivedOnly: archived,
		})
		if err != nil {
			return err
		}

		c := templates.BucketListPage(templates.BucketListPageViewModel{
			Buckets: buckets,
			Archived: archived,
		})
		return component(w, r, http.StatusOK, c)
	}
}

func bucketCreateHandler(logger *slog.Logger, generator *namemyserver.Generator) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		result, err := generator.Generate(ctx, namemyserver.GenerateOptions{})
		if err != nil {
			logger.Error("failed to automatically generate a name for a new bucket", slog.Any("err", err))
		}

		vm := templates.BucketCreatePageViewModel{
			GeneratedName: result.Name,
		}
		c := templates.BucketCreatePage(vm)
		return component(w, r, http.StatusOK, c)
	}
}

func bucketCreateSubmitHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		name := r.FormValue("name")
		description := r.FormValue("description")

		b := namemyserver.Bucket{
			Name:        name,
			Description: description,
		}
		if err := bucketStore.Create(ctx, &b); err != nil {
			return err
		}

		// TODO handle creating buckets with filters
		f := namemyserver.RandomPairFilters{}
		if err := bucketStore.FillBucketValues(ctx, b, f); err != nil {
			return err
		}
		http.Redirect(w, r, fmt.Sprintf("/buckets/%d", b.ID), http.StatusFound)
		return nil
	}
}

func bucketDetailsHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		rawID := r.PathValue("id")
		id, _ := strconv.ParseInt(rawID, 10, 32)
		b, err := bucketStore.OneByID(ctx, int32(id))
		if err != nil {
			return err
		}

		c := templates.BucketDetailsPage(templates.BucketDetailsPageViewModel{Bucket: b})
		return component(w, r, http.StatusOK, c)
	}
}

func bucketArchiveHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		rawID := r.PathValue("id")
		id, _ := strconv.ParseInt(rawID, 10, 32)
		b, err := bucketStore.OneByID(ctx, int32(id))
		if err != nil {
			return err
		}

		b.MarkArchived()

		if err := bucketStore.Save(ctx, &b); err != nil {
			return err
		}

		http.Redirect(w, r, fmt.Sprintf("/buckets/%d", id), http.StatusFound)
		return nil
	}
}


func bucketRecoverHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		rawID := r.PathValue("id")
		id, _ := strconv.ParseInt(rawID, 10, 32)
		b, err := bucketStore.OneByID(ctx, int32(id))
		if err != nil {
			return err
		}

		b.Recover()

		if err := bucketStore.Save(ctx, &b); err != nil {
			return err
		}

		http.Redirect(w, r, fmt.Sprintf("/buckets/%d", id), http.StatusFound)
		return nil
	}
}
