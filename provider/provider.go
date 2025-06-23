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
	ImageToText(ctx context.Context, mimeType string, image []byte) (string, error)
	Close() error
}

type EmbeddingProvider Provider
type ImageToTextProvider Provider

type Params struct {
	fx.In

	LifeCycle fx.Lifecycle
	Config    *config.Config
}

func NewEmbeddingProvider(params Params) (EmbeddingProvider, error) {
	return new("embedding", params)
}

func NewImageToTextProvider(params Params) (ImageToTextProvider, error) {
	return new("imageToText", params)
}

func new(class string, params Params) (Provider, error) {
	var providerType types.ProviderType
	switch class {
	case "embedding":
		providerType = params.Config.Embedding.Provider
	case "imageToText":
		providerType = params.Config.ImageToText.Provider
	default:
		return nil, fmt.Errorf("unknown provider class: %s", class)
	}
	newProvider, ok := availableProviders[providerType]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerType)
	}
	p, err := newProvider(&params.Config.LLMProvider)
	if err != nil {
		return nil, err
	}

	if params.LifeCycle != nil {
		params.LifeCycle.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return p.Close()
			},
		})
	}

	return p, nil
}

type newProviderFunc func(cfg *config.LLMProvider) (Provider, error)

var availableProviders = map[types.ProviderType]newProviderFunc{}

func RegisterProvider(name types.ProviderType, provider newProviderFunc) {
	availableProviders[name] = provider
}

const imageToTextPrompt = `
请仔细分析这张图片，其中包含一个表情包或梗图。请详细描述表情包的视觉内容，重点关注那些能够传达其含义、情绪或常见用途的元素。你的描述应该简洁但足够全面，能够捕捉到表情包的精髓，以便进行语义搜索。

请考虑以下几个方面：

    角色/人物： 图片中描绘了谁或什么？（例如：一个动漫角色、一个拟人化的动物、某个知名IP形象、原创角色）
    表情/神态： 表达了什么情绪？（例如：开心、悲伤、愤怒、困惑、惊讶、哭笑不得、嘲讽、无语）
    动作/姿态： 角色正在做什么？（例如：比心、抱头、摊手、思考、比耶、滑跪、捂脸）
    道具/物品： 图片中是否有重要的道具或物品？（例如：一把刀、一碗面、一个抱枕、一个墨镜、一个感叹号）
    文字内容（如有）： 请准确转录图片中出现的任何文字。
    背景/场景： 简要描述图片背景或场景，如果它对理解表情包有帮助。（例如：简约的纯色背景、某个特定的动漫场景、生活日常场景）
    常见语境/关联： 简要解释这个表情包通常代表哪种情境、感受或被用来表达什么。（例如：‘用于表达无奈和委屈’、‘表示惊讶和不可思议’、‘是某句流行语的视觉化表现’）

输出示例格式：

    描述： 一个留着粉色短发、穿着校服的动漫少女，张大嘴巴露出惊讶的表情，双眼圆睁，背景为蓝色。
    常见用途： 用于表达极度惊讶、震惊，或者对某些出乎意料的事情感到不可思议。

现在，请分析以下图片并提供类似的描述：
`
