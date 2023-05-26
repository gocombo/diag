.PHONY: tools .cover-packages

cover_dir=.cover
cover_profile=${cover_dir}/profile.out
cover_html=${cover_dir}/coverage.html

.DEFAULT_GOAL := all

all: test

bin/golangci-lint: .golangci-version
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(shell cat .golangci-version)

lint: bin/golangci-lint
	bin/golangci-lint run

${cover_dir}:
	mkdir -p ${cover_dir}

tools:
	go install github.com/mitranim/gow@latest

.cover-packages:
	go list ./... | grep -v -f .cover-ignore  > $@.tmp
	awk '{print $2}' $@.tmp | paste -s -d, - > $@
	rm $@.tmp

test: lint ${cover_dir} .cover-packages
	go test -coverpkg=$(shell cat .cover-packages) -coverprofile=${cover_profile} ./...
	go tool cover -html=${cover_profile} -o ${cover_html}

test-bench:
	go test -bench=. -benchmem

${cover_dir}/coverage-func.txt: ${cover_profile}
	go tool cover -func=${cover_profile} -o $@

${cover_dir}/coverage-total.txt: ${cover_dir}/coverage-func.txt
	 $(shell cat $? | grep total | grep -Eo '[0-9]+\.[0-9]+' > $@)

test-cover: test ${cover_dir}/coverage-total.txt .cover-threshold
	@eval total=$$(cat ${cover_dir}/coverage-total.txt); \
	eval threshold=$$(cat .cover-threshold); \
	echo "Total coverage: $${total}, threshold: $${threshold}"; \
	if [ $$(echo "$${total} < $${threshold}" | bc) -eq 1 ]; then \
	    echo "Error: coverage is below threshold" >&2; \
	    exit 1; \
	fi
