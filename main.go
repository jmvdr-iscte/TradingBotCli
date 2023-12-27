package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jmvdr-iscte/TradingBotCli/client"
	"github.com/jmvdr-iscte/TradingBotCli/enums"
	"github.com/jmvdr-iscte/TradingBotCli/initialize"
	"github.com/jmvdr-iscte/TradingBotCli/models"
	news "github.com/jmvdr-iscte/TradingBotCli/server"
	"github.com/jmvdr-iscte/TradingBotCli/worker"
	"github.com/rs/zerolog/log"
)

func main() {

	var risk_value string
	var risk enums.Risk
	var stop_gain float64
	var err error

	for {
		fmt.Println("Please select your preferred risk: Low, Medium, High")
		fmt.Scanln(&risk_value)
		risk_value = strings.ToLower(strings.TrimSpace(risk_value))
		risk, err = coerceToRisk(risk_value)
		if err != nil {
			fmt.Println("Invalid input. Please enter Low, Medium, or High.")
		} else {
			break
		}
	}
	fmt.Println("You selected:", risk.String())

	for {
		fmt.Println("Please select your expected gain today")
		_, err := fmt.Scanln(&stop_gain)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
		} else {
			break
		}
	}
	fmt.Println("You selected:", stop_gain)

	options := models.Options{
		Risk: risk,
		Gain: stop_gain,
	}

	redis_config := initialize.LoadRedisConfigs()

	redisOpt := asynq.RedisClientOpt{
		Addr:     redis_config.Address,
		Password: redis_config.Password,
	}
	go banana()

	task_distributor := worker.NewRedisTaskDistributor(redisOpt)
	go runTaskProcessor(redisOpt) // tem de ser numa go routine pois tal como um servidor http, ele bloqueia se não tiver pedidos
	server := news.NewServer(task_distributor, &options)
	err = client.ConnectToWebSocket(server)
	if err != nil {
		fmt.Println(err)
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		server.Shutdown()

		os.Exit(0)
	}()

	select {}
}

func runTaskProcessor(redisOpt asynq.RedisClientOpt) {
	task_processor := worker.NewRedisTaskProcessor(redisOpt)
	log.Info().Msg("start task processor")
	err := task_processor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}

func coerceToRisk(risk_str string) (enums.Risk, error) {
	switch strings.ToLower(risk_str) {
	case "low":
		return enums.Low, nil
	case "medium":
		return enums.Medium, nil
	case "high":
		return enums.High, nil
	default:
		return 0, fmt.Errorf("invalid value for Risk: %s", risk_str)
	}

}

func banana() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("Banana")
	}
}
