package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/samber/lo"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/types"
)

type OpenAI struct {
	client *openai.Client
	cfg    *config.LLMProviderOpenAI
}

var _ Provider = (*OpenAI)(nil)

func NewOpenAI(cfg *config.LLMProvider) (Provider, error) {
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.OpenAI.APIKey),
		option.WithRequestTimeout(cfg.OpenAI.Timeout),
	}
	if lo.IsNotEmpty(cfg.OpenAI.BaseURL) {
		opts = append(opts, option.WithBaseURL(cfg.OpenAI.BaseURL))
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
		client: &client,
		cfg:    &cfg.OpenAI,
	}

	return o, nil
}

func (o OpenAI) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	body := openai.EmbeddingNewParams{
		Input:          openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: inputs},
		Model:          o.cfg.Embedding.Model,
		Dimensions:     param.NewOpt(int64(o.cfg.Embedding.Dimensions)),
		EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
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

func (o OpenAI) ImageToText(ctx context.Context, mimeType string, image []byte) (string, error) {
	base64Image := base64.StdEncoding.EncodeToString(image)
	body := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(imageToTextPrompt),
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
				openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
					URL: fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image),
				}),
			}),
		},
		Model: o.cfg.ImageToText.Model,
	}
	resp, err := o.client.Chat.Completions.New(ctx, body)
	if err != nil {
		return "", fmt.Errorf("failed to completion: %w", err)
	}
	return resp.Choices[0].Message.Content, nil
}

func (o OpenAI) Close() error {
	return nil
}

func init() {
	RegisterProvider(types.TypeOpenAI, NewOpenAI)
}
