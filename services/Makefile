SHELL := /bin/bash

.PHONY: test test-% lint protobuf protolint golint

lint: protolint golint

GO ?= go
TEST_FLAGS ?= -race
SERVICE ?=

TEST_SERVICES ?= tasks

test:
	@set -e; \
	if [ -z "$(SERVICE)" ]; then \
		services="$(TEST_SERVICES)"; \
		echo "tests all started"; \
	else \
		services="$(SERVICE)"; \
		echo "tests $(SERVICE) started"; \
	fi; \
	pass_total=0; fail_total=0; skip_total=0; total_total=0; \
	for s in $$services; do \
		tmp=$$(mktemp -t go-test.$$s.XXXXXX); \
		set -o pipefail; \
		$(GO) test -v $(TEST_FLAGS) ./$$s/tests 2>&1 | tee $$tmp; \
		rc=$$?; \
		total=$$(grep -cE '^=== RUN' $$tmp || true); \
		pass=$$(grep -cE '^[[:space:]]*--- PASS:' $$tmp || true); \
		fail=$$(grep -cE '^[[:space:]]*--- FAIL:' $$tmp || true); \
		skip=$$(grep -cE '^[[:space:]]*--- SKIP:' $$tmp || true); \
		rm -f $$tmp; \
		if [ $$total -gt 0 ]; then pct=$$((100*pass/total)); else pct=0; fi; \
		echo "tests $$s finished: $$pass/$$total ($$pct%)"; \
		pass_total=$$((pass_total+pass)); \
		fail_total=$$((fail_total+fail)); \
		skip_total=$$((skip_total+skip)); \
		total_total=$$((total_total+total)); \
		if [ $$rc -ne 0 ]; then exit $$rc; fi; \
	done; \
	if [ -z "$(SERVICE)" ]; then \
		if [ $$total_total -gt 0 ]; then pct=$$((100*pass_total/total_total)); else pct=0; fi; \
		echo "tests all finished: $$pass_total/$$total_total ($$pct%)"; \
	fi

test-%:
	@$(MAKE) test SERVICE=$*

protobuf:
	protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               proto/tasks/categories.proto
	protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               proto/tasks/tasks.proto

protolint:
	protolint .

golint:
	golangci-lint run -E gocritic -v ./...
