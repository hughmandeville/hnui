# Makefile

## HELP:
.PHONY: help
## help: Show this help message.
help:
	@echo "Usage: make [target]\n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## :
## BUILD:

.PHONY: build-go
## build-go: Build Go code.
build-go:
	go build -o /dev/null ./...

.PHONY: build-node
## build-node: Build React client code.
build-node:
	cd client; yarn && yarn build

.PHONY: clean-node
## clean-node: Remove React build directory and Node modules.
clean-node:
	@rm -rf client/build
	@rm -rf client/node_modules

## :
## DEPENDENCIES:

.PHONY: dep-get-go
## dep-get-go: Get Go modules.
dep-get-go:
	go mod tidy

.PHONY: dep-get-node
## dep-get-node: Get Node modules.
dep-get-node:
	cd client; yarn

.PHONY: dep-init-go
## dep-init-go: Initialize Go modules.
dep-init-go:
	go mod init
	go mod tidy

.PHONY: dep-update-all
## dep-update-all: Update Go modules and Node modules.
dep-update-all: dep-update-go dep-update-node

.PHONY: dep-update-go
## dep-update-go: Update Go modules.
dep-update-go:
	go get -u ./...

.PHONY: dep-update-node
## dep-update-node: Update Node modules.
dep-update-node:
	cd client; yarn upgrade

## :
## GITHUB:

.PHONY: gh-deploy
## gh-deploy: Deploy React client to GitHub Pages.
gh-deploy:
	cd client; yarn deploy

## :
## HACKER_NEWS:

.PHONY: hn-ts-get
## hn-ts-get: Get Hacker News Top Stories JSON.
hn-ts-get:
	@go run cmd/get_hn/get_hn.go -out file -verbose

.PHONY: hn-ts-upload
## hn-ts-upload: Upload Hacker News Top Stories JSON to GCS.
hn-ts-upload:
	@go run cmd/get_hn/get_hn.go -out github -verbose


## :
## RUN:

.PHONY: run-node
## run-node: Run React client locally (on port 3000).
run-node:
	cd client; yarn start

## :
