// Package alpaca provides auxiliary functions to connect with
// the Alpaca API.
package alpaca

import (
	"fmt"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/jmvdr-iscte/TradingBotCli/enums"
	"github.com/jmvdr-iscte/TradingBotCli/initialize"
	"github.com/jmvdr-iscte/TradingBotCli/utils"
	"github.com/shopspring/decimal"
)

// A AlpacaClient serves as the client who interacts with the Alpaca API,
// it can interact via a tradeClient and a dataClient.
type AlpacaClient struct {
	tradeClient *alpaca.Client
	dataClient  *marketdata.Client
}

// LoadClient returns a pointer to the AlpacaClient
// The AlpacaClient is made up of the tradeClient and a dataClient.
func LoadClient() *AlpacaClient {

	configs := initialize.LoadAlpaca()

	return &AlpacaClient{
		tradeClient: alpaca.NewClient(alpaca.ClientOpts{
			APIKey:    configs.ID,
			APISecret: configs.Secret,
			BaseURL:   configs.Url,
		}),

		dataClient: marketdata.NewClient(marketdata.ClientOpts{
			APIKey:    configs.ID,
			APISecret: configs.Secret,
		}),
	}
}

// func (client *AlpacaClient) ClearOrders() error {
// 	orders, err := client.tradeClient.GetOrders(alpaca.GetOrdersRequest{
// 		Status: "open",
// 		Until:  time.Now(),
// 		Limit:  100,
// 	})

// 	if err != nil {
// 		return err
// 	}
// 	for _, order := range orders {
// 		if err := client.tradeClient.CancelOrder(order.ID); err != nil {
// 			return err
// 		}
// 	}
// 	fmt.Printf("%d order(s) cancelled\n", len(orders))
// 	return nil
// }

// ClosePositions returns an error if we were not able to connect to the API,
// otherwise it returns nil.
func (client *AlpacaClient) ClosePositions() error {
	req := alpaca.CloseAllPositionsRequest{
		CancelOrders: true,
	}
	_, err := client.tradeClient.CloseAllPositions(req)
	if err != nil {
		return fmt.Errorf("unable to close all positions %w", err)
	}
	return nil
}

// TradeOrder returns an error if it was not able to send an order to the API.
// It can make sorts, regular orders, stop loss orders, etc..., depending on the
// context that is called.
func (client *AlpacaClient) TradeOrder(symbol string, qty int64, side enums.OrderAction) error {
	orderAction, err := enums.ProcessOrderAction(side)
	if err != nil {
		return fmt.Errorf("wrong order action: %w", err)
	}

	if qty > 0 {
		adjSide := alpaca.Side(orderAction)
		decimalQty := decimal.NewFromInt((qty))
		order, err := client.tradeClient.PlaceOrder(alpaca.PlaceOrderRequest{
			Symbol:      symbol,
			Qty:         &decimalQty,
			Side:        adjSide,
			Type:        "market",
			TimeInForce: "day",
		})
		if err == nil {
			fmt.Printf("Market order of | %d %s %s | completed\n", qty, symbol, side)
			time.Sleep(3 * time.Second)
			err = client.stopLoss(order.ID)
			if err != nil {
				fmt.Println("Unable to set up a trailing stop order: %w", err)
			}
		} else {
			fmt.Printf("Order of | %d %s %s | did not go through: %s\n", qty, symbol, side, err)
		}
		return nil
	}
	fmt.Printf("Quantity is <= 0, order of | %d %s %s | not sent\n", qty, symbol, side)
	return nil
}

// IsMarketOpen returns true and nil if the market is currently open,
// otherwise it returns false and nil. If there is a problem getting the time it returns
// false and an error to go with it.
func (client *AlpacaClient) IsMarketOpen() (bool, error) {
	clock, err := client.tradeClient.GetClock()
	if err != nil {
		return false, fmt.Errorf("get clock: %w", err)
	}

	if clock.IsOpen {
		return true, nil
	}

	timeToOpen := int(clock.NextOpen.Sub(clock.Timestamp).Minutes())
	switch {
	case timeToOpen < 60:
		fmt.Printf("%d minutes until next market open\n", timeToOpen)

	case timeToOpen > 60 && timeToOpen < 1440:
		hoursToOpen := timeToOpen / 60
		fmt.Printf("%d hours until next market open\n", hoursToOpen)

	case timeToOpen > 1440:
		daysToOpen := timeToOpen / 1440
		fmt.Printf("%d days until next market open\n", daysToOpen)
	}

	return false, nil
}

