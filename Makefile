.PHONY: build clean deploy

build:
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/main cmd/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	serverless deploy --stage dev --verbose

function: clean build
	serverless deploy function --function Scraper --stage dev --verbose