
# files to scan with make run
FILES ?= ~/tmp/fw

all: compile
	@echo valid targets are: compile, test, fmt, clean and run

compile: build/molly

build/molly: build
	go build -o build/molly

run: compile
	rm -rf build/extracted
	rm -rf build/report
	mkdir build/report
	-build/molly $(O) -R data/rules \
		-outdir build/extracted  -logdir build/report \
		-tagop "elf: ls {name}" \
		-tagop "executable: echo executable {name} contains {size} bytes" \
		-enable create-file \
		-disable execute \
		$(FILES)
	python -m json.tool build/report/report.json > build/report/report.pretty.json

test:
	go test ./...

fmt:
	go fmt

vet:
	go tool vet .

build:
	mkdir build
clean:
	go clean
	rm -rf build

.PHYONY: fmt clean run compile test
