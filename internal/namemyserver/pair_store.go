package namemyserver

import (
	"context"
	"errors"
)

var ErrNoMatchingPairs = errors.New("no pairs match the specified filters")

type PairStore interface {
	OneRandom(context.Context, RandomPairFilters) (Pair, error)
	Stats(context.Context, RandomPairFilters) (Stats, error)
}

type RandomPairFilters struct {
	Length     int
	LengthMode LengthMode
}
