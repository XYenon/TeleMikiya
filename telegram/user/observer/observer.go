package observer

import (
	"context"
	"sync"

	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database"
	"github.com/xyenon/telemikiya/telegram"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	LifeCycle fx.Lifecycle
	Config    *config.Config
	Logger    *zap.Logger
	Telegram  *telegram.Telegram `name:"tgUser"`
	DataBase  *database.Database
}

type Observer struct {
	cfg    *config.Telegram
	logger *zap.Logger
	tg     *telegram.Telegram
	db     *database.Database

	dialogLock *sync.Map
}

func New(params Params) *Observer {
	r := &Observer{
		cfg:        &params.Config.Telegram,
		logger:     params.Logger,
		tg:         params.Telegram,
		db:         params.DataBase,
		dialogLock: &sync.Map{},
	}

	if params.LifeCycle != nil {
		params.LifeCycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				r.Start()
				return nil
			},
		})
	}

	return r
}

func (r Observer) Start() {
	dispatcher := r.tg.Dispatcher
	dispatcher.AddHandler(handlers.NewMessage(filters.Message.Text, r.record))
}
