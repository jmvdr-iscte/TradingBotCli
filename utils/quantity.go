package utils

import (
	"math"

	"github.com/jmvdr-iscte/TradingBotCli/enums"
)

func CalculateQuantity(buying_power float64, latest_quote float64, risk enums.Risk) int64 {
	switch risk {
	case enums.Safe:
		return int64(math.Abs((math.Min(buying_power*0.1/latest_quote, 20))))
	case enums.Power:
		return int64(math.Abs((math.Max(buying_power*0.1/latest_quote, 20))))
	default:
		return 0
	}
}
