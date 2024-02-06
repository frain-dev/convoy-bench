export GOOS=linux
export CGO_ENABLED=0
export GOARCH=amd64

go mod tidy
go build -o consumer .

docker buildx build --platform=linux/amd64 . -t rtukpe/benchmarks-consumer:v9
docker push  rtukpe/benchmarks-consumer:v9