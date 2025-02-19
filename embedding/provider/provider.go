package provider

import (
	"context"
	"fmt"

	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/types"
	"go.uber.org/fx"
)

type Provider interface {
	Embed(ctx context.Context, inputs []string) ([][]float32, error)
}

type Params struct {
	fx.In

	Config *config.Config
}

func New(params Params) (Provider, error) {
	newProvider, ok := availableProviders[params.Config.Embedding.Provider]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", params.Config.Embedding.Provider)
	}

	return newProvider(&params.Config.Embedding)
}

type newProviderFunc func(cfg *config.Embedding) (Provider, error)

var availableProviders = map[types.ProviderType]newProviderFunc{}

func RegisterProvider(name types.ProviderType, provider newProviderFunc) {
	availableProviders[name] = provider
}
