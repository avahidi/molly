
all:
	@echo valid targets are: build, test, fmt, clean and run

build:
	go build
	cd app && go build *.go

run:
	go build
	cd app && go run *.go

test:
	go test

fmt:
	go fmt
	cd app && go fmt


clean:
	go clean
	cd app && go clean

.PHYONY: fmt clean run build