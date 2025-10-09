package cmd

import (
	"context"
	"fmt"
	"mimir/internal/binance"
	"mimir/internal/common"
	"mimir/internal/config"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var today string = time.Now().Format("2006-01-02")

var syncKlineCmd = &cobra.Command{
	Use:   "sync-kline",
	Short: "同步 K 线数据",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[INFO] starting sync KLines ...")

		ctx := context.Background()

		common.Init()
		defer common.Close()

		operate := binance.BinanceUsdM{}

		rows, err := common.DB.QueryContext(ctx, "SELECT `name` FROM `mimir`.`symbols` ORDER BY `id` ASC")
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		// _, err = common.DB.ExecContext(ctx, "TRUNCATE TABLE `mimir`.`market_15m`")
		// if err != nil {
		// 	panic(err)
		// }
		for rows.Next() {
			var symbol string
			if err := rows.Scan(&symbol); err != nil {
				panic(err)
			}

			dailyKline(ctx, operate, symbol)
			fifteenMinutesKline(ctx, operate, symbol)
			forHourKline(ctx, operate, symbol)
		}

		// // 15m volume surge
		// miningVolumeSurge(ctx)

		fmt.Println("[INFO] successfully synced KLines")
	},
}

func init() {
	rootCmd.AddCommand(syncKlineCmd)
}

func dailyKline(ctx context.Context, operate binance.BinanceUsdM, symbol string) {
	klines, err := operate.KLines(symbol, "1d", 150)
	if err != nil {
		panic(err)
	}

	var prices []float64
	for _, kline := range klines {
		closePrice, err := strconv.ParseFloat(kline.Close, 64)
		if err != nil {
			panic(err)
		}
		prices = append(prices, closePrice)
	}
	ma5 := binance.MovingAverage(prices, 5)
	ma10 := binance.MovingAverage(prices, 10)
	ma20 := binance.MovingAverage(prices, 20)
	ma50 := binance.MovingAverage(prices, 50)

	for idx, kline := range klines {
		date := time.UnixMilli(kline.OpenTime).Format("2006-01-02")

		var changeClose, changeVolume float64
		isVolumeSpike := "no"
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

			if changeVolume > 0.5 {
				isVolumeSpike = "yes"
			}
		}

		var cMa5, cMa10, cMa20, cMa50 float64

		if idx < len(ma5) {
			cMa5 = ma5[idx]
		}
		if idx < len(ma10) {
			cMa10 = ma10[idx]
		}
		if idx < len(ma20) {
			cMa20 = ma20[idx]
		}
		if idx < len(ma50) {
			cMa50 = ma50[idx]
		}

		_, err := common.DB.ExecContext(
			ctx,
			"INSERT INTO `mimir`.`market_daily` (`symbol`, `date`, `close`, `volume`, `change_close`, `change_volume`, `ma5`, `ma10`, `ma20`, `ma50`, `is_volume_spike`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE `close` = ?, `volume` = ?, `change_close` = ?, `change_volume` = ?, `ma5` = ?, `ma10` = ?, `ma20` = ?, `ma50` = ?, `is_volume_spike` = ?",
			symbol,
			date,
			kline.Close,
			kline.Volume,
			changeClose,
			changeVolume,
			cMa5,
			cMa10,
			cMa20,
			cMa50,
			isVolumeSpike,
			kline.Close,
			kline.Volume,
			changeClose,
			changeVolume,
			cMa5,
			cMa10,
			cMa20,
			cMa50,
			isVolumeSpike,
		)
		if err != nil {
			panic(err)
		}
	}
}

