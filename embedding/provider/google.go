package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/generative-ai-go/genai"
	"github.com/samber/lo"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/types"
	"google.golang.org/api/option"
)

type Google struct {
	client *genai.Client
	cfg    *config.Embedding
}

var _ Provider = (*Google)(nil)

func NewGoogle(cfg *config.Embedding) (Provider, error) {
	ctx := context.Background()
	httpClient := &http.Client{Timeout: cfg.Timeout}

	opts := []option.ClientOption{
		option.WithAPIKey(cfg.Google.APIKey),
		option.WithHTTPClient(httpClient),
	}
	if lo.IsNotEmpty(cfg.BaseURL) {
		opts = append(opts, option.WithEndpoint(cfg.BaseURL))
	}
	if lo.IsNotEmpty(cfg.Google.QuotaProject) {
		opts = append(opts, option.WithQuotaProject(cfg.Google.QuotaProject))
	}

	client, err := genai.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create google client: %w", err)
	}
	o := &Google{
		client: client,
		cfg:    cfg,
	}

	return o, nil
}

func (g Google) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	em := g.client.EmbeddingModel(g.cfg.Model)
	b := em.NewBatch()
	for _, input := range inputs {
		b = b.AddContent(genai.Text(input))
	}
	resp, err := em.BatchEmbedContents(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("failed to embed: %w", err)
	}
	embeddings := lo.Map(resp.Embeddings,
		func(e *genai.ContentEmbedding, _ int) []float32 { return e.Values },
	)
	return embeddings, nil
}

func (g Google) Close() error {
	return g.client.Close()
}

func init() {
	RegisterProvider(types.TypeGoogle, NewGoogle)
}
