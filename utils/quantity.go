package utils

import (
	"fmt"
	"math"

	"github.com/jmvdr-iscte/TradingBotCli/enums"
)

func SellQuantity(response int64, buying_power float64, latest_quote float64, risk enums.Risk) int64 {
	multiplier, err := RiskMultiplier(risk)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	switch {
	case response > 20:
		return int64(math.Abs((math.Min(buying_power*0.02/latest_quote, 4) * multiplier)))
	case response > 10:
		return int64(math.Abs((math.Min(buying_power*0.05/latest_quote, 10) * multiplier)))
	case response > 5:
		return int64(math.Abs((math.Min(buying_power*0.07/latest_quote, 14) * multiplier)))
	case response >= 0:
		return int64(math.Abs((math.Min(buying_power*0.10/latest_quote, 20) * multiplier)))
	}
	return 0
}

func BuyQuantity(response int64, buying_power float64, latest_quote float64, risk enums.Risk) int64 {

	multiplier, err := RiskMultiplier(risk)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	switch {
	case response < 80:
		return int64(math.Abs((math.Min(buying_power*0.02/latest_quote, 4) * multiplier)))
	case response < 90:
		return int64(math.Abs((math.Min(buying_power*0.05/latest_quote, 10) * multiplier)))
	case response < 95:
		return int64(math.Abs((math.Min(buying_power*0.07/latest_quote, 14) * multiplier)))
	case response <= 100:
		return int64(math.Abs((math.Min(buying_power*0.10/latest_quote, 20) * multiplier)))
	}
	return 0
}

func RiskMultiplier(risk enums.Risk) (float64, error) {
	switch risk {
	case enums.Low:
		return 0.5, nil
	case enums.Medium:
		return 1.0, nil
	case enums.High:
		return 2.0, nil
	default:
		return 0, fmt.Errorf("invalid risk level")
	}
}
