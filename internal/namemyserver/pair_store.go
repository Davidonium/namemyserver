package namemyserver

import "context"

type PairStore interface {
	GetSinglePair(ctx context.Context) (Pair, error)
}
