package transaction

import (
	"context"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/models"
	"money_forward_code_challenge/internal/domain/transaction/repo"
	"money_forward_code_challenge/pkgs/repo_pool_async"
)

type DeleteReq struct {
	AccountId     uint32 `json:"account_id"`
	TransactionId uint32 `json:"transaction_id"`
}
type DeleteTransactionById[TxType any] interface {
	Execute(ctx context.Context, req *DeleteReq, tx TxType) (*aggregate.TransactionByDetails, *repo_pool_async.Job, error)
}

type defaultDeleteTransactionByAccountIdUseCase[TxType any] struct {
	persistentRepo repo.TransactionRepo[TxType]
	cacheRepo      repo.TransactionCacheRepo
	pool           *repo_pool_async.RepoUpdatePoolBusyWaiting
	logger         *zap.Logger
}

func NewDeleteTransactionByAccountId[TxType any](persistentRepo repo.TransactionRepo[TxType],
	cacheRepo repo.TransactionCacheRepo, logger *zap.Logger,
	poolSizeWorker int,
) DeleteTransactionById[TxType] {
	return &defaultDeleteTransactionByAccountIdUseCase[TxType]{
		persistentRepo: persistentRepo,
		cacheRepo:      cacheRepo,
		logger:         logger,
		pool:           repo_pool_async.NewPool(context.TODO(), poolSizeWorker, logger),
	}
}

func (d *defaultDeleteTransactionByAccountIdUseCase[TxType]) Execute(ctx context.Context, req *DeleteReq, tx TxType) (
	*aggregate.TransactionByDetails,
	*repo_pool_async.Job,
	error,
) {

	detail, err := d.persistentRepo.GetById(ctx, req.TransactionId)
	if err != nil {
		// not found transaction with id
		return nil, nil, err
	}

	err = d.persistentRepo.Delete(ctx, &models.Transaction{ID: detail.Id}, tx)
	if err != nil {
		return nil, nil, err
	}

	asyncDeleteJob := d.pool.PushPriority(ctx, func(ctx context.Context) {
		d.logger.Info("Delete Transaction From Cache [cacheRepo.Delete(ctx, detail.Id]")
		_ = d.cacheRepo.Delete(ctx, detail.Id)
	})

	return detail, asyncDeleteJob, nil
}
