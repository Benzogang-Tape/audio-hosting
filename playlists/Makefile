GEN=./api/protogen
PROTOPATH=./api

protoc:
	protoc --go_out=${GEN} --go_opt=paths=source_relative \
    --go-grpc_out=${GEN} --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=${GEN} --grpc-gateway_opt=paths=source_relative \
	--validate_out="lang=go:${GEN}" --validate_opt=paths=source_relative \
	${PROTOPATH}/*.proto
	mv ${GEN}/api/* ${GEN}
	rmdir ${GEN}/api

tests:
	go test ./internal/... -timeout 1m -v -cover