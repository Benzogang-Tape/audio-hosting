echo "Running unit tests with coverage..."
go mod tidy
go test -v -coverpkg=./internal/... -coverprofile=profile ./internal/... -timeout 1m
NUM=$(go tool cover -func profile | grep -E -o '[0-9]+\.[0-9]+' | tail -1 | tr -d '.')
if (( $NUM > 300 )); then
  NUM="${NUM%?}.${NUM: -1}"
  echo "Total coverage: $NUM%"
  echo "Unit tests passed successfully"
  exit 0
else
  echo "Coverage below 30%!"
  NUM="${NUM%?}.${NUM: -1}"
  echo "Total coverage: $NUM%"
  exit 1
fi