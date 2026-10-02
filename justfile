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

# Tag and push a new release to trigger GitHub Actions
release tag:
    #!/usr/bin/env bash
    echo "This will create and push the tag '{{tag}}' to trigger the GitHub release workflow."
    read -p "Are you sure you want to proceed? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag "{{tag}}"
        git push origin "{{tag}}"
        echo "Successfully pushed tag {{tag}}! GitHub Actions is now building the release."
    else
        echo "Release aborted."
    fi
