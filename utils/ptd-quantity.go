// Package utils encapsulates all the utilities.
package utils

import (
	"fmt"
	"math"

	"github.com/jmvdr-iscte/TradingBotCli/enums"
)

const (
	lowResponseThreshold       = 20
	mediumResponseThreshold    = 10
	highResponseThreshold      = 5
	veryHighResponseThreshold  = 0
	maxResponseThreshold       = 80
	nearMaxResponseThreshold   = 90
	almostMaxResponseThreshold = 95
	maxResponseValue           = 100
)

// SellPDTQuantity returns the quantity of the sell, given the buying power, risk, sentiment
// response and the latest quote of the stock.
func SellPDTQuantity(response int64, buying_power float64, latest_quote float64, risk enums.Risk) int64 {
	multiplier, err := RiskPDTMultiplier(risk)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	switch {
	case response > lowResponseThreshold:
		return int64(math.Abs((math.Min(buying_power*0.02/latest_quote, 4) * multiplier)))
	case response > mediumResponseThreshold:
		return int64(math.Abs((math.Min(buying_power*0.05/latest_quote, 10) * multiplier)))
	case response > highResponseThreshold:
		return int64(math.Abs((math.Min(buying_power*0.07/latest_quote, 14) * multiplier)))
	case response >= veryHighResponseThreshold:
		return int64(math.Abs((math.Min(buying_power*0.10/latest_quote, 20) * multiplier)))
	}
	return 0
}

// BuyPDTQuantity returns the quantity of the buy, given the buying power, risk, sentiment,
// response and the latest quote of the stock.
func BuyPDTQuantity(response int64, buying_power float64, latest_quote float64, risk enums.Risk) int64 {

	multiplier, err := RiskPDTMultiplier(risk)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	switch {
	case response < maxResponseThreshold:
		return int64(math.Abs((math.Min(buying_power*0.02/latest_quote, 4) * multiplier)))
	case response < nearMaxResponseThreshold:
		return int64(math.Abs((math.Min(buying_power*0.05/latest_quote, 10) * multiplier)))
	case response < almostMaxResponseThreshold:
		return int64(math.Abs((math.Min(buying_power*0.07/latest_quote, 14) * multiplier)))
	case response <= maxResponseValue:
		return int64(math.Abs((math.Min(buying_power*0.10/latest_quote, 20) * multiplier)))
	}
	return 0
}

// RiskPDTMultiplier returns the multiplier, given the risk.
func RiskPDTMultiplier(risk enums.Risk) (float64, error) {
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
