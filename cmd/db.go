package cmd

import "github.com/spf13/cobra"

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long: `Database management commands for TeleMikiya.
These commands help you manage the database schema and data.`,
}

func init() {
	rootCmd.AddCommand(dbCmd)
}
