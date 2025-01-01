package namemyserver

import "context"

type PairStore interface {
	GetSinglePair(context.Context) (Pair, error)
	Stats(context.Context) (Stats, error)
}
