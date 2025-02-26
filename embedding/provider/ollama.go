package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	ollamaapi "github.com/ollama/ollama/api"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/types"
)

type Ollama struct {
	client *ollamaapi.Client
	cfg    *config.Embedding
}

var _ Provider = (*Ollama)(nil)

func NewOllama(cfg *config.Embedding) (Provider, error) {
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	httpClient := &http.Client{Timeout: cfg.Timeout}

	o := &Ollama{
		client: ollamaapi.NewClient(baseURL, httpClient),
		cfg:    cfg,
	}

	return o, nil
}

func (o Ollama) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	req := &ollamaapi.EmbedRequest{
		Model:     o.cfg.Model,
		Input:     inputs,
		KeepAlive: &ollamaapi.Duration{Duration: o.cfg.Ollama.KeepAlive},
		Options:   o.cfg.Ollama.ModelParameters,
	}
	resp, err := o.client.Embed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to embed: %w", err)
	}
	return resp.Embeddings, nil
}

func (o Ollama) Close() error {
	return nil
}

func init() {
	RegisterProvider(types.TypeOllama, NewOllama)
}
