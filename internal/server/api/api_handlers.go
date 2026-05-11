package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/davidonium/serverplate/internal/serverplate"
)

var ErrArchived = errors.New("the bucket is archived")

type Handlers struct {
	generator   *serverplate.Generator
	bucketStore serverplate.BucketStore
}

func New(generator *serverplate.Generator, bucketStore serverplate.BucketStore) *Handlers {
	return &Handlers{
		generator:   generator,
		bucketStore: bucketStore,
	}
}

// bucketNotFound returns a ProblemDetail for 404 "bucket not found" errors.
// The return value can be type-converted to any *404JSONResponse type.
func bucketNotFound() ProblemDetail {
	return ProblemDetail{
		Status: 404,
		Type:   "not_found",
		Title:  "Bucket not found",
		Detail: new("The requested bucket does not exist"),
	}
}

// bucketArchived returns a ProblemDetail for 409 "bucket archived" conflicts.
// The return value can be type-converted to any *409JSONResponse type.
func bucketArchived() ProblemDetail {
	return ProblemDetail{
		Status: 409,
		Type:   "operation_conflict",
		Title:  "Operation conflict. Bucket is read only.",
		Detail: new("The bucket is archived. Only read operations can be issued against it."),
	}
}

func (s *Handlers) GenerateName(
	ctx context.Context,
	request GenerateNameRequestObject,
) (GenerateNameResponseObject, error) {
	opts := serverplate.GenerateOptions{}

	if request.Body != nil && request.Body.Filters != nil {
		filters := request.Body.Filters

		if filters.LengthEnabled != nil && *filters.LengthEnabled {
			opts.LengthEnabled = true

			if filters.Length == nil {
				return GenerateName400JSONResponse{
					Status: 400,
					Type:   "validation_error",
					Title:  "Validation failed",
					Detail: new("length is required when length_enabled is true"),
				}, nil
			}

			opts.LengthValue = *filters.Length

			if filters.LengthMode != nil {
				opts.LengthMode = serverplate.LengthMode(*filters.LengthMode)
			} else {
				opts.LengthMode = serverplate.LengthModeUpto
			}
		}
	}

	res, err := s.generator.Generate(ctx, opts)
	if err != nil {
		if errors.Is(err, serverplate.ErrNoMatchingPairs) {
			return GenerateName400JSONResponse{
				Status: 400,
				Type:   "no_matches",
				Title:  "No names match the specified filters",
				Detail: new(
					"The length constraints are too restrictive. No adjective-noun combinations match the criteria.",
				),
			}, nil
		}
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

	b := serverplate.Bucket{
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
				b.FilterLengthMode = serverplate.LengthMode(*request.Body.Filters.LengthMode)
			} else {
				b.FilterLengthMode = serverplate.LengthModeUpto
			}
		}
	}

	if err := s.bucketStore.Create(ctx, &b); err != nil {
		return nil, err
	}

	if err := s.bucketStore.FillBucketValues(ctx, b, b.Filters()); err != nil {
		return nil, err
	}

	remaining, err := s.bucketStore.RemainingValuesTotal(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("failed to get remaining pairs count: %w", err)
	}

	response := CreateBucket201JSONResponse{
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
		response.Filters.Length = &b.FilterLengthValue
		lengthMode := BucketDetailsFiltersLengthMode(b.FilterLengthMode)
		response.Filters.LengthMode = &lengthMode
	}

	return response, nil
}

