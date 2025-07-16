package cmd

import (
	"context"
	"fmt"
	"mimir/internal/binance"
	"mimir/internal/common"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var syncKlineCmd = &cobra.Command{
	Use:   "sync-kline",
	Short: "同步 K 线数据",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[INFO] starting sync KLines ...")

		ctx := context.Background()

		common.Init()
		defer common.Close()

		operate := binance.BinanceUsdM{}

		rows, err := common.DB.QueryContext(ctx, "SELECT `name` FROM `mimir`.`symbols`")
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		for rows.Next() {
			var symbol string
			if err := rows.Scan(&symbol); err != nil {
				panic(err)
			}

			klines, err := operate.KLines(symbol)
			if err != nil {
				panic(err)
			}

			for idx, kline := range klines {
				date := time.UnixMilli(kline.OpenTime).Format("2006-01-02")

				var changeClose, changeVolume float64
				if idx != 0 {
					prevKine := klines[idx-1]

					close, err := strconv.ParseFloat(kline.Close, 64)
					if err != nil {
						panic(err)
					}
					volume, err := strconv.ParseFloat(kline.Volume, 64)
					if err != nil {
						panic(err)
					}

					prevClose, err := strconv.ParseFloat(prevKine.Close, 64)
					if err != nil {
						panic(err)
					}
					prevVolume, err := strconv.ParseFloat(prevKine.Volume, 64)
					if err != nil {
						panic(err)
					}

					changeClose = (close - prevClose) / prevClose
					changeVolume = (volume - prevVolume) / prevVolume
				}

				_, err := common.DB.ExecContext(
					ctx,
					"INSERT INTO `mimir`.`market_daily` (`symbol`, `date`, `close`, `volume`, `change_close`, `change_volume`) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE `close` = ?, `volume` = ?, `change_close` = ?, `change_volume` = ?",
					symbol,
					date,
					kline.Close,
					kline.Volume,
					changeClose,
					changeVolume,
					kline.Close,
					kline.Volume,
					changeClose,
					changeVolume,
				)
				if err != nil {
					panic(err)
				}
			}
		}

		fmt.Println("[INFO] successfully synced KLines")
	},
}

func init() {
	rootCmd.AddCommand(syncKlineCmd)
}
