package user

import (
	"context"
	"money_forward_code_challenge/internal/domain/transaction/models"
	"money_forward_code_challenge/internal/domain/transaction/repo"
	"money_forward_code_challenge/pkgs/repo_pool_async"

	"go.uber.org/zap"
)

type UpdateBalanceAccountReq struct {
	AccountId       uint32
	Amount          float32
	OldBalance      float32
	TransactionType string
}
type UpdateBalanceAccountUseCase[TxType any] interface {
	Execute(ctx context.Context, req *UpdateBalanceAccountReq, tx TxType) (*repo_pool_async.Job, error)
}

type defaultUpdateBalanceAccountUseCase[TxType any] struct {
	persistentRepo repo.UserRepo[TxType]
	cacheRepo      repo.UserCacheRepo
	pool           *repo_pool_async.RepoUpdatePoolBusyWaiting
	logger         *zap.Logger
}

func NewUpdateBalanceAccountUseCase[TxType any](persistentRepo repo.UserRepo[TxType],
	cacheRepo repo.UserCacheRepo, logger *zap.Logger,
	poolSizeWorker int,
) UpdateBalanceAccountUseCase[TxType] {
	return &defaultUpdateBalanceAccountUseCase[TxType]{
		persistentRepo: persistentRepo,
		cacheRepo:      cacheRepo,
		pool:           repo_pool_async.NewPool(context.TODO(), poolSizeWorker, logger),
		logger:         logger,
	}
}

func (d *defaultUpdateBalanceAccountUseCase[TxType]) Execute(ctx context.Context, req *UpdateBalanceAccountReq, tx TxType) (*repo_pool_async.Job, error) {

	account, err := d.cacheRepo.GetAccountByAccountId(ctx, req.AccountId)
	if err != nil {
		return nil, err
	}

	if req.TransactionType == models.TRANSACTIONTYPEDEPOSIT {
		account.Balance += req.Amount
	} else if req.TransactionType == models.TRANSACTIONTYPEWITHDRAW {
		account.Balance -= req.Amount
	}

	err = d.persistentRepo.UpdateBalance(ctx, req.AccountId, account.Balance, tx)

	if err != nil {
		return nil, err
	}

	job := d.pool.PushPriority(ctx, func(ctx context.Context) {
		d.logger.Info("update balance account")
		d.cacheRepo.SetAccount(ctx, account)
	})

	return job, nil
}
