package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
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

func TestGetTransferAPI(t *testing.T) {
	transfer := randomTransfer(1, 2, 100.45)

	testCases := []struct {
		name          string
		transferID    int64
		fromAccountID int64
		toAccountID   int64
		amount        float64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:          "OK",
			transferID:    transfer.ID,
			fromAccountID: transfer.FromAccountID,
			toAccountID:   transfer.ToAccountID,
			amount:        transfer.Amount,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(transfer, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfer(t, recorder.Body, transfer)
			},
		},
		{
			name:          "BadRequest",
			transferID:    0,
			fromAccountID: transfer.FromAccountID,
			toAccountID:   transfer.ToAccountID,
			amount:        transfer.Amount,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:          "NotFound",
			transferID:    transfer.ID,
			fromAccountID: transfer.FromAccountID,
			toAccountID:   transfer.ToAccountID,
			amount:        transfer.Amount,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(db.Transfer{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:          "InternalServerError",
			transferID:    transfer.ID,
			fromAccountID: transfer.FromAccountID,
			toAccountID:   transfer.ToAccountID,
			amount:        transfer.Amount,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(db.Transfer{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/transfers/%d", tc.transferID)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchTransfer(t *testing.T, body *bytes.Buffer, transfer db.Transfer) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotTransfer db.Transfer
	err = json.Unmarshal(data, &gotTransfer)
	require.NoError(t, err)
	require.Equal(t, transfer, gotTransfer)
}

func TestListTransfersApI(t *testing.T) {
	n := 20
	transfers := make([]db.Transfer, n)

	fromAccountID := util.RandomInt(1, 100)
	toAccountID := util.RandomInt(1, 100)
	amount := util.RandomMoney()
	pageID := 0
	pageSize := 10

	for i := 0; i < n; i++ {
		transfers[i] = randomTransfer(fromAccountID, toAccountID, amount)
	}

	testCases := []struct {
		name          string
		pageID        int32
		pageSize      int32
		fromAccountID int64
		toAccountID   int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:          "ListTransfersOK",
			pageID:        int32(pageID),
			pageSize:      int32(pageSize),
			fromAccountID: 0,
			toAccountID:   0,
			buildStubs:    func(store *mockdb.MockStore) {
				store.EXPECT().ListTransfers(gomock.Any(), gomock.Eq(db.ListTransfersParams{
					ID: int64(pageID),
					Limit: int32(pageSize),
				})).
				Times(1).
				Return(transfers[:pageSize], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransferSlice(t, recorder.Body, transfers[:pageSize])
			},
		},
		{
			name:          "ListTransfersByFromAccountOK",
			pageID:        int32(pageID),
			pageSize:      int32(pageSize),
			fromAccountID: fromAccountID,
			toAccountID:   0,
			buildStubs:    func(store *mockdb.MockStore) {
				store.EXPECT().ListTransfersByFromAccount(gomock.Any(), gomock.Eq(db.ListTransfersByFromAccountParams{
					ID: int64(pageID),
					Limit: int32(pageSize),
					FromAccountID: fromAccountID,
				})).
				Times(1).
				Return(transfers[:pageSize], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransferSlice(t, recorder.Body, transfers[:pageSize])
			},
		},
		{
			name:          "ListTransfersByToAccountOK",
			pageID:        int32(pageID),
			pageSize:      int32(pageSize),
			fromAccountID: 0,
			toAccountID:   toAccountID,
			buildStubs:    func(store *mockdb.MockStore) {
				store.EXPECT().ListTransfersByToAccount(gomock.Any(), gomock.Eq(db.ListTransfersByToAccountParams{
					ID: int64(pageID),
					Limit: int32(pageSize),
					ToAccountID: toAccountID,
				})).
				Times(1).
				Return(transfers[:pageSize], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransferSlice(t, recorder.Body, transfers[:pageSize])
			},
		},
		{
			name:          "ListTransfersByFromAndToAccountOK",
			pageID:        int32(pageID),
			pageSize:      int32(pageSize),
			fromAccountID: fromAccountID,
			toAccountID:   toAccountID,
			buildStubs:    func(store *mockdb.MockStore) {
				store.EXPECT().ListTransfersByFromAndToAccount(gomock.Any(), gomock.Eq(db.ListTransfersByFromAndToAccountParams{
					ID: int64(pageID),
					Limit: int32(pageSize),
					FromAccountID: fromAccountID,
					ToAccountID: toAccountID,
				})).
				Times(1).
				Return(transfers[:pageSize], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransferSlice(t, recorder.Body, transfers[:pageSize])
			},
		},
		{
			name:          "BadRequest",
			pageID:        int32(pageID),
			pageSize:      0,
			fromAccountID: fromAccountID,
			toAccountID:   toAccountID,
			buildStubs:    func(store *mockdb.MockStore) {
				store.EXPECT().ListTransfersByFromAndToAccount(gomock.Any(), gomock.Eq(db.ListTransfersByFromAndToAccountParams{
					ID: int64(pageID),
					Limit: int32(pageSize),
					FromAccountID: fromAccountID,
					ToAccountID: toAccountID,
				})).
				Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:          "NotFound",
			pageID:        int32(pageID),
			pageSize:      int32(pageSize),
			fromAccountID: fromAccountID,
			toAccountID:   toAccountID,
			buildStubs:    func(store *mockdb.MockStore) {
				store.EXPECT().ListTransfersByFromAndToAccount(gomock.Any(), gomock.Eq(db.ListTransfersByFromAndToAccountParams{
					ID: int64(pageID),
					Limit: int32(pageSize),
					FromAccountID: fromAccountID,
					ToAccountID: toAccountID,
				})).
				Times(1).
				Return([]db.Transfer{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:          "BadRequest",
			pageID:        int32(pageID),
			pageSize:      int32(pageSize),
			fromAccountID: fromAccountID,
			toAccountID:   toAccountID,
			buildStubs:    func(store *mockdb.MockStore) {
				store.EXPECT().ListTransfersByFromAndToAccount(gomock.Any(), gomock.Eq(db.ListTransfersByFromAndToAccountParams{
					ID: int64(pageID),
					Limit: int32(pageSize),
					FromAccountID: fromAccountID,
					ToAccountID: toAccountID,
				})).
				Times(1).
				Return([]db.Transfer{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

			url := fmt.Sprintf("/transfers?pageId=%d&pageSize=%d&fromAccountId=%d&toAccountId=%d", tc.pageID, tc.pageSize, tc.fromAccountID, tc.toAccountID)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchTransferSlice(t *testing.T, body *bytes.Buffer, transfers []db.Transfer) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotTransfers []db.Transfer
	err = json.Unmarshal(data, &gotTransfers)
	require.NoError(t, err)
	require.Equal(t, transfers, gotTransfers)
}