// Package models serve as structs used in the application.
package models

import "github.com/jmvdr-iscte/TradingBotCli/enums"

// Options is a type used in the configuration of the server.
type Options struct {
	Risk          enums.Risk `json:"risk"`
	Gain          float64    `json:"gain"`
	StartingValue float64    `json:"starting_value"`
}
