package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database"
	"github.com/xyenon/telemikiya/embedding"
	"github.com/xyenon/telemikiya/embedding/provider"
	"github.com/xyenon/telemikiya/libs"
	"github.com/xyenon/telemikiya/searcher"
	"github.com/xyenon/telemikiya/telegram"
	tgbotsearcher "github.com/xyenon/telemikiya/telegram/bot/searcher"
	"github.com/xyenon/telemikiya/telegram/user/observer"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

var (
	cfgFile string
	debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "telemikiya",
	Short: "TeleMikiya - A hybrid Telegram message search tool",
	Long: `TeleMikiya is a Telegram message search tool that combines:
- Semantic similarity search using vector embeddings
- Full-text search powered by PGroonga
- Automatic message syncing and indexing
- Both CLI and Bot interaction methods`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func fxOptions() fx.Option {
	opts := fx.Options(
		fx.Supply(
			fx.Annotate(debug, fx.ResultTags(`name:"debug"`)),
			fx.Annotate(cfgFile, fx.ResultTags(`name:"cfgFile"`)),
		),
		fx.Provide(fx.Annotate(libs.NewLogger, fx.ParamTags(`name:"debug"`))),
		fx.Provide(fx.Annotate(config.New, fx.ParamTags(`name:"cfgFile"`))),
		fx.Invoke(func(logger *zap.Logger, cfg *config.Config) {
			logger.Debug("config loaded", zap.Reflect("config", cfg))
		}),
		fx.Provide(database.New),
		fx.Provide(provider.New),
		fx.Provide(searcher.New),
		fx.Provide(embedding.New),
		fx.Provide(
			observer.New,
			tgbotsearcher.New,
			fx.Annotate(telegram.NewUser, fx.ResultTags(`name:"tgUser"`)),
			fx.Annotate(telegram.NewBot, fx.ResultTags(`name:"tgBot"`)),
		),
	)

	if debug {
		opts = fx.Options(opts,
			fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{Logger: logger}
			}),
		)
	} else {
		opts = fx.Options(opts, fx.NopLogger)
	}

	return opts
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "C", "", "path to the configuration file")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "D", false, "enable debug logging")
}
