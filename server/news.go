package server

import (
	"sync"

	"github.com/jmvdr-iscte/TradingBot/models"
	"github.com/jmvdr-iscte/TradingBot/worker"
	"golang.org/x/net/websocket"
)

type NewsServer struct {
	Conns            map[*websocket.Conn]bool
	Mu               sync.Mutex
	shutdownCh       chan struct{}
	Options          models.Options
	Task_distributor worker.TaskDistributor
}

func NewServer(task_distributor worker.TaskDistributor, options *models.Options) *NewsServer {
	server := &NewsServer{
		Conns:            make(map[*websocket.Conn]bool),
		Mu:               sync.Mutex{},
		shutdownCh:       make(chan struct{}),
		Task_distributor: task_distributor,
		Options:          *options,
	}

	go func() {
		<-server.shutdownCh
		server.Shutdown()
	}()

	return server
}

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
