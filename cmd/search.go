package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/xyenon/telemikiya/libs"
	"github.com/xyenon/telemikiya/searcher"
	"go.uber.org/fx"
)

var (
	count        uint
	startTimeStr string
	endTimeStr   string
	dialogID     int64

	startTime time.Time
	endTime   time.Time
)

var searchCmd = &cobra.Command{
	Use:   "search <keywords...>",
	Short: "Search for messages using hybrid search",
	Long: `Search for messages using a combination of semantic similarity and full-text search.
The results are ranked based on both semantic relevance and text matching scores.`,
	Example: `  telemikiya search how is the weather today
  telemikiya search --count 20 --dialog-id 123456789 recommend a movie
  telemikiya search --start-time "2024-01-01 00:00:00" happy new year`,
	ValidArgs: []string{"keywords"},
	Args:      cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if lo.IsNotEmpty(startTimeStr) {
			if startTime, err = time.Parse(time.DateTime, startTimeStr); err != nil {
				return fmt.Errorf("failed to parse start time: %w", err)
			}
		}
		if lo.IsNotEmpty(endTimeStr) {
			if endTime, err = time.Parse(time.DateTime, endTimeStr); err != nil {
				return fmt.Errorf("failed to parse end time: %w", err)
			}
		}

		return
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		app := fx.New(
			fxOptions(),
			fx.Invoke(func(s *searcher.Searcher) error {
				params := searcher.SearchParams{
					Input:     strings.Join(args, " "),
					Count:     count,
					StartTime: startTime,
					EndTime:   endTime,
					DialogID:  dialogID,
				}

				messages, err := s.Search(context.Background(), params)
				if err != nil {
					return err
				}

				for i, message := range messages {
					fmt.Printf("%d. %s\n", i+1, libs.DeepLink(message))
					fmt.Println(libs.Indent(message.Text, 4))
					if i < len(messages)-1 {
						fmt.Println()
					}
				}

				return nil
			}),
		)

		return app.Start(context.Background())
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().UintVarP(&count, "count", "c", 10, "maximum number of messages to return")
	searchCmd.Flags().StringVar(&startTimeStr, "start-time", "", "search messages after this time (format: YYYY-MM-DD HH:mm:ss)")
	searchCmd.Flags().StringVar(&endTimeStr, "end-time", "", "search messages before this time (format: YYYY-MM-DD HH:mm:ss)")
	searchCmd.Flags().Int64Var(&dialogID, "dialog-id", 0, "search in specific dialog")
}
