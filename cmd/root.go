package cmd

import (
	"fmt"
	"mimir/internal/config"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "mimir",
	Short: "Mimir CLI Tool",
	Long:  `用來同步或分析 Coin Trade 資料的 CLI 工具`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			viper.AddConfigPath(".")
			viper.SetConfigName(".env")
			viper.SetConfigType("yaml")
		}

		err := viper.ReadInConfig()
		if err != nil {
			fmt.Println("[INFO] load failed %w", err)
			os.Exit(1)
		}

		config.LoadFromViper()
		fmt.Println("[INFO] loaded ", viper.ConfigFileUsed())
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "設定檔路徑 (預設為 ./.env.yaml)")
}
