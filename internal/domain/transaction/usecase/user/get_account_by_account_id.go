package user

import (
	"context"
	"fmt"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/repo"

	"go.uber.org/zap"
)

type GetAccountByAccountId[TxType any] interface {
	Execute(ctx context.Context, req *GetAccountByAccountIdReq) (*aggregate.AccountByDetails, error)
}

type GetAccountByAccountIdReq struct {
	AccountId uint32 `json:"account_id"`
}

type defaultGetAccountByAccountId[TxType any] struct {
	persistentRepo repo.UserRepo[TxType]
	cacheRepo      repo.UserCacheRepo
	logger         *zap.Logger
}

func NewGetAccountByAccountId[TxType any](persistentRepo repo.UserRepo[TxType], cacheRepo repo.UserCacheRepo, logger *zap.Logger) GetAccountByAccountId[TxType] {

	return &defaultGetAccountByAccountId[TxType]{
		persistentRepo: persistentRepo,
		cacheRepo:      cacheRepo,
		logger:         logger,
	}
}

func (d *defaultGetAccountByAccountId[TxType]) Execute(ctx context.Context, req *GetAccountByAccountIdReq) (*aggregate.AccountByDetails, error) {
	accountDetail, err := d.cacheRepo.GetAccountByAccountId(ctx, req.AccountId)
	if err != nil {
		d.logger.Error("Get From Cache Failed, Try To Get From Persistent DB")
		accountDetail, err = d.persistentRepo.GetAccountByAccountId(ctx, req.AccountId)
		fmt.Println(accountDetail, req)
		if err != nil {
			return nil, err
		}

		_ = d.cacheRepo.SetAccount(ctx, accountDetail)
	} else {
		d.logger.Info("Get From Cache Success")
	}
	return accountDetail, nil
}
