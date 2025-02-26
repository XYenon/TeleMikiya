package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xyenon/telemikiya/embedding"
	"github.com/xyenon/telemikiya/telegram/bot"
	"github.com/xyenon/telemikiya/telegram/observer"
	"go.uber.org/fx"
)

var enableObserver, enableEmbedding, enableBot bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start TeleMikiya services",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !enableObserver && !enableEmbedding && !enableBot {
			return fmt.Errorf("at least one service should be enabled")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		opts := []fx.Option{fxOptions()}
		if enableObserver {
			opts = append(opts, fx.Invoke(func(*observer.Observer) {}))
		}
		if enableEmbedding {
			opts = append(opts, fx.Invoke(func(*embedding.Embedding) {}))
		}
		if enableBot {
			opts = append(opts, fx.Invoke(func(*bot.Bot) {}))
		}
		fx.New(opts...).Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVar(&enableObserver, "observer", true, "start telegram message observer service")
	runCmd.Flags().BoolVar(&enableEmbedding, "embedding", true, "start text embedding service")
	runCmd.Flags().BoolVar(&enableBot, "bot", true, "start telegram bot service")
}
