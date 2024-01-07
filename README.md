# TradingBotCli

TradingBotCli is a command-line tool written in Go that automates trading decisions based on various factors such as news, market data, and risk analysis. It uses the Alpaca Trade API to interact with the stock market.

## Features

- Automated trading decisions based on news, market data, and risk analysis.
- Interacts with the Alpaca Trade API to execute trades.
- Utilizes Asynq for background job processing.
- Leverages Redis for caching and storing data.
- Uses OpenAI for AI-based decision making.
- Executes stop-loss to mitigate potential losses.
- Gives the user capacity to choose the risk.

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

```bash
git clone https://github.com/jmvdr-iscte/TradingBotCli.git
cd TradingBotCli
make build
```


## Usage

After installation, you can run the trading bot with the following command:

`make start`

After starting the program, you will have to pick your prefered risk level:

`Safe, Low, Medium, High, Power`

These setting will influence not only the amount of trades that you can do, but also the
value of said traids. All the values are influenced by the AI sentiment analysis, that range
from 0 to 100.

The `Safe` and `Power` will only make trades above 95 and below 5 of sentiment analysis. 
Which will make for fewer trades, but with higher gain possibility.
They differ in amount of money invested in each trade.
I personally recomend this configuration if you have less than 25.000$ in your account, so you wont be affected 
by the PTD rule.

The `Low`, `Medium`, and `High` make trades above 75 and below 25 of sentiment analysis.
These configurations allow for more trades per day and differ in amount of money invested in each 
trade.
I personaly recomend these configurations if you have more than 25.000$ in your account, because at that point you can safely day trade.

After you selected the risk you can pick the amount of money you want to gain per day. The bot will stop 
as soon as it reaches that limit. but if you want it to run until the end of the day select a ridiculos amount
of earning like 1.000.000.0

Also please keep your pc always on so to not kill the connection.
