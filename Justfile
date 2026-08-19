# Divoene CRM — Go backend + React admin

default:
    @just --list

setup:
    npm install
    go mod download

build-admin:
    cd packages/admin && npm run build

build-server:
    podman build -f deploy/docker/Dockerfile.server -t divoene-server .

test:
    go test ./... -v -cover

vet:
    go vet ./...

coverage:
    go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

clean:
    rm -rf packages/admin/dist coverage.out
