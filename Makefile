
all:
	@echo valid targets are: build, test, fmt, clean and run

build:
	go build
	cd app && go build *.go

run:
	go build
	rm -rf ~/tmp/fw_output
	cd app && go run *.go -R ../rules -O ~/tmp/fw_output ~/tmp/fw/

test:
	go test

fmt:
	go fmt
	cd app && go fmt


clean:
	go clean
	cd app && go clean

.PHYONY: fmt clean run build