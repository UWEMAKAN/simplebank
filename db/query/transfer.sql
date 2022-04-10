-- name: CreateTransfer :one
INSERT INTO transfers (
  from_account_id,
  to_account_id,
  amount,
  sender,
  recipient
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers
WHERE id = $1
AND (sender = sqlc.arg(username)
OR recipient = sqlc.arg(username))
LIMIT 1;

-- name: ListTransfers :many
SELECT * FROM transfers
WHERE id > $1
AND sender = $2
ORDER BY id
LIMIT $3;

-- name: ListTransfersByFromAccount :many
SELECT * FROM transfers
WHERE from_account_id = $1
AND id > $2
AND sender = $3
ORDER BY id
LIMIT $4;

-- name: ListTransfersByToAccount :many
SELECT * FROM transfers
WHERE to_account_id = $1
AND id > $2
AND sender = $3
ORDER BY id
LIMIT $4;

-- name: ListTransfersByFromAndToAccount :many
SELECT * FROM transfers
WHERE from_account_id = $1
AND to_account_id = $2
AND id > $3
AND sender = $4
ORDER BY id
LIMIT $5;
