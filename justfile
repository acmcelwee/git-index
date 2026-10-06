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
release tag="":
    #!/usr/bin/env bash
    actual_tag="{{tag}}"
    if [ -z "$actual_tag" ]; then
        latest=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
        if [[ $latest =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
            major="${BASH_REMATCH[1]}"
            minor="${BASH_REMATCH[2]}"
            patch="${BASH_REMATCH[3]}"
            actual_tag="v${major}.${minor}.$((patch + 1))"
        else
            actual_tag="v0.0.1"
        fi
        echo "No tag provided. Latest tag is $latest. Suggested next tag is $actual_tag."
    fi

    echo "This will create and push the tag '$actual_tag' to trigger the GitHub release workflow."
    read -p "Are you sure you want to proceed? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag "$actual_tag"
        git push origin "$actual_tag"
        echo "Successfully pushed tag $actual_tag! GitHub Actions is now building the release."
    else
        echo "Release aborted."
    fi
