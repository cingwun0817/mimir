package cmd

import (
	"context"
	"fmt"

	"mimir/internal/binance" // Adjust the import path as necessary
	"mimir/internal/common"

	"github.com/spf13/cobra"
)

var syncSymbolCmd = &cobra.Command{
	Use:   "sync-symbol",
	Short: "同步合約幣種",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[INFO] starting sync symbols ...")

		ctx := context.Background()

		operate := binance.BinanceUsdM{}

		symbols, err := operate.GetSymbols()
		if err != nil {
			panic(err)
		}

		common.Init()
		defer common.Close()

		for _, symbol := range symbols {
			if symbol.Status != "TRADING" {
				continue
			}

			var dataId int
			common.DB.QueryRowContext(ctx, "SELECT `id` FROM `mimir`.`symbols` WHERE `name` = ?", symbol.Symbol).Scan(&dataId)

			if dataId == 0 {
				_, err = common.DB.Exec("INSERT INTO `mimir`.`symbols` (`name`) VALUES (?)", symbol.Symbol)
				if err != nil {
					panic(err)
				}
			}
		}

		fmt.Println("[INFO] successfully synced symbols")
	},
}

func init() {
	rootCmd.AddCommand(syncSymbolCmd)
}
