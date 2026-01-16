package namemyserver

import "errors"

var (
	// ErrBucketNotFound is returned when a bucket cannot be found
	ErrBucketNotFound = errors.New("bucket not found")

	// ErrNoMatchingPairs is returned when no pairs match the specified filters
	ErrNoMatchingPairs = errors.New("no pairs match the specified filters")
)
