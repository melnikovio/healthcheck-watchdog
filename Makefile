################################################################################
# Variables                                                                    #
################################################################################
export GO111MODULE ?= on
# By default, disable CGO_ENABLED. See the details on https://golang.org/cmd/cgo
CGO         ?= 0

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
   TARGET_OS_LOCAL = linux
else ifeq ($(LOCAL_OS),Darwin)
   TARGET_OS_LOCAL = darwin
else
   TARGET_OS_LOCAL ?= windows
endif
export GOOS ?= $(TARGET_OS_LOCAL)

################################################################################
# Target: build                                                                #
################################################################################
CGO_ENABLED=$(CGO) GOOS=$(3) GOARCH=amd64 go build -a -o service ./cmd/
