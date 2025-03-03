#!/bin/sh
set -e

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/runner_linux_amd64 main.go

# Check the operating system and handle UPX accordingly
compress_binary() {
    if ! "$1" bin/runner_linux_amd64; then
        echo "Warning: UPX compression failed, continuing with uncompressed binary"
    fi
}

if [ "$(uname)" = "Darwin" ]; then
    # On macOS, use brew-installed upx if available
    if command -v upx >/dev/null 2>&1; then
        compress_binary upx
    else
        echo "Warning: UPX not found. Install with 'brew install upx' for binary compression"
    fi
else
    # On Linux, download and use UPX as before
    if ! curl -L https://github.com/upx/upx/releases/download/v3.96/upx-3.96-amd64_linux.tar.xz | tar xJ; then
        echo "Warning: Failed to download/extract UPX, continuing without compression"
        exit 0
    fi
    compress_binary ./upx-3.96-amd64_linux/upx
    rm -rf upx-3.96-amd64_linux
fi

go install github.com/GeertJohan/go.rice/rice@v1.0.2
RICEBIN="$GOBIN"
if [ -z "$RICEBIN" ]; then
    if [ -z "$GOPATH" ]; then
        RICEBIN="$HOME"/go/bin
    else
        RICEBIN="$GOPATH"/bin
    fi
fi

"$RICEBIN"/rice embed-go -i github.com/khulnasoft/dazzle/pkg/test/runner

if [ $(ls -l bin/runner_linux_amd64 | cut -d ' ' -f 5) -gt 3437900 ]; then
    echo "runner binary is too big (> gRPC message size)"
    exit 1
fi