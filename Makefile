
# default location for files to scan
FILES ?= ~/tmp/fw

all: compile
	@echo valid targets are: compile, test, fmt, clean and run

compile: build/molly

build/molly: build
	go build -o build/molly

run: compile
	rm -rf build/extracted build/reports
	-build/molly $(O) -R data/rules\
		-outdir build/extracted  -repdir build/reports \
		-tagop "elf: ls -l {name}" \
		-enable create-file \
		-disable execute \
		$(FILES)

show: run
	less  build/report/report.json

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go tool vet .

build:
	mkdir build
clean:
	go clean
	rm -rf build

.PHYONY: fmt clean run compile test
