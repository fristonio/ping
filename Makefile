SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

ifeq ($(origin .RECIPEPREFIX), undefined)
  $(error This Make does not support .RECIPEPREFIX. Please use GNU Make 4.0 or later)
endif
.RECIPEPREFIX = >

export GO111MODULE := on
ROOTDIR := $(shell pwd)
VENDORDIR := $(ROOTDIR)/vendor
QUIET=@

GOOS ?=
GOOS := $(if $(GOOS),$(GOOS),linux)
GOARCH ?=
GOARCH := $(if $(GOARCH),$(GOARCH),amd64)
GOENV  := CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO     := $(GOENV) go
GO_BUILD := $(GO) build -trimpath

pkgs = $(shell $(GO) list ./... | grep -v vendor)

build:
> $(QUIET)echo "[*] Building ping"
> $(QUIET)./scripts/build.sh

build-linux: export GOOS=linux
build-linux: export GOARCH=amd64
build-linux:
> $(QUIET)echo "[*] Building ping"
> $(QUIET)./scripts/build.sh

format:
> $(QUIET)echo "[*] Formatting code"
> $(QUIET)$(GO) fmt $(pkgs)

govet:
> $(QUIET)echo "[*] Vetting code, checking for mistakes"
> $(QUIET)$(GO) vet $(pkgs)

.PHONY: build build-linux format govet