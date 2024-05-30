package transaction

import (
	"context"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/repo"
)

type GetTransactionByAccountIdReq struct {
	AccountId uint32
	Query     *repo.Query
}

type GetTransactionsByAccountId[TxType any] interface {
	Execute(ctx context.Context, req *GetTransactionByAccountIdReq) ([]*aggregate.TransactionByDetails, error)
}

type defaultGetTransactionsByAccountIdUseCase[TxType any] struct {
	persistentRepo repo.TransactionRepo[TxType]
	cacheRepo      repo.TransactionCacheRepo
	logger         *zap.Logger
}

func NewDefaultGetTransactionsByAccountId[TxType any](persistentRepo repo.TransactionRepo[TxType], cacheRepo repo.TransactionCacheRepo, logger *zap.Logger) GetTransactionsByAccountId[TxType] {
	return &defaultGetTransactionsByAccountIdUseCase[TxType]{
		persistentRepo: persistentRepo,
		cacheRepo:      cacheRepo,
		logger:         logger,
	}

}

func (d *defaultGetTransactionsByAccountIdUseCase[TxType]) Execute(ctx context.Context, req *GetTransactionByAccountIdReq) ([]*aggregate.TransactionByDetails, error) {
	return d.persistentRepo.GetByAccountId(ctx, req.AccountId, req.Query)
}