func fifteenMinutesKline(ctx context.Context, operate binance.BinanceUsdM, symbol string) {
	klines, err := operate.KLines(symbol, "15m", 4)
	if err != nil {
		panic(err)
	}

	for idx, kline := range klines {
		date := time.UnixMilli(kline.OpenTime).Format("2006-01-02")
		hourMinute := time.UnixMilli(kline.OpenTime).Format("1504")

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

		if changeClose != 0 && changeVolume != 0 {
			_, err := common.DB.ExecContext(
				ctx,
				"INSERT INTO `mimir`.`market_15m` (`symbol`, `date`, `hour_minute`, `close`, `volume`, `change_close`, `change_volume`) VALUES (?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE `close` = ?, `volume` = ?, `change_close` = ?, `change_volume` = ?",
				symbol,
				date,
				hourMinute,
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
}

func forHourKline(ctx context.Context, operate binance.BinanceUsdM, symbol string) {
	klines, err := operate.KLines(symbol, "4h", 150)
	if err != nil {
		panic(err)
	}

	var prices []float64
	for _, kline := range klines {
		closePrice, err := strconv.ParseFloat(kline.Close, 64)
		if err != nil {
			panic(err)
		}
		prices = append(prices, closePrice)
	}
	ma5 := binance.MovingAverage(prices, 5)
	ma10 := binance.MovingAverage(prices, 10)
	ma20 := binance.MovingAverage(prices, 20)
	ma50 := binance.MovingAverage(prices, 50)

	for idx, kline := range klines {
		date := time.UnixMilli(kline.OpenTime).Format("2006-01-02")
		hourMinute := time.UnixMilli(kline.OpenTime).Format("1504")

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

		var cMa5, cMa10, cMa20, cMa50 float64

		if idx < len(ma5) {
			cMa5 = ma5[idx]
		}
		if idx < len(ma10) {
			cMa10 = ma10[idx]
		}
		if idx < len(ma20) {
			cMa20 = ma20[idx]
		}
		if idx < len(ma50) {
			cMa50 = ma50[idx]
		}

		_, err := common.DB.ExecContext(
			ctx,
			"INSERT INTO `mimir`.`market_4h` (`symbol`, `date`, `hour_minute`, `close`, `volume`, `change_close`, `change_volume`, `ma5`, `ma10`, `ma20`, `ma50`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE `close` = ?, `volume` = ?, `change_close` = ?, `change_volume` = ?, `ma5` = ?, `ma10` = ?, `ma20` = ?, `ma50` = ?",
			symbol,
			date,
			hourMinute,
			kline.Close,
			kline.Volume,
			changeClose,
			changeVolume,
			cMa5,
			cMa10,
			cMa20,
			cMa50,
			kline.Close,
			kline.Volume,
			changeClose,
			changeVolume,
			cMa5,
			cMa10,
			cMa20,
			cMa50,
		)
		if err != nil {
			panic(err)
		}
	}
}

func miningVolumeSurge(ctx context.Context) {
	// var btcChangeClose, btcChangeVolume float64
	rows, err := common.DB.QueryContext(ctx, "SELECT `hour_minute`, `change_close`, `change_volume` FROM `mimir`.`market_15m` WHERE `symbol` = 'BTCUSDT' AND `date` = ? ORDER BY `hour_minute` DESC LIMIT 4", today)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var maxChangeVolume, findChangeClose float64
	var queryHourMinute string
	for rows.Next() {
		var hourMinute string
		var changeClose, changeVolume float64
		if err := rows.Scan(&hourMinute, &changeClose, &changeVolume); err != nil {
			panic(err)
		}

		if changeVolume > maxChangeVolume {
			maxChangeVolume = changeVolume
			findChangeClose = changeClose
			queryHourMinute = hourMinute
		}

	}

	fmt.Printf("[INFO] BTCUSDT %s %s 漲幅 %.8f 量能 %.8fx\n", today, queryHourMinute, findChangeClose, maxChangeVolume)
	if findChangeClose < 0 && maxChangeVolume > 3 {
		message := "🚀 交易量激增＋BTC跌勢，抗跌且交易量放大:\n\n"

		rows, err = common.DB.QueryContext(ctx, "SELECT `symbol`, `change_close`, `change_volume` FROM `mimir`.`market_15m` WHERE `symbol` != 'BTCUSDT' AND `date` = ? AND `hour_minute` = ? AND `change_volume` > 2 ORDER BY `change_close` ASC LIMIT 10", today, queryHourMinute)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		for rows.Next() {
			var symbol string
			var changeClose, changeVolume float64
			if err := rows.Scan(&symbol, &changeClose, &changeVolume); err != nil {
				panic(err)
			}

			message += fmt.Sprintf("%s：%.8f %.0fx\n", symbol, changeClose, changeVolume)
		}

		tg := common.NewTelegramNotifier(
			config.Cfg.Telegram.BotToken,
			config.Cfg.Telegram.ChatID,
		)

		err = tg.Notify(message)
		if err != nil {
			panic(err)
		}
	}
}
