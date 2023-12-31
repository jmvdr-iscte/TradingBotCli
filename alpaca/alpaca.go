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

type AlpacaClient struct {
	tradeClient *alpaca.Client
	dataClient  *marketdata.Client
}

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

func (client *AlpacaClient) ClearOrders() error {
	orders, err := client.tradeClient.GetOrders(alpaca.GetOrdersRequest{
		Status: "open",
		Until:  time.Now(),
		Limit:  100,
	})

	if err != nil {
		return err
	}
	for _, order := range orders {
		if err := client.tradeClient.CancelOrder(order.ID); err != nil {
			return err
		}
	}
	fmt.Printf("%d order(s) cancelled\n", len(orders))
	return nil
}

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

func (client *AlpacaClient) getBuyingPower() (float64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.BuyingPower.InexactFloat64(), nil
}

func (client *AlpacaClient) GetDayTradingBuyingPower() (float64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.DaytradingBuyingPower.InexactFloat64(), nil
}

func (client *AlpacaClient) GetEquity() (float64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.Equity.InexactFloat64(), nil
}

func (client *AlpacaClient) GetDayTradingCount() (int64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.DaytradeCount, nil
}

func (client *AlpacaClient) IsBlocked() (bool, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return true, fmt.Errorf("get account %w", err)
	}
	return account.AccountBlocked, nil
}

func (client *AlpacaClient) GetCash() (float64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.Cash.InexactFloat64(), nil
}

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

func (client *AlpacaClient) getLastQuote(symbol string) (float64, error) {
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

	return gmeSnapshot.LatestQuote.AskPrice, nil
}

func (client *AlpacaClient) GetQuantity(response int, symbol string, side enums.OrderAction, risk enums.Risk) (int64, error) {
	buyingPower, err := client.getBuyingPower()
	if err != nil {
		return 0, fmt.Errorf("error getting buying power: %w", err)
	}

	latestQuote, err := client.getLastQuote(symbol)
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
