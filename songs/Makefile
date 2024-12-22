GEN=./api/protogen
PROTOPATH=./api

protoc:
	protoc --go_out=${GEN} --go_opt=paths=source_relative \
    --go-grpc_out=${GEN} --go-grpc_opt=paths=source_relative \
	${PROTOPATH}/*.proto
	mv ${GEN}/api/* ${GEN}
	rmdir ${GEN}/api

tests:
	go test -timeout 1m -v --coverage ./internal/...

buf-update:
	buf dep update api

buf:
	buf generate

rerun:
	go build -o bin/app cmd/main/main.go
	./bin/app --config-path=bin/songs.yaml

dev:
	docker compose -f docker-compose.dev.yaml up -d

sqlc:
	sqlc generate