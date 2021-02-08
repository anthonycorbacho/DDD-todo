# Default Go binary.
ifndef GOROOT
  GOROOT = /usr/local/go
endif

# Determine the OS to build.
ifeq ($(OS),)
  ifeq ($(shell  uname -s), Darwin)
    GOOS = darwin
  else
    GOOS = linux
  endif
else
  GOOS = $(OS)
endif

GOCMD = GOOS=$(GOOS) go
GOBUILD = CGO_ENABLED=0 $(GOCMD) build
GOTEST = $(GOCMD) test -race
RM = rm -rf
PROJECT = todo
DIST_DIR = ./dist
BUILD_PACKAGE = ./cmd/todo
GO_PKGS?=$$(go list ./... | grep -v /vendor/)

.PHONY: build

build:
		mkdir -p $(DIST_DIR)
		$(GOBUILD) -i -o $(DIST_DIR)/$(PROJECT)-$(GOOS) -v $(BUILD_PACKAGE)

test:
		$(GOTEST) -v $(GO_PKGS)

integration:
		$(GOTEST) -count=1 -v -tags integration $(GO_PKGS)

bench:
		$(GOCMD) test -tags integration -bench=. ./... -benchmem

clean:
		find . -type f -name '*~' -exec rm {} +
		find . -type f -name '\#*\#' -exec rm {} +
		find . -type f -name '*.coverprofile' -exec rm {} +
		$(RM) checkstyle.xml
		$(RM) $(DIST_DIR)/*

fclean: clean
		$(RM) $(DIST_DIR)

version:
		@echo $(VERSION)
