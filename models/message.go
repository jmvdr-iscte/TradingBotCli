package models

import (
	"github.com/google/uuid"
	"github.com/jmvdr-iscte/TradingBotCli/enums"
)

type Message struct {
	Uid      uuid.UUID  `json:"uid"`
	Headline string     `json:"headline"`
	Symbols  []string   `json:"symbols"`
	Risk     enums.Risk `json:"risk"`
}
