package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/samber/lo"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/types"
	"google.golang.org/api/option"
)

type Google struct {
	client *genai.Client
	cfg    *config.LLMProviderGoogle
}

var _ Provider = (*Google)(nil)

func NewGoogle(cfg *config.LLMProvider) (Provider, error) {
	ctx := context.Background()
	httpClient := &http.Client{Timeout: cfg.Google.Timeout}

	opts := []option.ClientOption{
		option.WithAPIKey(cfg.Google.APIKey),
		option.WithHTTPClient(httpClient),
	}
	if lo.IsNotEmpty(&cfg.Google.BaseURL) {
		opts = append(opts, option.WithEndpoint(cfg.Google.BaseURL))
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
		cfg:    &cfg.Google,
	}

	return o, nil
}

func (g Google) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	em := g.client.EmbeddingModel(g.cfg.Embedding.Model)
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

func (g Google) ImageToText(ctx context.Context, mimeType string, image []byte) (string, error) {
	gm := g.client.GenerativeModel(g.cfg.ImageToText.Model)
	gm.SystemInstruction = genai.NewUserContent(genai.Text(imageToTextPrompt))
	format := strings.TrimPrefix(mimeType, "image/")
	resp, err := gm.GenerateContent(ctx, genai.ImageData(format, image))
	if err != nil {
		return "", fmt.Errorf("failed to generate text from image: %w", err)
	}
	var content string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textPart, ok := part.(genai.Text); ok {
			content += string(textPart)
		}
	}
	return content, nil
}

func (g Google) Close() error {
	return g.client.Close()
}

func init() {
	RegisterProvider(types.TypeGoogle, NewGoogle)
}
