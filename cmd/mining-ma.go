package cmd

import (
	"context"
	"fmt"
	"mimir/internal/common"
	"mimir/internal/config"
	"time"

	"github.com/spf13/cobra"
)

type MiningSymbolCount struct {
	Symbol string
	Count  int
}

var miningMaCmd = &cobra.Command{
	Use:   "mining-ma",
	Short: "挖掘成型幣別，依靠 MA 指標",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[INFO] starting mining ma ...")

		ctx := context.Background()

		common.Init()
		defer common.Close()

		now := time.Now()
		beginDate := now.AddDate(0, 0, -7).Format("2006-01-02")
		endDate := now.AddDate(0, 0, -1).Format("2006-01-02")

		rows, err := common.DB.QueryContext(ctx, `
			SELECT symbol, count(1) as count
				FROM market_daily 
				WHERE date BETWEEN ? AND ?
				AND ma50 > 0 
				AND close > ma50 
				AND ma5 > ma10 
				AND ma10 > ma20 
				AND ma20 > ma50 
				GROUP BY symbol
				ORDER BY count DESC`,
			beginDate,
			endDate,
		)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		var symbolCounts []MiningSymbolCount
		for rows.Next() {
			var symbol string
			var count int
			if err := rows.Scan(&symbol, &count); err != nil {
				panic(err)
			}

			symbolCounts = append(symbolCounts, MiningSymbolCount{Symbol: symbol, Count: count})
		}

		top10p := int(float64(len(symbolCounts)) * 0.2)

		for i := 0; i < top10p; i++ {
			_, err := common.DB.ExecContext(
				ctx,
				"INSERT INTO `mimir`.`mining_ma_stat` (`date`, `symbol`, `count`) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE `count` = ?",
				endDate,
				symbolCounts[i].Symbol,
				symbolCounts[i].Count,
				symbolCounts[i].Count,
			)
			if err != nil {
				panic(err)
			}
		}

		rows, err = common.DB.QueryContext(ctx, `
			SELECT symbol, SUM(count) as count
				FROM mimir.mining_ma_stat 
				WHERE date BETWEEN ? AND ?
				GROUP BY symbol
				ORDER BY count DESC
				LIMIT 10`,
			beginDate,
			endDate,
		)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		message := "挖掘 MA 數據推薦幣別:\n\n"
		for rows.Next() {
			var symbol string
			var count int
			if err := rows.Scan(&symbol, &count); err != nil {
				panic(err)
			}

			message += fmt.Sprintf("%s\n", symbol)
		}

		tg := common.NewTelegramNotifier(
			config.Cfg.Telegram.BotToken,
			config.Cfg.Telegram.ChatID,
		)

		tg.Notify(message)
	},
}

func init() {
	rootCmd.AddCommand(miningMaCmd)
}
