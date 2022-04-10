package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uwemakan/simplebank/util"
)

func createRandomTransfer(t *testing.T) Transfer {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)

	arg := CreateTransferParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        util.RandomMoney(),
		Sender: fromAccount.Owner,
		Recipient: toAccount.Owner,
	}

	a, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotZero(t, a.ID)
	require.NotZero(t, a.CreatedAt)
	require.Equal(t, fromAccount.ID, a.FromAccountID)
	require.Equal(t, toAccount.ID, a.ToAccountID)
	require.Equal(t, arg.Amount, a.Amount)

	return a
}

func TestCreateTransfer(t *testing.T) {
	createRandomTransfer(t)
}

func TestGetTransfer(t *testing.T) {
	t1 := createRandomTransfer(t)

	t2, err := testQueries.GetTransfer(context.Background(), GetTransferParams{
		ID: t1.ID,
		Username: t1.Sender,
	})
	require.NoError(t, err)
	require.Equal(t, t1.ID, t2.ID)
	require.Equal(t, t1.Amount, t2.Amount)
	require.Equal(t, t1.CreatedAt, t2.CreatedAt)
	require.Equal(t, t1.FromAccountID, t2.FromAccountID)
	require.Equal(t, t1.ToAccountID, t2.ToAccountID)
	require.Equal(t, t1.Sender, t2.Sender)
	require.Equal(t, t1.Recipient, t2.Recipient)
}

func TestListTransfers(t *testing.T) {
	var transfer Transfer
	for i := 0; i < 10; i++ {
		transfer = createRandomTransfer(t)
	}

	arg := ListTransfersParams{
		ID:    0,
		Limit: 10,
		Sender : transfer.Sender,
	}

	ts, err := testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(ts), 1)
	for _, st := range ts {
		require.NotEmpty(t, st)
	}
}

func createRandomTransfersFromOneAccount(t *testing.T, fromAccountId int64, sender string, n int) {
	for i := 0; i < n; i++ {
		toAccount := createRandomAccount(t)

		arg := CreateTransferParams{
			FromAccountID: fromAccountId,
			ToAccountID:   toAccount.ID,
			Amount:        util.RandomMoney(),
			Sender: sender,
			Recipient: toAccount.Owner,
		}

		testQueries.CreateTransfer(context.Background(), arg)
	}
}

func TestListTransfersByFromAccount(t *testing.T) {
	fromAccount := createRandomAccount(t)
	createRandomTransfersFromOneAccount(t, fromAccount.ID, fromAccount.Owner, 10)

	arg := ListTransfersByFromAccountParams{
		ID:            0,
		FromAccountID: fromAccount.ID,
		Limit:         5,
		Sender: fromAccount.Owner,
	}

	ts, err := testQueries.ListTransfersByFromAccount(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, ts, 5)

	for _, tf := range ts {
		require.Equal(t, fromAccount.ID, tf.FromAccountID)
		require.Equal(t, fromAccount.Owner, tf.Sender)
	}
}

func createRandomTransfersFromAndToTwoAccounts(t *testing.T, fromAccountId int64, toAccountId int64, sender string, recipient string, n int) {
	for i := 0; i < n; i++ {
		arg := CreateTransferParams{
			FromAccountID: fromAccountId,
			ToAccountID:   toAccountId,
			Amount:        util.RandomMoney(),
			Sender: sender,
			Recipient: recipient,
		}

		testQueries.CreateTransfer(context.Background(), arg)
	}
}

func TestListTransfersByFromAndToAccount(t *testing.T) {
	fromAccount := createRandomAccount(t)
	toAccount := createRandomAccount(t)
	createRandomTransfersFromAndToTwoAccounts(t, fromAccount.ID, toAccount.ID, fromAccount.Owner, toAccount.Owner, 10)

	arg := ListTransfersByFromAndToAccountParams{
		ID:            0,
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Limit:         5,
		Sender: fromAccount.Owner,
	}

	ts, err := testQueries.ListTransfersByFromAndToAccount(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, ts, 5)

	for _, tf := range ts {
		require.Equal(t, fromAccount.ID, tf.FromAccountID)
		require.Equal(t, toAccount.ID, tf.ToAccountID)
	}
}

func createRandomTransfersToOneAccount(t *testing.T, toAccountId int64, recipient string, n int) Account {
	fromAccount := createRandomAccount(t)

	for i := 0; i < n; i++ {
		arg := CreateTransferParams{
			FromAccountID: fromAccount.ID,
			ToAccountID:   toAccountId,
			Amount:        util.RandomMoney(),
			Sender: fromAccount.Owner,
			Recipient: recipient,
		}

		testQueries.CreateTransfer(context.Background(), arg)
	}
	return fromAccount
}

func TestListTransfersByToAccount(t *testing.T) {
	toAccount := createRandomAccount(t)
	fromAccount := createRandomTransfersToOneAccount(t, toAccount.ID, toAccount.Owner, 10)

	arg := ListTransfersByToAccountParams{
		ID:          0,
		ToAccountID: toAccount.ID,
		Limit:       5,
		Sender: fromAccount.Owner,
	}

	ts, err := testQueries.ListTransfersByToAccount(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, ts, 5)

	for _, tf := range ts {
		require.Equal(t, toAccount.ID, tf.ToAccountID)
	}
}
