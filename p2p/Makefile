PACKAGE := github.com/taurusgroup/multi-party-sig/p2p/cmd
REVISION := $(shell git rev-parse --short HEAD)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD | tr -d '\040\011\012\015\n')
DATE := $(shell date  +"%Y-%m-%d")


p2p:
	go build -ldflags "-X $(PACKAGE).Version=1.0.0 -X $(PACKAGE).Commit=$(REVISION) -X $(PACKAGE).BuildTime=$(DATE)"