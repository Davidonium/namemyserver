package namemyserver

import "context"

type PairStore interface {
	OneRandom(context.Context, RandomPairFilters) (Pair, error)
	Stats(context.Context) (Stats, error)
}

type RandomPairFilters struct {
	Length int
	LengthMode LengthMode
}
