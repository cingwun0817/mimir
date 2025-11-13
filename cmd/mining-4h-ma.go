package cmd

import (
	"context"
	"fmt"
	"mimir/internal/common"
	"mimir/internal/config"
	"time"

	"github.com/spf13/cobra"
)

type MiningSymbol4hData struct {
	Symbol string
	Close  float64
	Ma5    float64
	Ma10   float64
	Ma20   float64
	Ma50   float64
}

var mining4hMaCmd = &cobra.Command{
	Use:   "mining-4h-ma",
	Short: "依照 4H MA 指標挖掘幣別",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[INFO] starting mining 4h ma ...")

		ctx := context.Background()

		common.Init()
		defer common.Close()

		now := time.Now()
		today := now.Format("2006-01-02")
		hour := fmt.Sprintf("%02d00", (now.Hour()/4)*4)
		fmt.Printf("[INFO] querying market_4h for date=%s, hour=%s\n", today, hour)

		rows, err := common.DB.QueryContext(ctx, "SELECT symbol, close, ma5, ma10, ma20, ma50 FROM `market_4h` WHERE `date` = ? AND `hour_minute` = ? AND `close` > `ma50` AND `close` > `ma20` AND `close` > `ma10`",
			today,
			hour,
		)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		var symbol4hData []MiningSymbol4hData
		for rows.Next() {
			var data MiningSymbol4hData
			if err := rows.Scan(&data.Symbol, &data.Close, &data.Ma5, &data.Ma10, &data.Ma20, &data.Ma50); err != nil {
				panic(err)
			}

			if data.Ma5 == 0 || data.Ma10 == 0 || data.Ma20 == 0 || data.Ma50 == 0 {
				continue
			}

			symbol4hData = append(symbol4hData, data)
		}

		var alertSymbols []string
		message := "📊 4H MA 成型:\n\n"
		for _, rowdata := range symbol4hData {
			if rowdata.Close > rowdata.Ma5 && rowdata.Ma5 > rowdata.Ma10 && rowdata.Ma10 > rowdata.Ma20 && rowdata.Ma20 > rowdata.Ma50 {
				var symbolCount int
				common.DB.QueryRowContext(ctx, "SELECT COUNT(1) as count FROM `history_records` WHERE `symbol` = ? AND `type` = ?", rowdata.Symbol, "4h_ma").Scan(&symbolCount)

				var icon string
				if symbolCount == 0 {
					icon = "✨"
				}

				message += fmt.Sprintf("%s %s\n", icon, rowdata.Symbol)

				alertSymbols = append(alertSymbols, rowdata.Symbol)
			}
		}

		common.DB.ExecContext(ctx, "TRUNCATE TABLE `mimir`.`history_records`")
		for _, symbol := range alertSymbols {
			_, err := common.DB.ExecContext(ctx, "INSERT INTO `history_records` (`symbol`, `type`) VALUES (?, ?)", symbol, "4h_ma")
			if err != nil {
				panic(err)
			}
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
	rootCmd.AddCommand(mining4hMaCmd)
}
