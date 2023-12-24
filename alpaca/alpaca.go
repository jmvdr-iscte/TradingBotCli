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

func (client *AlpacaClient) TradeOrder(symbol string, qty int64, side enums.OrderAction) error {
	order_action, err := enums.ProcessOrderAction(side)
	if err != nil {
		return fmt.Errorf("wrong order action: %w", err)
	}

	if qty > 0 {
		adjSide := alpaca.Side(order_action)
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

func (client *AlpacaClient) getBuyingPower() (float64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.BuyingPower.InexactFloat64(), nil
}

func (client *AlpacaClient) GetCash() (float64, error) {
	account, err := client.tradeClient.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("get account %w", err)
	}
	return account.Cash.InexactFloat64(), nil
}
func (client *AlpacaClient) SellPosition(symbol string, response int, risk enums.Risk) error {
	buying_power, err := client.getBuyingPower()
	if err != nil {
		return fmt.Errorf("unable to get account: %w", err)
	}

	position, err := client.tradeClient.GetPosition(symbol)
	if err != nil && buying_power >= 2000.0 {

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
	buying_power, err := client.getBuyingPower()
	if err != nil {
		return 0, fmt.Errorf("error getting buying power: %w", err)
	}

	latest_quote, err := client.getLastQuote(symbol)
	if err != nil {
		fmt.Println("error getting last quote: %w", err)
	}

	if latest_quote == 0.0 {
		latest_quote = 1.0
	}

	if side == enums.Buy {
		return utils.BuyQuantity(int64(response), buying_power, latest_quote, risk), nil
	} else {
		return utils.SellQuantity(int64(response), buying_power, latest_quote, risk), nil
	}
}

// func (client *AlpacaClient) getPercentChanges() error {
// 	symbols := make([]string, len(alp.allStocks))
// 	for i, stock := range algo.allStocks {
// 		symbols[i] = stock.name
// 	}

// 	// 20 minute percent changes
// 	end := time.Now()
// 	start := end.Add(-20 * time.Minute)
// 	feed := "iex"

// 	multiBars, err := client.dataClient.GetMultiBars(symbols, marketdata.GetBarsRequest{
// 		TimeFrame: marketdata.OneMin,
// 		Start:     start,
// 		End:       end,
// 		Feed:      feed,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("get multi bars: %w", err)
// 	}

// 	for i, symbol := range symbols {
// 		bars := multiBars[symbol]
// 		if len(bars) != 0 {
// 			percentChange := (bars[len(bars)-1].Close - bars[0].Open) / bars[0].Open
// 			algo.allStocks[i].pc = float64(percentChange)
// 		}
// 	}

//	return nil
//
// //	}
func (client *AlpacaClient) stopLoss(order_id string) error {

	fmt.Printf("order_id %s", order_id)
	order, err := client.tradeClient.GetOrder(order_id)
	if err != nil {
		return fmt.Errorf("order has not been filled, %w", err)
	}
	stop_loss_side := alpaca.Buy
	if order.Side == alpaca.Buy {
		stop_loss_side = alpaca.Sell
	} else if order.Side == alpaca.Sell {
		stop_loss_side = alpaca.Buy
	}
	fmt.Printf("THis is the order %v\n", order.FilledQty)
	if order.FilledAvgPrice == nil {
		return fmt.Errorf("FilledAvgPrice is nil")
	}
	stop_price := order.FilledAvgPrice.Mul(decimal.NewFromFloat(0.90).Abs()).Round(2)
	_, err = client.tradeClient.PlaceOrder(alpaca.PlaceOrderRequest{
		Symbol:      order.Symbol,
		Qty:         order.Qty,
		Side:        stop_loss_side,
		Type:        "stop",
		StopPrice:   &stop_price,
		TimeInForce: "day",
	})
	if err != nil {
		return fmt.Errorf("unable to set a stop loss: %w", err)
	}
	fmt.Println("managed to set stop loss order")

	return nil
}
