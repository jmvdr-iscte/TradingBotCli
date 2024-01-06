# TradingBotCli

TradingBotCli is a command-line tool written in Go that automates trading decisions based on various factors such as news, market data, and risk analysis. It uses the Alpaca Trade API to interact with the stock market.

## Features

- Automated trading decisions based on news, market data, and risk analysis.
- Interacts with the Alpaca Trade API to execute trades.
- Utilizes Asynq for background job processing.
- Leverages Redis for caching and storing data.
- Uses Go-OpenAI for AI-based decision making.

## Directory Structure

- `alpaca/`: Contains a Go file (`alpaca.go`) related to interacting with the Alpaca API.
- `client/`: Contains a Go file (`client.go`) related to the client functionality of the trading bot.
- `enums/`: Contains a Go file (`risk.go`) defining risk-related enums.
- `handlers/`: Contains a Go file (`news_socket.go`) related to handling news socket functionality.
- `initialize/`: Contains Go files (`alpaca.go`, `openai.go`, `redis_ops.go`) related to initializing various components of the trading bot.
- `models/`: Contains Go files (`message.go`, `options.go`) defining various models used in the project.
- `open_ai/`: Contains a Go file (`open_ai.go`) related to interacting with the OpenAI API.
- `server/`: Contains a Go file (`news.go`) related to the server functionality of the trading bot.
- `utils/`: Contains Go files (`ptd-quantity.go`, `quantity.go`) defining utility functions for quantity calculations.
- `worker/`: Contains Go files (`distributor.go`, `processor.go`, `task_process_order.go`) related to the worker functionality of the trading bot.

## Installation

To install TradingBotCli, clone the repository and run the following commands:

bash git clone https://github.com/yourusername/TradingBotCli.git
```bash
git clone https://github.com/yourusername/TradingBotCli.git
cd TradingBotCli
make install
```


## Usage

After installation, you can run the trading bot with the following command:

bash ./TradingBotCli
```bash
./TradingBotCli
```


## Contributing

Contributions are welcome! Please feel free to submit a pull request.

## License

This project is licensed under the MIT License.