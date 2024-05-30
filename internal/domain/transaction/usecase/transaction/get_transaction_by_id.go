package transaction

import (
	"context"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/repo"
)

type GetTransactionByIdReq struct {
	Id uint32
}

type GetTransactionById[TxType any] interface {
	Execute(ctx context.Context, req *GetTransactionByIdReq) (*aggregate.TransactionByDetails, error)
}

type defaultGetTransactionByIdUseCase[TxType any] struct {
	persistentRepo repo.TransactionRepo[TxType]
	cacheRepo      repo.TransactionCacheRepo
	logger         *zap.Logger
}

func NewDefaultGetTransactionById[TxType any](persistentRepo repo.TransactionRepo[TxType], cacheRepo repo.TransactionCacheRepo, logger *zap.Logger) GetTransactionById[TxType] {
	return &defaultGetTransactionByIdUseCase[TxType]{
		persistentRepo: persistentRepo,
		cacheRepo:      cacheRepo,
		logger:         logger,
	}
}

func (d *defaultGetTransactionByIdUseCase[TxType]) Execute(ctx context.Context, req *GetTransactionByIdReq) (*aggregate.TransactionByDetails, error) {
	transactionDetail, err := d.cacheRepo.GetById(ctx, req.Id)
	if err != nil {
		transactionDetail, err = d.persistentRepo.GetById(ctx, req.Id)
		if err != nil {
			return nil, err
		}

		_ = d.cacheRepo.Set(ctx, transactionDetail)
		return transactionDetail, nil
	}

	return transactionDetail, nil
}
