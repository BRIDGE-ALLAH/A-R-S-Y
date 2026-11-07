lint:
	golangci-lint run -c ./golangci.yml ./...

test:
	go test ./... -v --cover

jstypes:
	go run ./plugins/jsvm/internal/types/types.go

test-report:
	go test ./... -v --cover -coverprofile=coverage.out
	go tool cover -html=coverage.out

dashboard:
	cd dashboard && npm install --no-audit --no-fund && npm run build

build: dashboard
	mkdir -p .builds
	CGO_ENABLED=0 go build -o .builds/arsy ./examples/base
