.PHONY: tools

cover_dir=.cover
cover_profile=${cover_dir}/profile.out
filtered_cover_profile=${cover_dir}/filtered-profile.out
cover_html=${cover_dir}/coverage.html

.DEFAULT_GOAL := all

all: test

bin/golangci-lint: .golangci-version
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(shell cat .golangci-version)

lint: bin/golangci-lint
	bin/golangci-lint run

tools:
	go install github.com/mitranim/gow@latest

${cover_dir}:
	mkdir -p ${cover_dir}

${cover_profile}: ${cover_dir}
	go test -shuffle=on -failfast -coverprofile=${cover_profile} ./...
.PHONY: ${cover_profile}

${filtered_cover_profile}: ${cover_profile}
	cat ${cover_profile} | grep -E -v -f .cover-ignore > ${filtered_cover_profile}.tmp
	mv ${filtered_cover_profile}.tmp ${filtered_cover_profile}
	go tool cover -html=${filtered_cover_profile} -o ${cover_html}
	@echo "Coverage report: $(shell realpath ${cover_html})"

${cover_dir}/coverage-func.txt: ${filtered_cover_profile}
	go tool cover -func=${filtered_cover_profile} -o $@

${cover_dir}/coverage-total.txt: ${cover_dir}/coverage-func.txt
	 $(shell cat $? | grep total | grep -Eo '[0-9]+\.[0-9]+' > $@)

test: ${cover_dir}/coverage-total.txt .cover-threshold
	@eval total=$$(cat ${cover_dir}/coverage-total.txt); \
	eval threshold=$$(cat .cover-threshold); \
	echo "Total coverage: $${total}, threshold: $${threshold}"; \
	if [ $$(echo "$${total} < $${threshold}" | bc) -eq 1 ]; then \
	    echo "Error: coverage is below threshold" >&2; \
	    exit 1; \
	fi

test-bench:
	go test -bench=. -benchmem
