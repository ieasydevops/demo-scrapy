.PHONY: build docker-build docker-run docker-stop docker-push

IMAGE_NAME=demo-scrapy
IMAGE_TAG=latest
REGISTRY=

build:
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/server ./cmd/server/main.go

run: build
	./bin/server

run-dev:
	go run -ldflags="-s -w" ./cmd/server/main.go

docker-build:
	docker-compose build

docker-build-backend:
	docker build -t $(IMAGE_NAME)-backend:$(IMAGE_TAG) .

docker-build-frontend:
	docker build -t $(IMAGE_NAME)-frontend:$(IMAGE_TAG) ./frontend

docker-build-lowmem:
	docker build --memory=2g --memory-swap=2g -t $(IMAGE_NAME)-backend:$(IMAGE_TAG) .

docker-build-cn:
	docker-compose -f docker-compose.cn.yml build

docker-build-cn-backend:
	docker build -f Dockerfile.cn -t $(IMAGE_NAME)-backend:$(IMAGE_TAG) .

docker-build-cn-frontend:
	docker build -f frontend/Dockerfile.cn -t $(IMAGE_NAME)-frontend:$(IMAGE_TAG) ./frontend

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-push:
ifdef REGISTRY
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
else
	@echo "请设置 REGISTRY 变量，例如: make docker-push REGISTRY=your-registry.com"
endif
