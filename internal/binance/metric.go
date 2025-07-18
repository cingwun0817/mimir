package binance

func MovingAverage(prices []float64, period int) []float64 {
	if period <= 0 || len(prices) < period {
		return nil
	}

	ma := make([]float64, len(prices))
	var sum float64

	for i := 0; i < len(prices); i++ {
		sum += prices[i]

		if i >= period {
			sum -= prices[i-period]
			ma[i] = sum / float64(period)
		} else if i == period-1 {
			ma[i] = sum / float64(period)
		} else {
			ma[i] = 0
		}
	}

	return ma
}
