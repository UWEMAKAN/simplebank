DB_URL=postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable

postgres:
	docker run --name pg1 --network bank-network -e POSTGRES_PASSWORD=password -d -p 5432:5432 postgres

createdb:
	docker exec -it pg1 createdb --username=postgres --owner=postgres simple_bank

dropdb:
	docker exec -it pg1 dropdb --username=postgres simple_bank

migrateup:
	migrate -path db/migration/ -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration/ -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration/ -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration/ -database "$(DB_URL)" -verbose down 1

dbDocs:
	dbdocs build doc/db.dbml

dbSchema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen -package mockdb --build_flags=--mod=mod -destination db/mock/store.go github.com/uwemakan/simplebank/db/sqlc Store
	mockgen -package mockwk --build_flags=--mod=mod -destination worker/mock/distributor.go github.com/uwemakan/simplebank/worker TaskDistributor

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=api \
    proto/*.proto
	statik -src=./doc/swagger -dest=./doc

evans:
	evans --host localhost --port 9090 -r repl

redis:
	docker run --name simple_bank_redis -p 6379:6379 -d redis:7-alpine

redisPing:
	docker exec -it simple_bank_redis redis-cli ping

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

clearCache:
	go clean -testcache

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 dbDocs dbSchema sqlc test server mock proto evans redis redisPing new_migration clearCache
