VERSION=`git rev-parse HEAD`
BUILD=`date +%FT%T%z`
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

.PHONY: help
help: ## - Show help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build:	## - Build the smallest and secured golang docker image based on scratch
	@docker build -f Dockerfile -t smallest-secured-golang .

.PHONY: build-no-cache
build-no-cache:	## - Build the smallest and secured golang docker image based on scratch with no cache
	@docker build --no-cache -f Dockerfile -t smallest-secured-golang .

.PHONY: ls
ls: ## - List 'smallest-secured-golang' docker images
	@docker image ls smallest-secured-golang

.PHONY: push-to-azure
push-to-azure:	## - Push docker image to azurecr.io container registry
	@az acr login --name chemidy
	@docker push chemidy.azurecr.io/smallest-secured-golang-docker-image:$(VERSION)

.PHONY: push-to-gcp
push-to-gcp:	## - Push docker image to gcr.io container registry
	@gcloud auth application-default login
	@gcloud auth configure-docker
	@docker push gcr.io/chemidy/smallest-secured-golang-docker-image:$(VERSION)

.PHONY: start
start:	## - Run the smallest and secured golang docker image based on scratch
	@docker-compose up -d

.PHONY: stop
stop:	## - Run the smallest and secured golang docker image based on scratch
	@docker-compose down

.PHONY: logs
logs:	## - Run the smallest and secured golang docker image based on scratch
	@docker-compose logs

#run: dataset insert

#dataset:
#	wget http://mlg.ucd.ie/files/datasets/bbcsport-fulltext.zip
#	unzip bbcsport-fulltext.zip

.PHONY: insert
insert: ## insert news articlew
#	@docker run -v ./bbcsport:/bbcsport --dataset=./bbcsport -es=http://localhost:9200 --textbox=http://localhost:8000
	@docker run --rm --network textbox_elastic_indexer_default -v $$(pwd)/bbcsport/bbc-text.csv:/bbcsport/bbc-text.csv smallest-secured-golang:latest --dataset=/bbcsport/bbc-text.csv -es=http://elasticsearch:9200 --textbox=http://textbox:8080
