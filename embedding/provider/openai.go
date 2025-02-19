package provider

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/samber/lo"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/types"
)

type OpenAI struct {
	client *openai.Client
	cfg    *config.Embedding
}

var _ Provider = (*OpenAI)(nil)

func NewOpenAI(cfg *config.Embedding) (Provider, error) {
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.OpenAI.APIKey),
		option.WithRequestTimeout(cfg.Timeout),
	}
	if lo.IsNotEmpty(cfg.BaseURL) {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	} else {
		opts = append(opts, option.WithEnvironmentProduction())
	}
	if lo.IsNotEmpty(cfg.OpenAI.Organization) {
		opts = append(opts, option.WithOrganization(cfg.OpenAI.Organization))
	}
	if lo.IsNotEmpty(cfg.OpenAI.Project) {
		opts = append(opts, option.WithProject(cfg.OpenAI.Project))
	}

	client := openai.NewClient(opts...)
	o := &OpenAI{
		client: client,
		cfg:    cfg,
	}

	return o, nil
}

func (o OpenAI) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	body := openai.EmbeddingNewParams{
		Input:          openai.F(openai.EmbeddingNewParamsInputUnion(openai.EmbeddingNewParamsInputArrayOfStrings(inputs))),
		Model:          openai.F(o.cfg.Model),
		Dimensions:     openai.F(int64(o.cfg.Dimensions)),
		EncodingFormat: openai.F(openai.EmbeddingNewParamsEncodingFormatFloat),
	}
	resp, err := o.client.Embeddings.New(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("failed to embed: %w", err)
	}
	embeddings := make([][]float32, len(resp.Data))
	for _, e := range resp.Data {
		embeddings[e.Index] = lo.Map(e.Embedding,
			func(v float64, _ int) float32 {
				return float32(v)
			},
		)
	}
	return embeddings, nil
}

func init() {
	RegisterProvider(types.TypeOpenAI, NewOpenAI)
}
