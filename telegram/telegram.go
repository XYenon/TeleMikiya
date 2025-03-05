package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Params struct {
	fx.In

	LifeCycle fx.Lifecycle
	Config    *config.Config
	Logger    *zap.Logger
	Database  *database.Database
}

type Telegram struct {
	logger *zap.Logger

	*gotgproto.Client
	username string
}

func NewBot(params Params) (*Telegram, error) {
	return New(params, "bot")
}

func NewUser(params Params) (*Telegram, error) {
	return New(params, "phone")
}

func New(params Params, clientType string) (*Telegram, error) {
	cfg := &params.Config.Telegram

	// gotgproto.clientType is not exported, so we can't use it directly
	tgClientType := gotgproto.ClientTypeBot("")
	var sessionDialector gorm.Dialector
	switch clientType {
	case "phone":
		tgClientType = gotgproto.ClientTypePhone(cfg.PhoneNumber)
		sessionDialector = postgres.New(postgres.Config{Conn: params.Database.UserSessionConn})
	case "bot":
		tgClientType = gotgproto.ClientTypeBot(cfg.BotToken)
		sessionDialector = postgres.New(postgres.Config{Conn: params.Database.BotSessionConn})
	default:
		return nil, fmt.Errorf("invalid client type: %s", clientType)
	}

	client, err := gotgproto.NewClient(
		cfg.APIID, cfg.APIHash,
		tgClientType,
		&gotgproto.ClientOpts{
			Logger:  params.Logger,
			Session: sessionMaker.SqlSession(sessionDialector),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram client: %w", err)
	}

	tg := &Telegram{
		logger: params.Logger,
		Client: client,
	}

	if params.LifeCycle != nil {
		params.LifeCycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go tg.Run()
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

func (t *Telegram) Run() (err error) {
	if username, ok := t.Client.Self.GetUsername(); ok {
		t.username = username
	} else {
		firstName, _ := t.Client.Self.GetFirstName()
		lastName, _ := t.Client.Self.GetLastName()
		t.username = strings.Join([]string{firstName, lastName}, " ")
	}
	t.logger.Info("telegram client has been started", zap.String("username", t.username))

	return t.Client.Idle()
}

func (t *Telegram) Stop() {
	t.Client.Stop()
	t.logger.Info("telegram client has been stopped", zap.String("username", t.username))
}
