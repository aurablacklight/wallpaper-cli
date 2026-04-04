package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove duplicates and invalid files",
	Long:  `Clean up the wallpaper collection by removing duplicates and invalid files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Cleaning wallpaper collection...")
		return fmt.Errorf("not yet implemented - see S05, S06")
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
