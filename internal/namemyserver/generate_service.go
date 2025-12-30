package namemyserver

import (
	"context"
	"fmt"
)

type LengthMode string

const (
	LengthModeExactly LengthMode = "exactly"
	LengthModeUpto    LengthMode = "upto"
)

type Generator struct {
	pairStore PairStore
}

func NewGenerator(pairStore PairStore) *Generator {
	return &Generator{
		pairStore: pairStore,
	}
}

func (g *Generator) Generate(ctx context.Context, opts GenerateOptions) (GenerateResult, error) {
	filters := RandomPairFilters{}
	if opts.LengthEnabled {
		filters.Length = opts.LengthValue
		filters.LengthMode = opts.LengthMode
	}

	p, err := g.pairStore.OneRandom(ctx, filters)
	if err != nil {
		return GenerateResult{}, fmt.Errorf("could not generate a name pair: %w", err)
	}

	return GenerateResult{
		Name: fmt.Sprintf("%s-%s", p.Adjective, p.Noun),
	}, nil
}

type GenerateResult struct {
	Name string
}

type GenerateOptions struct {
	LengthEnabled bool
	LengthMode    LengthMode
	LengthValue   int
}
