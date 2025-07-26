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

		message := fmt.Sprintf("📊 MA 挖掘推薦幣別 (%s ~ %s):\n\n", beginDate, endDate)
		for rows.Next() {
			var symbol string
			var count int
			if err := rows.Scan(&symbol, &count); err != nil {
				panic(err)
			}

			var spikeCount int
			err := common.DB.QueryRowContext(ctx, `
				SELECT count(1) as count 
				FROM mimir.market_daily 
				WHERE symbol = ? 
				AND date BETWEEN ? AND ?
				AND is_volume_spike = 'yes'`,
				symbol,
				beginDate,
				endDate,
			).Scan(&spikeCount)
			if err != nil {
				panic(err)
			}

			var emoji string
			switch {
			case spikeCount == 0:
				emoji = "⚠️"
			case spikeCount <= 2:
				emoji = "🚀"
			case spikeCount <= 5:
				emoji = "🚀🔥"
			default:
				emoji = "🚀🔥💥"
			}

			message += fmt.Sprintf("%s %s：爆量 %d 次\n", emoji, symbol, spikeCount)
		}

		tg := common.NewTelegramNotifier(
			config.Cfg.Telegram.BotToken,
			config.Cfg.Telegram.ChatID,
		)

		err = tg.Notify(message)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(miningMaCmd)
}
