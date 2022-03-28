-- name: CreateTransfer :one
INSERT INTO transfers (
  from_account_id,
  to_account_id,
  amount
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers
WHERE id = $1 LIMIT 1;

-- name: ListTransfers :many
SELECT * FROM transfers
WHERE id > $1
ORDER BY id
LIMIT $2;

-- name: ListTransfersByFromAccount :many
SELECT * FROM transfers
WHERE from_account_id = $1
AND id > $2
ORDER BY id
LIMIT $3;

-- name: ListTransfersByToAccount :many
SELECT * FROM transfers
WHERE to_account_id = $1
AND id > $2
ORDER BY id
LIMIT $3;

-- name: ListTransfersByFromAndToAccount :many
SELECT * FROM transfers
WHERE from_account_id = $1
AND to_account_id = $2
AND id > $3
ORDER BY id
LIMIT $4;
