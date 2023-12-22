package enums

import (
	"fmt"
)

type Risk int

const (
	Low Risk = iota + 1
	Medium
	High
)

func (r Risk) String() string {
	return [...]string{"low", "medium", "high"}[r-1]
}

func (r Risk) EnumIndex() int {
	return int(r)
}

func ProcessRisk(r Risk) (string, error) {
	switch r {
	case Low:
		return "low", nil
	case Medium:
		return "medium", nil
	case High:
		return "high", nil
	default:
		return "", fmt.Errorf("invalid value for filter")
	}
}
