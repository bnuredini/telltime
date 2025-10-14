binary_name = telltime
binary_path = ./bin/${binary_name}
main_package_path = ./cmd/telltime

curr_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty)
linker_flags = '-s -X github.com/bnuredini/telltime/internal/conf.buildTime=${curr_time} -X github.com/bnuredini/telltime/internal/conf.version=${git_description}'

## build: build the application
.PHONY: build
build: 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=${binary_path} ${main_package_path}

## run: run the application
.PHONY: run
run:
	${binary_path}

