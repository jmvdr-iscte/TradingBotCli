package enums

import (
	"fmt"
)

type Risk int

const (
	Safe Risk = iota + 1
	Low
	Medium
	High
	Power
)

func (r Risk) String() string {
	return [...]string{"safe", "low", "medium", "high", "power"}[r-1]
}

func (r Risk) EnumIndex() int {
	return int(r)
}

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
