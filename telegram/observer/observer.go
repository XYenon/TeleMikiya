package observer

import (
	"context"
	"sync"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database"
	"github.com/xyenon/telemikiya/telegram/observer/record"
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
	Database  *database.Database
}

type Observer struct {
	cfg    *config.Telegram
	logger *zap.Logger
	db     *database.Database

	tgClient *gotgproto.Client
}

func New(params Params) *Observer {
	observer := &Observer{
		cfg:    &params.Config.Telegram,
		logger: params.Logger,
		db:     params.Database,
	}

	if params.LifeCycle != nil {
		params.LifeCycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go observer.Run()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				observer.Stop()
				return nil
			},
		})
	}

	return observer
}

func (o *Observer) Run() (err error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, types.CtxKeyLogger{}, o.logger)
	ctx = context.WithValue(ctx, types.CtxKeyConfigTelegram{}, o.cfg)
	ctx = context.WithValue(ctx, types.ContextKeyDB{}, o.db)
	ctx = context.WithValue(ctx, types.CtxKeyDialogLock{}, &sync.Map{})

	if o.tgClient, err = gotgproto.NewClient(
		o.cfg.APIID,
		o.cfg.APIHash,
		gotgproto.ClientTypePhone(o.cfg.PhoneNumber),
		&gotgproto.ClientOpts{
			Logger: o.logger,
			Session: sessionMaker.SqlSession(
				postgres.New(postgres.Config{Conn: o.db.UserSessionConn}),
			),
			Context: ctx,
		},
	); err != nil {
		return err
	}

	o.logger.Info("telegram observer has been started", zap.String("username", o.tgClient.Self.Username))
	dispatcher := o.tgClient.Dispatcher
	dispatcher.AddHandler(handlers.NewMessage(filters.Message.Text, record.Recorder))

	return o.tgClient.Idle()
}

func (o *Observer) Stop() {
	o.tgClient.Stop()
	o.logger.Info("telegram observer has been stopped")
}
