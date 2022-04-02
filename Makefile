postgres:
	docker run --name pg1 -e POSTGRES_PASSWORD=password -d -p 5432:5432 postgres

createdb:
	docker exec -it pg1 createdb --username=postgres --owner=postgres simple_bank

dropdb:
	docker exec -it pg1 dropdb --username=postgres simple_bank

migrateup:
	migrate -path db/migration/ -database "postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration/ -database "postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration/ -database "postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration/ -database "postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb --build_flags=--mod=mod -destination db/mock/store.go github.com/uwemakan/simplebank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server mock