func (s *Handlers) ListBuckets(
	ctx context.Context,
	request ListBucketsRequestObject,
) (ListBucketsResponseObject, error) {
	archived := request.Params.Archived != nil

	buckets, err := s.bucketStore.List(ctx, serverplate.ListOptions{
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
		if errors.Is(err, serverplate.ErrBucketNotFound) {
			return GetBucketDetails404JSONResponse(bucketNotFound()), nil
		}
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
		response.Filters.Length = &b.FilterLengthValue
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
		if errors.Is(err, serverplate.ErrBucketNotFound) {
			return PopBucketName404JSONResponse(bucketNotFound()), nil
		}
		return nil, fmt.Errorf("failed to retrieve bucket by id: %w", err)
	}

	if b.Archived() {
		return PopBucketName409JSONResponse(bucketArchived()), nil
	}

	name, err := s.bucketStore.PopName(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("failed to pop a name from the bucket: %w", err)
	}

	return PopBucketName200JSONResponse{
		Name: name,
	}, nil
}

func (s *Handlers) UpdateBucket(
	ctx context.Context,
	request UpdateBucketRequestObject,
) (UpdateBucketResponseObject, error) {
	if request.Body == nil {
		return nil, fmt.Errorf("request body is required")
	}

	b, err := s.bucketStore.OneByID(ctx, request.Id)
	if err != nil {
		if errors.Is(err, serverplate.ErrBucketNotFound) {
			return UpdateBucket404JSONResponse(bucketNotFound()), nil
		}
		return nil, fmt.Errorf("failed to retrieve bucket by id: %w", err)
	}

	if b.Archived() {
		return UpdateBucket409JSONResponse(bucketArchived()), nil
	}

	if request.Body.Description != nil {
		newDesc := *request.Body.Description

		if len(newDesc) > 2048 {
			return UpdateBucket400JSONResponse{
				Status: 400,
				Type:   "validation_error",
				Title:  "Validation failed",
				Detail: new("Description must not exceed 2048 characters"),
			}, nil
		}

		b.Description = newDesc
	}

	if err := s.bucketStore.Save(ctx, &b); err != nil {
		return nil, fmt.Errorf("failed to save bucket: %w", err)
	}

	remaining, err := s.bucketStore.RemainingValuesTotal(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("failed to get remaining pairs count: %w", err)
	}

	response := UpdateBucket200JSONResponse{
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
		response.Filters.Length = &b.FilterLengthValue
		lengthMode := BucketDetailsFiltersLengthMode(b.FilterLengthMode)
		response.Filters.LengthMode = &lengthMode
	}

	return response, nil
}

func (s *Handlers) ArchiveBucket(
	ctx context.Context,
	request ArchiveBucketRequestObject,
) (ArchiveBucketResponseObject, error) {
	b, err := s.bucketStore.OneByID(ctx, request.Id)
	if err != nil {
		if errors.Is(err, serverplate.ErrBucketNotFound) {
			return ArchiveBucket404JSONResponse(bucketNotFound()), nil
		}
		return nil, fmt.Errorf("failed to retrieve bucket by id: %w", err)
	}

	b.MarkArchived()

	if err := s.bucketStore.Save(ctx, &b); err != nil {
		return nil, fmt.Errorf("failed to save bucket: %w", err)
	}

	remaining, err := s.bucketStore.RemainingValuesTotal(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("failed to get remaining pairs count: %w", err)
	}

	response := ArchiveBucket200JSONResponse{
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
		response.Filters.Length = &b.FilterLengthValue
		lengthMode := BucketDetailsFiltersLengthMode(b.FilterLengthMode)
		response.Filters.LengthMode = &lengthMode
	}

	return response, nil
}

func (s *Handlers) RecoverBucket(
	ctx context.Context,
	request RecoverBucketRequestObject,
) (RecoverBucketResponseObject, error) {
	b, err := s.bucketStore.OneByID(ctx, request.Id)
	if err != nil {
		if errors.Is(err, serverplate.ErrBucketNotFound) {
			return RecoverBucket404JSONResponse(bucketNotFound()), nil
		}
		return nil, fmt.Errorf("failed to retrieve bucket by id: %w", err)
	}

	b.Recover()

	if err := s.bucketStore.Save(ctx, &b); err != nil {
		return nil, fmt.Errorf("failed to save bucket: %w", err)
	}

	remaining, err := s.bucketStore.RemainingValuesTotal(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("failed to get remaining pairs count: %w", err)
	}

	response := RecoverBucket200JSONResponse{
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
		response.Filters.Length = &b.FilterLengthValue
		lengthMode := BucketDetailsFiltersLengthMode(b.FilterLengthMode)
		response.Filters.LengthMode = &lengthMode
	}

	return response, nil
}
