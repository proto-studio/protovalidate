ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

test:
	go test ./...

test-docker:
	docker run -it -v "${ROOT_DIR}:/usr/src/build" -w /usr/src/build --rm golang:1.20 make test

bench:
	go test -bench=. -benchmem ./...

coverage:
	go test -coverpkg=./... -coverprofile=coverage.out ./...
	go tool cover -func coverage.out

coverage-html: coverage
	go tool cover -html=coverage.out

reportcard:
	goreportcard-cli -v

generated:
	go generate ./pkg/...
