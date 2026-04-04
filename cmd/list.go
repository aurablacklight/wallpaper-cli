package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List already downloaded wallpapers",
	Long:  `List and filter downloaded wallpapers from the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Listing downloaded wallpapers...")
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
