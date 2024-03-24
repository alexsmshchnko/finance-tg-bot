include .env

.PHONY: dc run cover gen

clean:
	rm -f app

build: clean
	go build -o app cmd/finance_bot.go

dc:
	docker-compose -f docker-compose.local.yml up -d --remove-orphans --build

gen:
	mockgen -source=pkg/repository/repository.go -destination=pkg/repository/mocks/mock_repository.go 
	mockgen -source=pkg/repository/reports.go -destination=pkg/repository/mocks/mock_reports.go 
	mockgen -source=pkg/repository/users.go -destination=pkg/repository/mocks/mock_users.go 

cover: gen
	go test ./internal/usecase/repo/... -short -count=10 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

webhook_info:
	curl --request POST --url "https://api.telegram.org/bot$(TELEGRAM_APITOKEN)/getWebhookInfo"

webhook_delete:
	curl --request POST --url "https://api.telegram.org/bot$(TELEGRAM_APITOKEN)/deleteWebhook"

webhook_create: webhook_delete
	curl --request POST --url "https://api.telegram.org/bot$(TELEGRAM_APITOKEN)/setWebhook" --header 'content-type: application/json' --data "{\"url\": \"$(SERVERLESS_APIGW_URL)\"}"

build_yc_test:
	docker build -f Dockerfile.yctest -t cr.yandex/$(YC_IMAGE_REGISTRY_ID)/$(SERVERLESS_CONTAINER_NAME) .

push_yc_test: build_yc_test
	docker push cr.yandex/$(YC_IMAGE_REGISTRY_ID)/$(SERVERLESS_CONTAINER_NAME)