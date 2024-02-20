export GOOS=linux
export CGO_ENABLED=0
export GOARCH=amd64

go mod tidy
go build -o consumer .

docker buildx build --platform=linux/amd64 . -t frain/bench-convoy-consumer:v9
