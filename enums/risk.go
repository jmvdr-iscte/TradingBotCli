// Package enums serves as the enums class.
package enums

import (
	"fmt"
)

// Risk is the risk enum type.
// Defines the risk of the bot.
type Risk int

const (
	Safe Risk = iota + 1
	Low
	Medium
	High
	Power
)

// String returns the string value of the respective enum.
func (r Risk) String() string {
	return [...]string{"safe", "low", "medium", "high", "power"}[r-1]
}

// EnumIndex returns the inndex of the respective enum.
func (r Risk) EnumIndex() int {
	return int(r)
}

// ProcessRisk turns a enum into a string.
func ProcessRisk(r Risk) (string, error) {
	switch r {
	case Safe:
		return "safe", nil
	case Low:
		return "low", nil
	case Medium:
		return "medium", nil
	case High:
		return "high", nil
	case Power:
		return "power", nil
	default:
		return "", fmt.Errorf("invalid value for filter")
	}
}
