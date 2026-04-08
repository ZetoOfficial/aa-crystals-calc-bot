IMAGE     := aa-crystals-calc-bot
CONTAINER := aa-crystals-calc-bot
ENV_FILE  := .env

.PHONY: help build run stop rm clean logs

help:
	@echo "Targets:"
	@echo "  make build   — собрать docker-образ $(IMAGE)"
	@echo "  make run     — снести старый контейнер+образ, собрать новый и запустить (главная команда)"
	@echo "  make stop    — остановить контейнер $(CONTAINER)"
	@echo "  make rm      — остановить и удалить контейнер"
	@echo "  make clean   — удалить контейнер и образ"
	@echo "  make logs    — follow логов работающего контейнера"

# Главный цикл деплоя: остановить и удалить старый контейнер,
# удалить старый образ, собрать новый и запустить.
run: clean build
	docker run -d \
		--name $(CONTAINER) \
		--restart unless-stopped \
		--env-file $(ENV_FILE) \
		$(IMAGE)

build:
	docker build -t $(IMAGE) .

stop:
	-docker stop $(CONTAINER) 2>/dev/null

rm: stop
	-docker rm $(CONTAINER) 2>/dev/null

clean: rm
	-docker rmi $(IMAGE) 2>/dev/null

logs:
	docker logs -f $(CONTAINER)
