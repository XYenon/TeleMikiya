package bot

import (
	"context"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database"
	"github.com/xyenon/telemikiya/searcher"
	"github.com/xyenon/telemikiya/telegram/bot/search"
	"github.com/xyenon/telemikiya/types"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
)

type Params struct {
	fx.In

	LifeCycle fx.Lifecycle
	Config    *config.Config
	Logger    *zap.Logger
	Searcher  *searcher.Searcher
	Database  *database.Database
}

type Bot struct {
	cfg      *config.Telegram
	logger   *zap.Logger
	searcher *searcher.Searcher
	db       *database.Database

	tgClient *gotgproto.Client
}

func New(params Params) *Bot {
	bot := &Bot{
		cfg:      &params.Config.Telegram,
		logger:   params.Logger,
		searcher: params.Searcher,
		db:       params.Database,
	}

	if params.LifeCycle != nil {
		params.LifeCycle.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				bot.Stop()
				return nil
			},
		})
	}

	return bot
}

func (o *Bot) Run() (err error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, types.CtxKeyLogger{}, o.logger)
	ctx = context.WithValue(ctx, types.CtxKeyConfigTelegram{}, o.cfg)
	ctx = context.WithValue(ctx, types.CtxKeySearcher{}, o.searcher)

	if o.tgClient, err = gotgproto.NewClient(
		o.cfg.APIID,
		o.cfg.APIHash,
		gotgproto.ClientTypeBot(o.cfg.BotToken),
		&gotgproto.ClientOpts{
			Logger: o.logger,
			Session: sessionMaker.SqlSession(
				postgres.New(postgres.Config{Conn: o.db.BotSessionConn}),
			),
			Context: ctx,
		},
	); err != nil {
		return err
	}

	o.logger.Info("telegram bot has been started", zap.String("username", o.tgClient.Self.Username))
	dispatcher := o.tgClient.Dispatcher
	dispatcher.AddHandler(handlers.NewCommand("search", search.Search))

	return o.tgClient.Idle()
}

func (o *Bot) Stop() {
	o.tgClient.Stop()
	o.logger.Info("telegram bot has been stopped")
}
