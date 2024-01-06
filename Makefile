start:
	sudo docker-compose up -d redis
	sudo docker-compose run trading_botcli

build:
	sudo docker-compose build

prune:
	sudo docker system prune