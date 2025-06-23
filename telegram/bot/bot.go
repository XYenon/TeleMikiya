package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/telebot.v4"
)

type Params struct {
	fx.In

	LifeCycle fx.Lifecycle
	Config    *config.Config
	Logger    *zap.Logger
	Database  *database.Database
	Debug     bool `name:"debug" optional:"true"`
}

type TelegramBot struct {
	logger *zap.Logger

	*telebot.Bot
	username string
}

func New(params Params) (*TelegramBot, error) {
	cfg := &params.Config.Telegram
	bot, err := telebot.NewBot(telebot.Settings{
		Token:     cfg.BotToken,
		Poller:    &telebot.LongPoller{Timeout: 30 * time.Second},
		Verbose:   params.Debug,
		ParseMode: telebot.ModeMarkdownV2,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot client: %w", err)
	}

	tg := &TelegramBot{
		logger: params.Logger,
		Bot:    bot,
	}

	if params.LifeCycle != nil {
		params.LifeCycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				tg.Start()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				tg.Stop()
				return nil
			},
		})
	}

	return tg, nil
}

func (t *TelegramBot) Start() {
	t.Bot.Start()
	t.username = t.Bot.Me.Username
	t.logger.Info("telegram bot client has been started", zap.String("username", t.username))
}

func (t *TelegramBot) Stop() {
	t.Bot.Stop()
	t.logger.Info("telegram bot client has been stopped", zap.String("username", t.username))
}
