APP := ms-teams-tui
BINARY := teams
VERSION := $(shell tr -d '[:space:]' < VERSION)
TAG := v$(VERSION)
LDFLAGS := -s -w -X main.version=$(TAG)
GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: version check-version bump-patch bump-minor bump-major test vet security build install update release-check release clean

version:
	@printf '%s\n' "$(VERSION)"

check-version:
	@./scripts/check-version.sh

bump-patch:
	@./scripts/bump-version.sh patch

bump-minor:
	@./scripts/bump-version.sh minor

bump-major:
	@./scripts/bump-version.sh major

test: check-version
	go test ./...

vet:
	go vet ./...

security:
	go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
	go run github.com/securego/gosec/v2/cmd/gosec@v2.27.1 -exclude-generated ./...

build: check-version
	@mkdir -p bin
	go build -trimpath -ldflags="$(LDFLAGS)" -o bin/$(BINARY) .

install: check-version
	@mkdir -p "$(GOBIN)"
	go build -trimpath -ldflags="$(LDFLAGS)" -o "$(GOBIN)/$(BINARY)" .
	@printf 'Installed %s %s to %s/%s\n' "$(APP)" "$(TAG)" "$(GOBIN)" "$(BINARY)"

update:
	git pull --ff-only origin main
	$(MAKE) test vet install

release-check: check-version
	@test "$$(git branch --show-current)" = "main" || { echo "release requires the main branch" >&2; exit 1; }
	@test -z "$$(git status --porcelain)" || { echo "release requires a clean worktree" >&2; exit 1; }
	@git fetch origin main --tags
	@test "$$(git rev-parse HEAD)" = "$$(git rev-parse origin/main)" || { echo "push main before releasing" >&2; exit 1; }
	@! git rev-parse "$(TAG)" >/dev/null 2>&1 || { echo "tag $(TAG) already exists" >&2; exit 1; }
	@! git ls-remote --exit-code --tags origin "refs/tags/$(TAG)" >/dev/null 2>&1 || { echo "remote tag $(TAG) already exists" >&2; exit 1; }
	$(MAKE) test vet security build
	@./bin/$(BINARY) --version | grep -Fx "$(APP) $(TAG)"

release: release-check
	git tag -a "$(TAG)" -m "$(APP) $(TAG)"
	git push origin "$(TAG)"
	@printf 'Release tag %s pushed. GitHub Actions will publish the release.\n' "$(TAG)"

clean:
	rm -rf bin dist
