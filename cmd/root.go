package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gounder",
	Short: "GoUnder is a tool for pre-penetration built by go",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {}
