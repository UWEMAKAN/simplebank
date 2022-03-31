package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uwemakan/simplebank/util"
)

func createRandomEntry(t *testing.T) Entry {
	a := createRandomAccount(t)
	arg := CreateEntryParams{
		AccountID: a.ID,
		Amount:    util.RandomMoney(),
	}

	e, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, e)
	require.Equal(t, a.ID, e.AccountID)
	require.Equal(t, arg.Amount, e.Amount)
	require.NotZero(t, e.ID)
	require.NotZero(t, e.CreatedAt)

	return e
}

func TestCreateEntry(t *testing.T) {
	createRandomEntry(t)
}

func TestGetEntry(t *testing.T) {
	e1 := createRandomEntry(t)

	e2, err := testQueries.GetEntry(context.Background(), e1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, e2)
	require.Equal(t, e1.ID, e2.ID)
	require.Equal(t, e1.Amount, e2.Amount)
	require.Equal(t, e1.AccountID, e2.AccountID)
	require.Equal(t, e1.CreatedAt, e2.CreatedAt)

	e3, err := testQueries.GetEntry(context.Background(), 0)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, e3)
}

func TestListEntries(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomEntry(t)
	}

	arg := ListEntriesParams{
		ID:    0,
		Limit: 5,
	}
	es, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, es, 5)

	for _, e := range es {
		require.NotEmpty(t, e)
	}
}

func createAccountEntries(t *testing.T, accountId int64, n int) {
	for i := 0; i < n; i++ {
		arg := CreateEntryParams{
			AccountID: accountId,
			Amount:    util.RandomMoney(),
		}

		e, err := testQueries.CreateEntry(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, e)
		require.Equal(t, accountId, e.AccountID)
		require.Equal(t, arg.Amount, e.Amount)
		require.NotZero(t, e.ID)
		require.NotZero(t, e.CreatedAt)
	}
}

func TestListEntriesByAccountId(t *testing.T) {
	account := createRandomAccount(t)
	createAccountEntries(t, account.ID, 10)

	arg := ListEntriesByAccountIdParams{
		ID:        0,
		AccountID: account.ID,
		Limit:     5,
	}

	es, err := testQueries.ListEntriesByAccountId(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, es, 5)

	for _, e := range es {
		require.Equal(t, e.AccountID, account.ID)
		require.NotEmpty(t, e)
	}
}
