package searcher

import (
	"context"

	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/searcher"
	"github.com/xyenon/telemikiya/telegram"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	LifeCycle fx.Lifecycle
	Config    *config.Config
	Logger    *zap.Logger
	Telegram  *telegram.Telegram `name:"tg_bot"`
	Searcher  *searcher.Searcher
}

type Searcher struct {
	cfg      *config.Telegram
	logger   *zap.Logger
	tg       *telegram.Telegram
	searcher *searcher.Searcher
}

func New(params Params) *Searcher {
	s := &Searcher{
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

func (s Searcher) Start() {
	dispatcher := s.tg.Dispatcher
	dispatcher.AddHandler(handlers.NewCommand("search", s.search))
}
