postgres:
	docker run --name pg1 -e POSTGRES_PASSWORD=password -d -p 5432:5432 postgres

createdb:
	docker exec -it pg1 createdb --username=postgres --owner=postgres simple_bank

dropdb:
	docker exec -it pg1 dropdb --username=postgres simple_bank

migrateup:
	migrate -path db/migration/ -database "postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration/ -database "postgresql://postgres:password@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown sqlc
