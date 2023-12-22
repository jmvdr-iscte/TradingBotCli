package enums

import "fmt"

type OrderAction int

const (
	Buy OrderAction = iota + 1
	Sell
)

func (o OrderAction) String() string {
	return [...]string{"Buy", "Sell"}[o-1]
}

func (o OrderAction) EnumIndex() int {
	return int(o)
}

func ProcessOrderAction(o OrderAction) (string, error) {
	switch o {
	case Buy:
		return "buy", nil
	case Sell:
		return "sell", nil
	default:
		return "", fmt.Errorf("invalid value for order action")
	}
}
