// Package server is used to contain the app server.
package server

import (
	"log"
	"sync"

	"github.com/jmvdr-iscte/TradingBotCli/alpaca"
	"github.com/jmvdr-iscte/TradingBotCli/models"
	"github.com/jmvdr-iscte/TradingBotCli/worker"
	"golang.org/x/net/websocket"
)

// NewsServer has all the fields necessary for the program to run.
type NewsServer struct {
	Conns            map[*websocket.Conn]bool
	Mu               sync.Mutex
	shutdownCh       chan struct{}
	Options          models.Options
	Task_distributor worker.TaskDistributor
	AlpacaClient     *alpaca.AlpacaClient
}

// NewsServer instanciates a pointer of a new server with the correct run options and task distributors.
func NewServer(task_distributor worker.TaskDistributor, options *models.Options) *NewsServer {
	alpaca_client := *alpaca.LoadClient()
	var err error
	options.StartingValue, err = alpaca_client.GetEquity()
	if err != nil {
		log.Fatalf("Failed to get equity: %v", err)
		return nil
	}

	server := &NewsServer{
		Conns:            make(map[*websocket.Conn]bool),
		Mu:               sync.Mutex{},
		shutdownCh:       make(chan struct{}),
		Task_distributor: task_distributor,
		Options:          *options,
		AlpacaClient:     &alpaca_client,
	}

	go func() {
		<-server.shutdownCh
		server.Shutdown()
	}()

	return server
}

// Shutdown ends the server procedure and closes it's websockets.
func (s *NewsServer) Shutdown() {

	s.Mu.Lock()
	defer s.Mu.Unlock()

	if s.shutdownCh == nil {
		return
	}

	for ws := range s.Conns {
		ws.Close()
		delete(s.Conns, ws)
	}

	close(s.shutdownCh)
	s.shutdownCh = nil
}
