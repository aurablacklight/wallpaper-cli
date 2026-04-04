package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Incremental update of existing collection",
	Long:  `Update an existing wallpaper collection by fetching new images since last update.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Updating wallpaper collection...")
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
