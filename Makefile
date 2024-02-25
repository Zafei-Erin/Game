build:
	@go build -o bin/tracker cmd/tracker/main.go
	@go build -o bin/game cmd/game/main.go

tracker: build
	@./bin/tracker 10999 15 10

game: build
	@./bin/game