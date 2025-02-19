package search

import (
	"fmt"
	"strings"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/samber/lo"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database/ent"
	"github.com/xyenon/telemikiya/libs"
	"github.com/xyenon/telemikiya/searcher"
	"github.com/xyenon/telemikiya/types"
	"go.uber.org/zap"
)

func Search(ctx *ext.Context, update *ext.Update) error {
	logger := ctx.Value(types.CtxKeyLogger{}).(*zap.Logger)
	cfg := ctx.Value(types.CtxKeyConfigTelegram{}).(*config.Telegram)
	s := ctx.Value(types.CtxKeySearcher{}).(*searcher.Searcher)

	userID := update.EffectiveUser().GetID()
	if !lo.Contains(cfg.BotAllowedUserIDs, userID) {
		return fmt.Errorf("user %d is not allowed to use this bot", userID)
	}

	input := strings.TrimPrefix(update.EffectiveMessage.Text, "/search")
	input = strings.TrimSpace(input)
	logger.Info("searching messages", zap.String("text", input))

	params := searcher.SearchParams{
		Input: input,
		Count: 10,
	}
	messages, err := s.Search(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to search messages: %w", err)
	}

	messageStyledTextOptions := lo.FlatMap(messages,
		func(message *ent.Message, i int) []styling.StyledTextOption {
			opts := []styling.StyledTextOption{
				styling.Bold(fmt.Sprintf("%d. ", i+1)),
				styling.TextURL(message.Text+"\n", libs.DeepLink(message)),
			}
			if i < len(messages)-1 {
				opts = append(opts, styling.Plain("==========\n"))
			}
			return opts
		},
	)
	_, err = ctx.Reply(update, ext.ReplyTextStyledTextArray(messageStyledTextOptions), nil)

	return err
}
