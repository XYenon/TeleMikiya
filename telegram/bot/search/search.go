package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database/ent"
	"github.com/xyenon/telemikiya/libs"
	"github.com/xyenon/telemikiya/searcher"
	tgbot "github.com/xyenon/telemikiya/telegram/bot"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/telebot.v4"
)

type Params struct {
	fx.In

	LifeCycle fx.Lifecycle
	Config    *config.Config
	Logger    *zap.Logger
	Telegram  *tgbot.TelegramBot
	Searcher  *searcher.Searcher
}

type Search struct {
	cfg      *config.Telegram
	logger   *zap.Logger
	tg       *tgbot.TelegramBot
	searcher *searcher.Searcher
}

func New(params Params) *Search {
	s := &Search{
		cfg:      &params.Config.Telegram,
		logger:   params.Logger,
		tg:       params.Telegram,
		searcher: params.Searcher,
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

func (s Search) Start() {
	s.tg.Handle("/search", s.search)
}

func (s Search) search(c telebot.Context) error {
	userID := c.Sender().ID
	if !lo.Contains(s.cfg.BotAllowedUserIDs, userID) {
		return fmt.Errorf("user %d is not allowed to use this bot", userID)
	}

	input := c.Message().Payload
	s.logger.Info("searching messages", zap.String("text", input))

	ctx := context.Background()
	params := searcher.SearchParams{
		Input: input,
		Count: 10,
	}
	messages, err := s.searcher.Search(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to search messages: %w", err)
	}

	replyMessageParts := lo.Map(messages, func(message *ent.Message, i int) string {
		return fmt.Sprintf("*%d.* %s\n", i+1, libs.DeepLink(message))
	})
	replyMessage := strings.Join(replyMessageParts, "==========\n")
	return c.Reply(replyMessage)
}
