binary_name = telltime
binary_path = ./bin/${binary_name}
main_package_path = ./cmd/telltime

curr_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty)
linker_flags = '-s -X github.com/bnuredini/telltime/internal/conf.buildTime=${curr_time} -X github.com/bnuredini/telltime/internal/conf.version=${git_description}'

tailwindcss_input=./ui/static/css/input.css
tailwindcss_output=./ui/static/css/output.min.css

## build: build the application
.PHONY: build
build:
	go build -ldflags=${linker_flags} -o=${binary_path} ${main_package_path}

## build/linux: build the application for Linux
.PHONY: build/linux
build/linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=${binary_path} ${main_package_path}

## build/windows: build the application for Windows
.PHONY: build/windows
build/windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags=${linker_flags} -o=${binary_path} ${main_package_path}

## run: run the application
.PHONY: run
run:
	${binary_path}

## run/live: run the application with reloading on file changes
.PHONY: run/live
run/live:
	go run github.com/cosmtrek/air@v1.52.0 \
		--build.cmd "make build" \
		--build.bin "${binary_path}" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, gohtml, tpl, tmpl, html, css, scss, js, ts, sql" \
		--misc.clean_on_exit "true"

## tailwindcss: run tailwindcss
.PHONY: tailwindcss
tailwindcss:
	tailwindcss-linux-x64 -i ${tailwindcss_input} -o ${tailwindcss_output} --minify

## tailwindcss/live: run tailwindcss with reloading on file changes
.PHONY: tailwindcss/live
tailwindcss/live:
	tailwindcss-linux-x64 -i ${tailwindcss_input} -o ${tailwindcss_output} --watch --minify 2>&1 | sed "s/^/[TAILWINDCSS] /"
