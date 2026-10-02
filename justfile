# Default recipe
default: build

# Build the git-index binary
build:
    go build -ldflags="-s -w -X 'github.com/acmcelwee/git-index/cmd.Version=dev'" -o git-index main.go

# Run tests
test:
    go test -v ./...

# Tidy Go modules
tidy:
    go mod tidy

# Clean up built artifacts
clean:
    rm -f git-index
    rm -rf dist/

# Install the binary to ~/bin
install: build
    mkdir -p ~/bin
    cp git-index ~/bin/
