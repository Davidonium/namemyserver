package namemyserver

import (
	"context"
	"fmt"
)

type Generator struct {
	pairStore PairStore
}

func NewGenerator(pairStore PairStore) *Generator {
	return &Generator{
		pairStore: pairStore,
	}
}

func (g *Generator) Generate(ctx context.Context, _ GenerateOptions) (GenerateResult, error) {
	p, err := g.pairStore.GetSinglePair(ctx)
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

type GenerateOptions struct{}
