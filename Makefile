include .env

.PHONY: dc run

clean:
	rm -f app

build: clean
	go build -o app cmd/finance_bot.go

dc:
	docker-compose -f docker-compose.local.yml up -d --remove-orphans --build

.PHONY: cover
cover:
	go test internal/usecase/repo/reports/reports_test.go internal/usecase/repo/reports/reports.go -short -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

.PHONY: gen
gen:
	mockgen -source=pkg/repository/repository.go \
	-destination=pkg/repository/mocks/mock_repository.go 
	mockgen -source=pkg/repository/reports.go \
	-destination=pkg/repository/mocks/mock_reports.go 

build_yc_test:
	docker build -f Dockerfile.yctest -t cr.yandex/$(YC_IMAGE_REGISTRY_ID)/$(SERVERLESS_CONTAINER_NAME) .

push_yc_test: build_yc_test
	docker push cr.yandex/$(YC_IMAGE_REGISTRY_ID)/$(SERVERLESS_CONTAINER_NAME)