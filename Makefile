
# default location for files to scan
FILES ?= ~/tmp/fw

all: compile
	@echo valid targets are: compile, test, fmt, dist, clean and run

compile: molly

.PHONY: molly
molly:
	go build

run: compile output
	rm -rf output
	-./molly $(O) -R data/rules\
		-outdir output/extracted  -repdir output/reports \
		-on-tag "arm: arm-none-eabi-objdump -h {name}" \
		-on-rule "UImage: mkimage -l {name}" \
		-enable create-file \
		-disable execute \
		$(FILES)

show: run
	less  output/reports/report.json

test:
	go test -bench=. ./...

fmt:
	go fmt ./...

report:
	-go get -u github.com/client9/misspell/cmd/misspell
	-go get -u github.com/fzipp/gocyclo
	-misspell *.go lib
	-gocyclo -top 15 -avg .
	-go tool vet .


# published files are created here
dist: build compile
	mkdir -p build/dist
	VERSION=`./molly -V` make dist1

dist1:
	git archive master --format tar | bzip2 > build/dist/sources_$(VERSION).tar.bz2

	GOOS=linux GOARCH=amd64 make dist2
	GOOS=linux GOARCH=arm64 make dist2
	GOOS=linux GOARCH=arm make dist2
	GOOS=linux GOARCH=mips64 make dist2
#	GOOS=linux GOARCH=mipsel make dist2
	GOOS=freebsd GOARCH=amd64 make dist2
	GOOS=openbsd GOARCH=amd64 make dist2
	GOOS=windows GOARCH=amd64 make dist2
	GOOS=darwin GOARCH=amd64 make dist2

dist2: build
	mkdir -p build/dist/$(GOOS)_$(GOARCH)_$(VERSION)
	go build -o build/dist/$(GOOS)_$(GOARCH)_$(VERSION)/molly
	cp -r README.rst COPYING data/rules build/dist/$(GOOS)_$(GOARCH)_$(VERSION)
	cd build/dist/ && tar cjf $(GOOS)_$(GOARCH)_$(VERSION).tar.bz2 $(GOOS)_$(GOARCH)_$(VERSION)
	rm -rf "build/dist/$(GOOS)_$(GOARCH)_$(VERSION)"

build:
	mkdir build

output:
	mkdir output

clean:
	go clean
	rm -rf build output

.PHYONY: fmt clean run compile test molly
