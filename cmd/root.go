package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gounder",
	Short: "GoUnder 是一个用Go编写的安全工具，模拟gobuster风格",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {}
