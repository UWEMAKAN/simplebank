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
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb --build_flags=--mod=mod -destination db/mock/store.go github.com/uwemakan/simplebank/db/sqlc Store

proto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    proto/*.proto

evans:
	evans --host localhost --port 9090 -r repl

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 dbDocs dbSchema sqlc test server mock proto evans