// HaveTrades returns true and nil if the user is not considered a
// PTD with less than 25.000$ in the account and still has trades for the rest of the week.
// If he is a PTD with more than 25.000$ in the account this wil always return true and nil.
// In any other case this function returns false. And in case of an error, this function returns
// false and error.
func (client *AlpacaClient) HaveTrades() (bool, error) {
	dayTradingCount, err := client.GetDayTradingCount()
	if err != nil {
		return false, fmt.Errorf("get day trading count %w", err)
	}
	equity, err := client.GetEquity()
	if err != nil {
		return false, fmt.Errorf("get equity %w", err)
	}

	if dayTradingCount >= 3 && equity < 25000 {
		fmt.Println("warning: please do not make any more trades this week")
		return false, nil
	}
	return true, nil
}

// getBuyingPower returns a float64 of the user's buying power,
// if anything goes wrong it returns 0 and an error.
func (client *AlpacaClient) getBuyingPower() (float64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.BuyingPower.InexactFloat64(), nil
}

// GetDayTradingByingPower returns a float64 of the user's  day trading buying power,
// if anything goes wrong it returns 0 and an error.
func (client *AlpacaClient) GetDayTradingBuyingPower() (float64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.DaytradingBuyingPower.InexactFloat64(), nil
}

// GetEquity returns a float64 of the user's equity,
// if anything goes wrong it returns 0 and an error.
func (client *AlpacaClient) GetEquity() (float64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.Equity.InexactFloat64(), nil
}

// GetCash returns a float64 of the user's cash,
// if anything goes wrong it returns 0 and an error.
func (client *AlpacaClient) GetCash() (float64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.Cash.InexactFloat64(), nil
}

// GetDayTradingCount returns the amount of trades the user has done in a week.
// This is important so we can prevent accounts from getting tagged with PTD.
func (client *AlpacaClient) GetDayTradingCount() (int64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.DaytradeCount, nil
}

// isBlocked returns true if the account is blocked in Alpaca
// otherwise it returns false. And it returns an error if an error is found.
func (client *AlpacaClient) IsBlocked() (bool, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return true, fmt.Errorf("get account %w", err)
	}
	return account.AccountBlocked, nil
}

// getLastQuote returns the latest active quote of a stock( if you have unlimited subscription
// please change it to marketdata.SIP). If it is unable to get the latest quote it returns a
// default price of 20.
func (client *AlpacaClient) getLastQuote(symbol string, side enums.OrderAction) (float64, error) {
	req := marketdata.GetSnapshotRequest{
		Feed:     marketdata.IEX,
		Currency: "USD",
	}

	gmeSnapshot, err := client.dataClient.GetSnapshot(symbol, req)
	if err != nil {
		return 20.0, fmt.Errorf("get snapshot: %w", err)
	}

	if gmeSnapshot == nil || gmeSnapshot.LatestQuote == nil {
		return 20.0, fmt.Errorf("snapshot or latest quote is nil")
	}
	if enums.Buy == side {
		return gmeSnapshot.LatestQuote.AskPrice, nil
	}
	return gmeSnapshot.LatestQuote.BidPrice, nil
}

// SellPosition is a function that takes care of every variable and property regarding
// a sell or a short. It returns nil if a short or a sell was sucessfully placed, and an error
// otherwise.
func (client *AlpacaClient) SellPosition(symbol string, response int, risk enums.Risk) error {
	buyingPower, err := client.getBuyingPower()
	if err != nil {
		return fmt.Errorf("unable to get account: %w", err)
	}

	position, err := client.tradeClient.GetPosition(symbol)
	if err != nil && buyingPower >= 2000.0 {

		qty, err := client.GetQuantity(response, symbol, enums.Sell, risk)

		if err != nil {
			return fmt.Errorf("unable to get quantity %w", err)
		}

		client.TradeOrder(symbol, qty, enums.Sell)
		return nil
	}

	if position.QtyAvailable.IntPart() > 0 {
		qty := position.Qty.Abs()

		err := client.TradeOrder(symbol, qty.IntPart(), enums.Sell)
		if err != nil {
			return fmt.Errorf("error placing order %w", err)
		}
		return nil
	}
	return nil
}

