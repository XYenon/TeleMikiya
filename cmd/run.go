package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xyenon/telemikiya/embedding"
	"github.com/xyenon/telemikiya/telegram/bot"
	"github.com/xyenon/telemikiya/telegram/observer"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"
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
		fx.New(
			fxOptions(),
			fx.Invoke(func(o *observer.Observer, e *embedding.Embedding, b *bot.Bot) error {
				ctx := context.Background()
				g, ctx := errgroup.WithContext(ctx)

				if enableObserver {
					g.Go(func() error { return o.Run() })
				}
				if enableEmbedding {
					g.Go(func() error { return e.Run() })
				}
				if enableBot {
					g.Go(func() error { return b.Run() })
				}

				return g.Wait()
			}),
		).Run()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVar(&enableObserver, "observer", true, "start telegram message observer service")
	runCmd.Flags().BoolVar(&enableEmbedding, "embedding", true, "start text embedding service")
	runCmd.Flags().BoolVar(&enableBot, "bot", true, "start telegram bot service")
}
