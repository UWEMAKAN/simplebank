package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/uwemakan/simplebank/db/mock"
	db "github.com/uwemakan/simplebank/db/sqlc"
	"github.com/uwemakan/simplebank/util"
)

func TestCreateTransferAPI(t *testing.T) {
	transferTx := randomTransferTx()
	currency := util.RUB

	transferTx.FromAccount.Currency = currency
	transferTx.ToAccount.Currency = currency

	testCases := []struct {
		name          string
		fromAccountID int64
		toAccountID   int64
		amount        float64
		currency      string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:          "Created",
			fromAccountID: transferTx.FromAccount.ID,
			toAccountID:   transferTx.ToAccount.ID,
			amount:        transferTx.Transfer.Amount,
			currency:      currency,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.FromAccount.ID)).Times(1).Return(transferTx.FromAccount, nil),
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.ToAccount.ID)).Times(1).Return(transferTx.ToAccount, nil),
					store.EXPECT().
						TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
							FromAccountID: transferTx.FromAccount.ID,
							ToAccountID:   transferTx.ToAccount.ID,
							Amount:        transferTx.Transfer.Amount,
						})).
						Times(1).
						Return(transferTx, nil),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchTransferTx(t, recorder.Body, transferTx)
			},
		},
		{
			name:          "FromAccountNotFound",
			fromAccountID: transferTx.FromAccount.ID,
			toAccountID:   transferTx.ToAccount.ID,
			amount:        transferTx.Transfer.Amount,
			currency:      currency,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.FromAccount.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:          "FromAccountInternalServerError",
			fromAccountID: transferTx.FromAccount.ID,
			toAccountID:   transferTx.ToAccount.ID,
			amount:        transferTx.Transfer.Amount,
			currency:      currency,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.FromAccount.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:          "FromAccountCurrencyMismatchBadRequest",
			fromAccountID: transferTx.FromAccount.ID,
			toAccountID:   transferTx.ToAccount.ID,
			amount:        transferTx.Transfer.Amount,
			currency:      util.USD,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.FromAccount.ID)).Times(1).Return(transferTx.FromAccount, nil),
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.ToAccount.ID)).Times(0),
					store.EXPECT().
						TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
							FromAccountID: transferTx.FromAccount.ID,
							ToAccountID:   transferTx.ToAccount.ID,
							Amount:        transferTx.Transfer.Amount,
						})).
						Times(0),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:          "ToAccountNotFound",
			fromAccountID: transferTx.FromAccount.ID,
			toAccountID:   transferTx.ToAccount.ID,
			amount:        transferTx.Transfer.Amount,
			currency:      currency,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.FromAccount.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:          "ToAccountInternalServerError",
			fromAccountID: transferTx.FromAccount.ID,
			toAccountID:   transferTx.ToAccount.ID,
			amount:        transferTx.Transfer.Amount,
			currency:      currency,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.FromAccount.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:          "BadRequest",
			fromAccountID: transferTx.FromAccount.ID,
			toAccountID:   transferTx.ToAccount.ID,
			amount:        transferTx.Transfer.Amount,
			currency:      "ABC",
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.FromAccount.ID)).Times(0),
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.ToAccount.ID)).Times(0),
					store.EXPECT().
						TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
							FromAccountID: transferTx.FromAccount.ID,
							ToAccountID:   transferTx.ToAccount.ID,
							Amount:        transferTx.Transfer.Amount,
						})).
						Times(0))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:          "InternalServerError",
			fromAccountID: transferTx.FromAccount.ID,
			toAccountID:   transferTx.ToAccount.ID,
			amount:        transferTx.Transfer.Amount,
			currency:      currency,
			buildStubs: func(store *mockdb.MockStore) {
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.FromAccount.ID)).Times(1).Return(transferTx.FromAccount, nil),
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.ToAccount.ID)).Times(1).Return(transferTx.ToAccount, nil),
					store.EXPECT().
						TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
							FromAccountID: transferTx.FromAccount.ID,
							ToAccountID:   transferTx.ToAccount.ID,
							Amount:        transferTx.Transfer.Amount,
						})).
						Times(1).
						Return(db.TransferTxResult{}, sql.ErrConnDone),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:          "ToAccountCurrencyMismatchBadRequest",
			fromAccountID: transferTx.FromAccount.ID,
			toAccountID:   transferTx.ToAccount.ID,
			amount:        transferTx.Transfer.Amount,
			currency:      currency,
			buildStubs: func(store *mockdb.MockStore) {
				transferTx.ToAccount.Currency = util.USD
				gomock.InOrder(
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.FromAccount.ID)).Times(1).Return(transferTx.FromAccount, nil),
					store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(transferTx.ToAccount.ID)).Times(1).Return(transferTx.ToAccount, nil),
					store.EXPECT().
						TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
							FromAccountID: transferTx.FromAccount.ID,
							ToAccountID:   transferTx.ToAccount.ID,
							Amount:        transferTx.Transfer.Amount,
						})).
						Times(0),
				)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/transfers"

			data := transferRequest{
				FromAccounID: tc.fromAccountID,
				ToAccounID:   tc.toAccountID,
				Amount:       tc.amount,
				Currency:     tc.currency,
			}

			b, err := json.Marshal(data)
			require.NoError(t, err)

			body := bytes.NewReader(b)
			request, err := http.NewRequest(http.MethodPost, url, body)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchTransferTx(t *testing.T, body *bytes.Buffer, transferTx db.TransferTxResult) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotTransferTx db.TransferTxResult
	err = json.Unmarshal(data, &gotTransferTx)
	require.NoError(t, err)
	require.Equal(t, transferTx, gotTransferTx)
}

func randomTransferTx() db.TransferTxResult {
	fromAccount := randomAccount()
	toAccount := randomAccount()
	amount := util.RandomMoney()

	return db.TransferTxResult{
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		FromEntry:   randomEntry(fromAccount.ID, -amount),
		ToEntry:     randomEntry(toAccount.ID, amount),
		Transfer:    randomTransfer(fromAccount.ID, toAccount.ID, amount),
	}
}

func randomTransfer(fromAccountID, toAccountID int64, amount float64) db.Transfer {
	return db.Transfer{
		ID:            util.RandomInt(1, 1000),
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
	}
}

func randomEntry(accountID int64, amount float64) db.Entry {
	return db.Entry{
		ID:        util.RandomInt(1, 1000),
		AccountID: accountID,
		Amount:    amount,
	}
}
