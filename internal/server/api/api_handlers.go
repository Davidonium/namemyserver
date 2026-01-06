package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/davidonium/namemyserver/internal/namemyserver"
	"github.com/davidonium/namemyserver/internal/ptr"
)

var ErrArchived = errors.New("the bucket is archived")

type Handlers struct {
	generator   *namemyserver.Generator
	bucketStore namemyserver.BucketStore
}

func New(generator *namemyserver.Generator, bucketStore namemyserver.BucketStore) *Handlers {
	return &Handlers{
		generator:   generator,
		bucketStore: bucketStore,
	}
}

func (s *Handlers) GenerateName(
	ctx context.Context,
	request GenerateNameRequestObject,
) (GenerateNameResponseObject, error) {
	res, err := s.generator.Generate(ctx, namemyserver.GenerateOptions{})
	if err != nil {
		return nil, err
	}

	return GenerateName200JSONResponse{
		Name: res.Name,
	}, nil
}

func (s *Handlers) CreateBucket(
	ctx context.Context,
	request CreateBucketRequestObject,
) (CreateBucketResponseObject, error) {
	if request.Body == nil {
		return nil, fmt.Errorf("request body is required")
	}

	b := namemyserver.Bucket{
		Name: request.Body.Name,
	}

	if request.Body.Description != nil {
		b.Description = *request.Body.Description
	}

	if request.Body.Filters != nil {
		if request.Body.Filters.LengthEnabled != nil {
			b.FilterLengthEnabled = *request.Body.Filters.LengthEnabled
		}

		if b.FilterLengthEnabled {
			if request.Body.Filters.Length != nil {
				b.FilterLengthValue = *request.Body.Filters.Length
			}
			if request.Body.Filters.LengthMode != nil {
				b.FilterLengthMode = namemyserver.LengthMode(*request.Body.Filters.LengthMode)
			} else {
				b.FilterLengthMode = namemyserver.LengthModeUpto
			}
		}
	}

	if err := s.bucketStore.Create(ctx, &b); err != nil {
		return nil, err
	}

	if err := s.bucketStore.FillBucketValues(ctx, b, b.Filters()); err != nil {
		return nil, err
	}

	return CreateBucket201Response{}, nil
}

func (s *Handlers) ListBuckets(
	ctx context.Context,
	request ListBucketsRequestObject,
) (ListBucketsResponseObject, error) {
	archived := request.Params.Archived != nil

	buckets, err := s.bucketStore.List(ctx, namemyserver.ListOptions{
		ArchivedOnly: archived,
	})
	if err != nil {
		return nil, err
	}

	var items []BucketListItem
	for _, b := range buckets {
		items = append(items, BucketListItem{
			Id:          b.ID,
			Name:        b.Name,
			Description: b.Description,
			CreatedAt:   b.CreatedAt,
			UpdatedAt:   b.UpdatedAt,
			ArchivedAt:  b.ArchivedAt,
		})
	}

	return ListBuckets200JSONResponse{
		Buckets: items,
	}, nil
}

func (s *Handlers) GetBucketDetails(
	ctx context.Context,
	request GetBucketDetailsRequestObject,
) (GetBucketDetailsResponseObject, error) {
	b, err := s.bucketStore.OneByID(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	remaining, err := s.bucketStore.RemainingValuesTotal(ctx, b)
	if err != nil {
		return nil, err
	}

	response := GetBucketDetails200JSONResponse{
		Id:             b.ID,
		Name:           b.Name,
		Description:    b.Description,
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
		ArchivedAt:     b.ArchivedAt,
		RemainingPairs: remaining,
	}

	response.Filters.LengthEnabled = b.FilterLengthEnabled
	if b.FilterLengthEnabled {
		response.Filters.Length = ptr.To(b.FilterLengthValue)
		lengthMode := BucketDetailsFiltersLengthMode(b.FilterLengthMode)
		response.Filters.LengthMode = &lengthMode
	}

	return response, nil
}

func (s *Handlers) PopBucketName(
	ctx context.Context,
	request PopBucketNameRequestObject,
) (PopBucketNameResponseObject, error) {
	b, err := s.bucketStore.OneByID(ctx, request.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve bucket by id: %w", err)
	}

	if b.Archived() {
		return PopBucketName409JSONResponse{
			Status: 409,
			Type:   "operation_conflict",
			Title:  "Operation conflict. Bucket is read only.",
			Detail: ptr.To(
				"The bucket is archived. Only read operations can be issued against it.",
			),
		}, nil
	}

	name, err := s.bucketStore.PopName(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("failed to pop a name from the bucket: %w", err)
	}

	return PopBucketName200JSONResponse{
		Name: name,
	}, nil
}
