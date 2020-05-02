GOPATH:=$(shell go env GOPATH)

.PHONY: test
test:
	go test ./... -v -race -bench=. | sed ''/PASS/s//$$(printf "\033[32mPASS\033[0m")/'' | sed ''/FAIL/s//$$(printf "\033[31mFAIL\033[0m")/''
