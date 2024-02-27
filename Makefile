build:
	@go build -o bin/tracker cmd/tracker/main.go
	@go build -o bin/game cmd/game/*.go

tracker: build
	@./bin/tracker 10999 15 10

game: build
	@./bin/game 127.0.0.1 10999 $(id)