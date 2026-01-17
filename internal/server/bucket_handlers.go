package server

import (
	"fmt"
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
			Buckets:  buckets,
			Archived: archived,
		})
		return component(w, r, http.StatusOK, c)
	}
}

func bucketCreateHandler() appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		vm := templates.BucketCreatePageViewModel{}
		c := templates.BucketCreatePage(vm)
		return component(w, r, http.StatusOK, c)
	}
}

func bucketCreateSubmitHandler(bucketStore namemyserver.BucketStore) appHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		name := r.FormValue("name")
		description := r.FormValue("description")

		lengthEnabled := r.FormValue("filter_length_enabled") == "on"
		lengthMode := namemyserver.LengthModeUpto
		if r.FormValue("filter_length_mode") != "" {
			lengthMode = namemyserver.LengthMode(r.FormValue("filter_length_mode"))
		}
		lengthValue := 14 // default
		if val := r.FormValue("filter_length_value"); val != "" {
			if parsed, err := strconv.Atoi(val); err == nil {
				lengthValue = parsed
			}
		}

		b := namemyserver.Bucket{
			Name:                name,
			Description:         description,
			FilterLengthEnabled: lengthEnabled,
			FilterLengthMode:    lengthMode,
			FilterLengthValue:   lengthValue,
		}
		if err := bucketStore.Create(ctx, &b); err != nil {
			return err
		}

		if err := bucketStore.FillBucketValues(ctx, b, b.Filters()); err != nil {
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

		count, err := bucketStore.RemainingValuesTotal(ctx, b)
		if err != nil {
			return err
		}

		c := templates.BucketDetailsPage(
			templates.BucketDetailsPageViewModel{Bucket: b, RemainingPairs: count},
		)
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
