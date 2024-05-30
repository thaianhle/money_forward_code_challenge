package transaction

import (
	"context"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/repo"
)

type GetTransactionByUserIdReq struct {
	UserId uint32
	Query  *repo.Query
}

type GetTransactionsByUserId[TxType any] interface {
	Execute(ctx context.Context, req *GetTransactionByUserIdReq) ([]*aggregate.TransactionByDetails, error)
}

type defaultGetTransactionsByUserId[TxType any] struct {
	persistentRepo repo.TransactionRepo[TxType]
	cacheRepo      repo.TransactionCacheRepo
	logger         *zap.Logger
}

func NewGetTransactionsByUserId[TxType any](persistentRepo repo.TransactionRepo[TxType], cacheRepo repo.TransactionCacheRepo, logger *zap.Logger) GetTransactionsByUserId[TxType] {
	return &defaultGetTransactionsByUserId[TxType]{
		persistentRepo: persistentRepo,
		cacheRepo:      cacheRepo,
		logger:         logger,
	}
}

func (d *defaultGetTransactionsByUserId[TxType]) Execute(ctx context.Context, req *GetTransactionByUserIdReq) ([]*aggregate.TransactionByDetails, error) {
	return d.persistentRepo.GetByUserId(ctx, req.UserId, req.Query)
}
