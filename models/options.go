package models

import "github.com/jmvdr-iscte/TradingBot/enums"

type Options struct {
	Risk          enums.Risk `json:"risk"`
	Gain          float64    `json:"gain"`
	StartingValue float64    `json:"starting_value"`
}
