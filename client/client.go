package client

import (
	"fmt"
	"io"

	"github.com/jmvdr-iscte/TradingBot/alpaca"
	handler "github.com/jmvdr-iscte/TradingBot/handlers"
	"github.com/jmvdr-iscte/TradingBot/initialize"
	news "github.com/jmvdr-iscte/TradingBot/server"

	"golang.org/x/net/websocket"
)

const NewsURL = "wss://stream.data.alpaca.markets/v1beta1/news"

func ConnectToWebSocket(s *news.NewsServer) error {
	cfg := initialize.LoadAlpaca()
	alpaca_client := alpaca.LoadClient()
	serverURL := NewsURL
	wsConfig, err := websocket.NewConfig(serverURL, cfg.Url)
	fmt.Println("trying to connect to socket")

	if err != nil {
		return fmt.Errorf("error when connecting to the websocket: %w", err)
	}

	// Establish a WebSocket connection
	ws, err := websocket.DialConfig(wsConfig)

	if err != nil {
		return fmt.Errorf("error dialing configs: %w", err)
	}

	isMarketOpen, err := alpaca_client.IsMarketOpen()
	if err != nil {
		fmt.Println("Error_ ", err)
	}

	if !isMarketOpen {
		err = ws.Close()
		if err != nil {
			fmt.Println("Error Closing the websocket", err)
			return nil
		}
		s.Mu.Lock()
		delete(s.Conns, ws)
		s.Mu.Unlock()
		return nil
	}

	authMsg := map[string]interface{}{
		"action": "auth",
		"key":    cfg.ID,
		"secret": cfg.Secret,
	}

	if err := websocket.JSON.Send(ws, authMsg); err != nil {
		return fmt.Errorf("message error authentication: %w", err)

	}

	subscribeMsg := map[string]interface{}{
		"action": "subscribe",
		"news":   []string{"*"},
	}

	if err := websocket.JSON.Send(ws, subscribeMsg); err != nil {
		return fmt.Errorf("message error subsription: %w", err)
	}

	var response []map[string]interface{}
	if err := websocket.JSON.Receive(ws, &response); err != nil {
		if err == io.EOF {
			return fmt.Errorf("error connection closed by client: %w", err)
		} else {
			return fmt.Errorf("error in the websocket %w", err)
		}
	}
	handler.HandleWS(ws, s)
	return nil
}
