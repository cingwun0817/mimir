package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mimir",
	Short: "Mimir CLI Tool",
	Long:  `用來同步或分析 Coin Trade 資料的 CLI 工具`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