// BuyPosition is a function that takes care of every variable and property regarding
// a buy. It returns nil if a buywas sucessfully placed, and an error
// otherwise.
func (client *AlpacaClient) BuyPosition(response int, symbol string, risk enums.Risk) error {
	buy_quantity, err := client.GetQuantity(response, symbol, enums.Buy, risk)
	if err != nil {
		return fmt.Errorf("error setting buy quantity error ")
	}
	if client.TradeOrder(symbol, buy_quantity, enums.Buy) != nil {
		return fmt.Errorf("error making the trade: %w", err)
	}
	return nil
}

// GetQuantity returns the quantity in int64 of the stock to sell or buy.
// The quantity varies according to the action(side), the risk selected and the sentiment analysis.
// If there is a problem getting the quote or the buying power it will return 0 and an error.
func (client *AlpacaClient) GetQuantity(response int, symbol string, side enums.OrderAction, risk enums.Risk) (int64, error) {
	buyingPower, err := client.getBuyingPower()
	if err != nil {
		return 0, fmt.Errorf("error getting buying power: %w", err)
	}

	latestQuote, err := client.getLastQuote(symbol, side)
	if err != nil {
		fmt.Println("error getting last quote: %w", err)
	}

	if latestQuote == 0.0 {
		latestQuote = 1.0
	}

	if risk == enums.Medium || risk == enums.Low || risk == enums.High {
		if side == enums.Buy {
			return utils.BuyPDTQuantity(int64(response), buyingPower, latestQuote, risk), nil
		} else {
			return utils.SellPDTQuantity(int64(response), buyingPower, latestQuote, risk), nil
		}
	} else if risk == enums.Safe || risk == enums.Power {
		return utils.CalculateQuantity(buyingPower, latestQuote, risk), nil
	}
	return 0, nil
}

// stopLoss returns an error if a stop loss was not sucessfully set up.
// The stop loss currently is set at 10% of the original value.
// If everything goes well it returns nil.
func (client *AlpacaClient) stopLoss(orderId string) error {

	fmt.Printf("orderId %s", orderId)
	order, err := client.tradeClient.GetOrder(orderId)
	if err != nil {
		return fmt.Errorf("order has not been filled, %w", err)
	}
	stopLossSide := alpaca.Buy
	if order.Side == alpaca.Buy {
		stopLossSide = alpaca.Sell
	} else if order.Side == alpaca.Sell {
		stopLossSide = alpaca.Buy
	}

	if order.FilledAvgPrice == nil {
		return fmt.Errorf("FilledAvgPrice is nil")
	}
	stop_price := order.FilledAvgPrice.Mul(decimal.NewFromFloat(0.90).Abs()).Round(2)
	_, err = client.tradeClient.PlaceOrder(alpaca.PlaceOrderRequest{
		Symbol:      order.Symbol,
		Qty:         order.Qty,
		Side:        stopLossSide,
		Type:        "stop",
		StopPrice:   &stop_price,
		TimeInForce: "day",
	})
	if err != nil {
		return fmt.Errorf("unable to set a stop loss: %w", err)
	}
	fmt.Println("stop loss order set")
	return nil
}

// CanClosePositions returns true if there are 15 minutes left on the market hours
// and closes the positions. If there are more than 15 min it returns false.
// If there is a problem getting any data it returns false and an error.
func (client *AlpacaClient) CanClosePositions() (bool, error) {
	clock, err := client.tradeClient.GetClock()
	if err != nil {
		return false, fmt.Errorf("get clock: %w", err)
	}
	nextClose := clock.NextClose
	closeTime := nextClose.Add(-15 * time.Minute)
	if clock.IsOpen && time.Now().After(closeTime) {
		err := client.ClosePositions()
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}
