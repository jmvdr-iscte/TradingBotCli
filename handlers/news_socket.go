// Package handlers serves as the handler for connections
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jmvdr-iscte/TradingBotCli/models"
	"github.com/jmvdr-iscte/TradingBotCli/server"
	"github.com/jmvdr-iscte/TradingBotCli/worker"
	"golang.org/x/net/websocket"
)

// HandleWs handles the websocket connection, as it sends asynq
// tasks to redis, in order to deal with the buying and selling
// opperations.
func HandleWS(ws *websocket.Conn, s *server.NewsServer) {
	fmt.Println("new incoming connection from client: ", ws.RemoteAddr())
	options := []asynq.Option{
		asynq.ProcessIn(1 * time.Second),
		asynq.Queue(worker.QueueCritical),
		asynq.MaxRetry(1),
	}
	s.Mu.Lock()
	s.Conns[ws] = true
	s.Mu.Unlock()

	stopChan := make(chan bool)
	go monitorData(s, stopChan)
	if err := readData(ws, s, options, stopChan); err == io.EOF {
		s.Mu.Lock()
		delete(s.Conns, ws)
		s.Mu.Unlock()
		return

	} else if err != nil {
		fmt.Println("Error handling the websocket: %w", err)
		return
	}
	fmt.Println("websocket sucessfully closed")
	s.Mu.Lock()
	delete(s.Conns, ws)
	s.Mu.Unlock()
}

// readData returns an error if anything goes wrong with the connectio. It reads the data and
// sends it to redis.
func readData(ws *websocket.Conn, s *server.NewsServer, opts []asynq.Option, stopChan chan bool) error {

	var (
		message_buffer []byte
		buf            = make([]byte, 8192)
	)

	for {
		select {
		case <-stopChan:
			return nil
		default:
			ws.SetReadDeadline(time.Now().Add(2 * time.Second)) // Set a 2 second timeout
			n, err := ws.Read(buf)

			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				} else if err == io.EOF {
					break
				}
				fmt.Println("Read error: ", err)
				return err
			}

			message_buffer = append(message_buffer, buf[:n]...)
			var messages []models.Message

			if err := json.Unmarshal(message_buffer, &messages); err != nil {
				if err == io.ErrUnexpectedEOF {
					continue
				}
				fmt.Println("Error when getting the news: ", err)
				continue
			}

			for _, message := range messages {
				if len(message.Headline) != 0 {
					message.Risk = s.Options.Risk
					err = s.Task_distributor.DistributeTaskProcessOrder(context.Background(), &message, opts...)
					if err != nil {
						return fmt.Errorf("unable to distribute task %w", err)
					}
				}
			}
			fmt.Println("Received message: ", string(message_buffer))
			message_buffer = nil
		}
	}
}

// monitorData returns an error if it was an error cpnnecting to the api.
// It monitors the whole system in order to be able to correctly close
// positions and shutdown the system.
func monitorData(s *server.NewsServer, stopChan chan<- bool) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {

		haveTrades, err := s.AlpacaClient.HaveTrades()
		if err != nil {
			stopChan <- true
			return err
		}

		current_equity, err := s.AlpacaClient.GetEquity()
		if err != nil {
			stopChan <- true
			return err
		}

		can_close_positions, err := s.AlpacaClient.CanClosePositions()
		if err != nil {
			stopChan <- true
			return err
		}

		fmt.Printf("current equity %f\n", current_equity)
		fmt.Printf("possible gainz %f\n", s.Options.StartingValue+s.Options.Gain)
		if current_equity >= s.Options.StartingValue+s.Options.Gain {
			result := current_equity - s.Options.StartingValue
			fmt.Printf("you gained %f\n:", result)
			err = s.AlpacaClient.ClosePositions()
			if err != nil {
				stopChan <- true
				return err
			}
			stopChan <- true
			return nil
		}

		if !haveTrades {
			stopChan <- true
			return nil
		}

		if can_close_positions {
			stopChan <- true
			fmt.Println("15 minutes left to close \n Closing all positions")
			return nil
		}
	}
	return nil
}
