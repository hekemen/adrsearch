package service

import (
	"context"
	"fmt"

	"github.com/ollama/ollama/api"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type VectorGenerator struct {
	client *api.Client
}

func NewVectorGenerator() (*VectorGenerator, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("creating ollama client: %w", err)
	}

	return &VectorGenerator{
		client: client,
	}, nil
}

func (v *VectorGenerator) Generate(ctx context.Context, queryTerm string) ([]float64, error) {
	req := &api.EmbedRequest{
		Model: "qwen3-embedding:0.6b",
		Input: queryTerm,
	}

	resp, err := v.client.Embed(ctx, req)
	if err != nil {
		log.Err(err).Msg("error in model")
		return nil, err
	}

	return lo.Map(resp.Embeddings[0], func(v float32, _ int) float64 {
		return float64(v)
	}), nil
}
