package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/uwemakan/simplebank/db/sqlc"
	"github.com/uwemakan/simplebank/token"
)

type transferRequest struct {
	FromAccountID int64   `json:"fromAccountId" binding:"required,min=1"`
	ToAccountID   int64   `json:"toAccountId" binding:"required,min=1"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	toAccount, valid := server.validAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Sender:        fromAccount.Owner,
		Recipient:     toAccount.Owner,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, result)
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)

	if err != nil {
		if err == db.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		err = fmt.Errorf("account [%d] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	return account, true
}

type getTransferRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getTransfer(ctx *gin.Context) {
	var req getTransferRequest

	if err := ctx.BindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.GetTransferParams{
		ID:       req.ID,
		Username: authPayload.Username,
	}

	result, err := server.store.GetTransfer(ctx, arg)
	if err != nil {
		if err == db.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

type listTransfersRequest struct {
	PageID        int32 `form:"pageId" binding:"omitempty,min=0"`
	PageSize      int32 `form:"pageSize" binding:"required,min=5,max=10"`
	FromAccountID int64 `form:"fromAccountId" binding:"omitempty,min=1"`
	ToAccountID   int64 `form:"toAccountId" binding:"omitempty,min=1"`
}

func (server *Server) listTransfers(ctx *gin.Context) {
	var req listTransfersRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	result, err := getAllTransfers(ctx, server, req)

	if err != nil {
		if err == db.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func getAllTransfers(ctx *gin.Context, server *Server, req listTransfersRequest) (result []db.Transfer, err error) {
	if req.FromAccountID != 0 && req.ToAccountID == 0 {
		arg := db.ListTransfersByFromAccountParams{
			ID:            int64(req.PageID),
			Limit:         req.PageSize,
			FromAccountID: req.FromAccountID,
		}
		result, err = server.store.ListTransfersByFromAccount(ctx, arg)
	} else if req.FromAccountID == 0 && req.ToAccountID != 0 {
		arg := db.ListTransfersByToAccountParams{
			ID:          int64(req.PageID),
			Limit:       req.PageSize,
			ToAccountID: req.ToAccountID,
		}
		result, err = server.store.ListTransfersByToAccount(ctx, arg)
	} else if req.FromAccountID != 0 && req.ToAccountID != 0 {
		arg := db.ListTransfersByFromAndToAccountParams{
			ID:            int64(req.PageID),
			Limit:         req.PageSize,
			FromAccountID: req.FromAccountID,
			ToAccountID:   req.ToAccountID,
		}
		result, err = server.store.ListTransfersByFromAndToAccount(ctx, arg)
	} else {
		arg := db.ListTransfersParams{
			ID:    int64(req.PageID),
			Limit: req.PageSize,
		}
		result, err = server.store.ListTransfers(ctx, arg)
	}
	return result, err
}
