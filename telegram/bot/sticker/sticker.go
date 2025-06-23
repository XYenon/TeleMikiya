package sticker

import (
	"context"
	"fmt"

	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/provider"
	tgbot "github.com/xyenon/telemikiya/telegram/bot"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/telebot.v4"
)

type Params struct {
	fx.In

	LifeCycle   fx.Lifecycle
	Config      *config.Config
	Logger      *zap.Logger
	Telegram    *tgbot.TelegramBot
	LLMProvider provider.ImageToTextProvider
}

type Sticker struct {
	cfg         *config.Telegram
	logger      *zap.Logger
	tg          *tgbot.TelegramBot
	llmProvider provider.ImageToTextProvider
}

func New(params Params) *Sticker {
	s := &Sticker{
		cfg:         &params.Config.Telegram,
		logger:      params.Logger,
		tg:          params.Telegram,
		llmProvider: params.LLMProvider,
	}

	if params.LifeCycle != nil {
		params.LifeCycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				s.Start()
				return nil
			},
		})
	}

	return s
}

func (s Sticker) Start() {
	s.tg.Handle(telebot.OnSticker, s.indexStickerSet)
}

func (s Sticker) indexStickerSet(c telebot.Context) error {
	stickerSet, err := c.Bot().StickerSet(c.Message().Sticker.SetName)
	if err != nil {
		return fmt.Errorf("unable to get sticker set: %w", err)
	}
	var buf []byte
	for _, sticker := range stickerSet.Stickers {
		s.logger.Debug("fetch sticker", zap.String("url", sticker.FileURL))
		if _, err = sticker.FileReader.Read(buf); err != nil {
			s.logger.Error("fetch sticker", zap.Error(err))
		}
		break
	}

	ctx := context.Background()
	func() {
		text, err := s.llmProvider.ImageToText(ctx, "image/webp", buf)
		if err != nil {
			s.logger.Error("failed to convert sticker to text", zap.Error(err))
			return
		}
		s.logger.Debug("sticker text", zap.String("text", text))
	}()
	return nil
}
