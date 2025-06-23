package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	ollamaapi "github.com/ollama/ollama/api"
	"github.com/samber/lo"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/types"
)

type Ollama struct {
	client *ollamaapi.Client
	cfg    *config.LLMProviderOllama
}

var _ Provider = (*Ollama)(nil)

func NewOllama(cfg *config.LLMProvider) (Provider, error) {
	baseURL, err := url.Parse(cfg.Ollama.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	httpClient := &http.Client{Timeout: cfg.Ollama.Timeout}

	o := &Ollama{
		client: ollamaapi.NewClient(baseURL, httpClient),
		cfg:    &cfg.Ollama,
	}

	return o, nil
}

func (o Ollama) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	req := &ollamaapi.EmbedRequest{
		Model:     o.cfg.Embedding.Model,
		Input:     inputs,
		KeepAlive: &ollamaapi.Duration{Duration: o.cfg.KeepAlive},
		Options:   o.cfg.Embedding.ModelParameters,
	}
	resp, err := o.client.Embed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to embed: %w", err)
	}
	return resp.Embeddings, nil
}

func (o Ollama) ImageToText(ctx context.Context, mimeType string, image []byte) (string, error) {
	req := &ollamaapi.ChatRequest{
		Model: o.cfg.ImageToText.Model,
		Messages: []ollamaapi.Message{
			{Role: "system", Content: imageToTextPrompt},
			{Role: "user", Images: []ollamaapi.ImageData{image}},
		},
		Stream:    lo.ToPtr(false),
		KeepAlive: &ollamaapi.Duration{Duration: o.cfg.KeepAlive},
		Options:   o.cfg.ImageToText.ModelParameters,
		Think:     lo.ToPtr(false),
	}
	var resp ollamaapi.ChatResponse
	respFunc := func(r ollamaapi.ChatResponse) error {
		resp = r
		return nil
	}
	if err := o.client.Chat(ctx, req, respFunc); err != nil {
		return "", fmt.Errorf("failed to chat: %w", err)
	}
	return resp.Message.Content, nil
}

func (o Ollama) Close() error {
	return nil
}

func init() {
	RegisterProvider(types.TypeOllama, NewOllama)
}
