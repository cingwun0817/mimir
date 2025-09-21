package cmd

import (
	"context"
	"fmt"
	"math"
	"mimir/internal/common"
	"mimir/internal/config"
	"time"

	"github.com/spf13/cobra"
)

var miningStrongCmd = &cobra.Command{
	Use:   "mining-strong",
	Short: "挖掘強勢幣別",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[INFO] starting mining strong ...")

		ctx := context.Background()

		common.Init()
		defer common.Close()

		now := time.Now()
		yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

		var date string
		var changeClose float64
		err := common.DB.QueryRowContext(ctx, `
			SELECT date, change_close
				FROM market_daily 
				WHERE symbol = 'BTCUSDT'
				AND date = ?`,
			yesterday,
		).Scan(&date, &changeClose)
		if err != nil {
			panic(err)
		}

		if math.Abs(changeClose) > 0.015 {
			// 相對跌幅大
			rows, err := common.DB.QueryContext(ctx, `
			SELECT symbol, change_close
				FROM market_daily 
				WHERE date = ?
				AND symbol NOT LIKE '%BTC%'
				AND change_close < ?
				ORDER BY abs(change_close) DESC
				LIMIT 15`,
				yesterday,
				changeClose,
			)
			if err != nil {
				panic(err)
			}
			defer rows.Close()

			tg := common.NewTelegramNotifier(
				config.Cfg.Telegram.BotToken,
				config.Cfg.Telegram.ChatID,
			)

			message := fmt.Sprintf("📊 相對跌幅最大 (%s):\n\n", yesterday)
			for rows.Next() {
				var symbol string
				var changeClose float64
				if err := rows.Scan(&symbol, &changeClose); err != nil {
					panic(err)
				}

				message += fmt.Sprintf("%s：%.0f %%\n", symbol, changeClose*100)
			}

			err = tg.Notify(message)
			if err != nil {
				panic(err)
			}

			// 相對跌幅小
			rows, err = common.DB.QueryContext(ctx, `
			SELECT symbol, change_close
				FROM market_daily 
				WHERE date = ?
				AND symbol NOT LIKE '%BTC%'
				AND change_close > ?
				AND change_close < 0
				ORDER BY abs(change_close) ASC
				LIMIT 15`,
				yesterday,
				changeClose,
			)
			if err != nil {
				panic(err)
			}
			defer rows.Close()

			message = fmt.Sprintf("📊 相對跌幅最小 (%s):\n\n", yesterday)
			for rows.Next() {
				var symbol string
				var changeClose float64
				if err := rows.Scan(&symbol, &changeClose); err != nil {
					panic(err)
				}

				message += fmt.Sprintf("%s：%.4f %%\n", symbol, changeClose*100)
			}

			err = tg.Notify(message)
			if err != nil {
				panic(err)
			}
		}

		fmt.Printf("BTCUSDT %s 漲幅 %.8f%%\n", date, changeClose)
	},
}

func init() {
	rootCmd.AddCommand(miningStrongCmd)
}
