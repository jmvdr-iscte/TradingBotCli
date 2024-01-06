// Package client serves as the client who connects to the news
// server.
package client

import (
	"fmt"
	"io"

	"github.com/jmvdr-iscte/TradingBotCli/handlers"
	"github.com/jmvdr-iscte/TradingBotCli/initialize"
	news "github.com/jmvdr-iscte/TradingBotCli/server"

	"golang.org/x/net/websocket"
)

// NewsURL The news socket url.
const NewsURL = "wss://stream.data.alpaca.markets/v1beta1/news"

// ConnectToWebSocket makes the initial connection to the Alpaca news socket,
// it returns an error if anything goes wrong in the connection to the socket, or
// it the auxiliary functions of the api.
func ConnectToWebSocket(s *news.NewsServer) error {
	cfg := initialize.LoadAlpaca()
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

	isMarketOpen, err := s.AlpacaClient.IsMarketOpen()
	if err != nil {
		return fmt.Errorf("unable to check the market conditions %w", err)
	}

	haveTrades, err := s.AlpacaClient.HaveTrades()
	if err != nil {
		return fmt.Errorf("unable to check the current trades %w", err)
	}

	if !isMarketOpen || !haveTrades {
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
	handlers.HandleWS(ws, s)
	return nil
}
