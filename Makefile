CURDIR=$(shell pwd)
BINDIR=${CURDIR}/bin
GOVER=$(shell go version | perl -nle '/(go\d\S+)/; print $$1;')
MOCKGEN=${BINDIR}/mockgen_${GOVER}
SMARTIMPORTS=${BINDIR}/smartimports_${GOVER}
LINTVER=v1.49.0
LINTBIN=${BINDIR}/lint_${GOVER}_${LINTVER}
PACKAGE=github.com/SokolovAleksandr/AvitoTestBackend/cmd/back
GO_FILES=$(shell find $(CURDIR) -name '*.go')


all: format build test lint

build: bindir
	go build -o ${BINDIR}/back ${PACKAGE}

test:
	go test ./...

run: 
	mkdir -p logs/data
	go run ${PACKAGE} 2>&1 | tee logs/data/log.txt

lint: install-lint
	${LINTBIN} run

precommit: format build test lint
	echo "OK"

bindir:
	mkdir -p ${BINDIR}

format: install-smartimports
	${SMARTIMPORTS}

install-lint: bindir
	test -f ${LINTBIN} || \
		(GOBIN=${BINDIR} go install github.com/golangci/golangci-lint/cmd/golangci-lint@${LINTVER} && \
		mv ${BINDIR}/golangci-lint ${LINTBIN})

install-smartimports: bindir
	test -f ${SMARTIMPORTS} || \
		(GOBIN=${BINDIR} go install github.com/pav5000/smartimports/cmd/smartimports@latest && \
		mv ${BINDIR}/smartimports ${SMARTIMPORTS})

repository:
	cd repository && sudo docker compose up

docker:
	sudo docker build -t avito_test_backend .
	sudo docker run --rm -p 8080:8080 avito_test_backend 

docs:
	bash docs/gen.sh

.PHONY: all build test run lint precommit bindir format install-lint install-smartimports repository docker docs

