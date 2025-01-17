package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = &cobra.Command{
	Use:   "version",
	Short: "specifies the build version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Parent().Version)
	},
}
