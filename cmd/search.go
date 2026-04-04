package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search sources without downloading",
	Long:  `Search wallpaper sources and preview results without downloading.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Searching wallpaper sources...")
		return fmt.Errorf("not yet implemented - see S03")
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
