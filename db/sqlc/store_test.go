package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	_ "github.com/uwemakan/simplebank/util"
)

func TestTransferTx(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// run a concurrent transfer transactions

	n := 20
	amount := float64(10)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := testStore.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
				Sender:        account1.Owner,
				Recipient:     account2.Owner,
			})
			errs <- err
			results <- result
		}()
	}

	// check response
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		// check errors
		err := <-errs
		require.NoError(t, err)

		// check results
		result := <-results
		require.NotEmpty(t, result)

		// check transfer
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount, transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = testStore.GetTransfer(context.Background(), GetTransferParams{
			ID:       transfer.ID,
			Username: account1.Owner,
		})
		require.NoError(t, err)

		// check entries
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = testStore.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = testStore.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		// check accounts' balances
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balances
	updatedAccount1, err := testStore.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testStore.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-float64(n)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+float64(n)*amount, updatedAccount2.Balance)
}

func TestTransferTxDeadlock(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	// run a concurrent transfer transactions

	n := 20
	amount := float64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID
		sender := account1.Owner
		recipient := account2.Owner

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
			sender = account2.Owner
			recipient = account1.Owner
		}

		go func() {
			_, err := testStore.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
				Sender:        sender,
				Recipient:     recipient,
			})
			errs <- err
		}()
	}

	// check response
	for i := 0; i < n; i++ {
		// check errors
		err := <-errs
		require.NoError(t, err)
	}

	// check the final updated balances
	updatedAccount1, err := testStore.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testStore.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updatedAccount1.Balance)
	require.Equal(t, account2.Balance, updatedAccount2.Balance)
}
