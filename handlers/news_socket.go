package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jmvdr-iscte/TradingBotCli/models"
	"github.com/jmvdr-iscte/TradingBotCli/server"
	"github.com/jmvdr-iscte/TradingBotCli/worker"
	"golang.org/x/net/websocket"
)

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

	if err := readData(ws, s, options); err == io.EOF {
		s.Mu.Lock()

		delete(s.Conns, ws)
		s.Mu.Unlock()
	} else {
		fmt.Println("Error handling the websocket: ", err)
	}
}

func readData(ws *websocket.Conn, s *server.NewsServer, opts []asynq.Option) error {

	var (
		message_buffer []byte
		buf            = make([]byte, 8192)
	)

	for {
		n, err := ws.Read(buf)

		if err != nil {
			if err == io.EOF { // significa que a conecção fechou do outro lado
				break

			}
			fmt.Println("Read error: ", err)
			return err
		}
		current_cash, err := s.AlpacaClient.GetCash()
		if err != nil {
			fmt.Println("unable to get current cash: %w", err)
			return err
		}

		if current_cash >= s.Options.StartingValue+s.Options.Gain {
			result := current_cash - s.Options.StartingValue
			fmt.Printf("you gained %f\n:", result)
			err = s.AlpacaClient.ClearOrders()
			if err != nil {
				return err
			}
			return io.EOF
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
			if len(message.Headline) == 1 {
				message.Uid = uuid.New()
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
	return nil
}
