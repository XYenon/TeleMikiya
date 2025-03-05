package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/xyenon/telemikiya/database"
	"go.uber.org/fx"
)

var allowClearEmbedding bool

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database schema to latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := fx.New(
			fxOptions(),
			fx.Supply(fx.Annotate(allowClearEmbedding, fx.ResultTags(`name:"allowClearEmbedding"`))),
			fx.Invoke(func(db *database.Database) error {
				return db.Migrate(context.Background())
			}),
		)

		return app.Start(context.Background())
	},
}

func init() {
	dbCmd.AddCommand(dbMigrateCmd)
	dbMigrateCmd.Flags().BoolVar(&allowClearEmbedding, "allow-clear-embedding", false, "allow clearing embedding when embedding dimensions change")
}
