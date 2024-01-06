// Package models serve as structs used in the application.
package models

import (
	"github.com/google/uuid"
	"github.com/jmvdr-iscte/TradingBotCli/enums"
)

// Message type is used when connecting with alpaca API and openAi API.
// It has every information needed to buy or sell a position.
type Message struct {
	Uid      uuid.UUID  `json:"uid"`
	Headline string     `json:"headline"`
	Symbols  []string   `json:"symbols"`
	Risk     enums.Risk `json:"risk"`
}
