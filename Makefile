
# default location for files to scan
FILES ?= ~/tmp/fw

all: compile
	@echo valid targets are: compile, test, fmt, dist, clean and run

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

# published files are created here
dist: build compile
	mkdir -p build/dist
	VERSION=`build/molly -VV` make dist1

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

clean:
	go clean
	rm -rf build

.PHYONY: fmt clean run compile test
