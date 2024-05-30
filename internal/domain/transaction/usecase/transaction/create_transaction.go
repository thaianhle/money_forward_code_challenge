package transaction

import (
	"context"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/models"
	"money_forward_code_challenge/internal/domain/transaction/repo"
	"money_forward_code_challenge/pkgs/repo_pool_async"
)

type CreateReq struct {
	// userId, BankTypeName not required in json binding
	// because in url api
	// it be checked
	UserId          uint32  `json:"user_id"`
	BankType        string  `json:"bank_type"`
	AccountId       uint32  `json:"account_id,binding:required"`
	Amount          float32 `json:"amount,binding:required"`
	TransactionType string  `json:"transaction_type,binding:required"`
}

type CreateUseCase[TxType any] interface {
	Execute(ctx context.Context, req *CreateReq, tx TxType) (
		*aggregate.TransactionByDetails,
		*repo_pool_async.Job,
		error)
}

type defaultCreateUseCase[TxType any] struct {
	persistentRepo repo.TransactionRepo[TxType]
	cacheRepo      repo.TransactionCacheRepo
	pool           *repo_pool_async.RepoUpdatePoolBusyWaiting
	logger         *zap.Logger
}

func NewCreateUseCase[TxType any](persistentRepo repo.TransactionRepo[TxType],
	cacheRepo repo.TransactionCacheRepo, logger *zap.Logger,
	poolSizeWorker int,
) CreateUseCase[TxType] {
	return &defaultCreateUseCase[TxType]{
		logger:         logger,
		persistentRepo: persistentRepo,
		cacheRepo:      cacheRepo,
		pool:           repo_pool_async.NewPool(context.TODO(), poolSizeWorker, logger),
		//pool:           repo_pool_async.NewRepoUpdatePoolBusyWaiting[repo.TransactionCacheRepo, *aggregate.TransactionByDetails](),
	}
}

func (d *defaultCreateUseCase[TxType]) Execute(ctx context.Context, req *CreateReq, tx TxType) (
	*aggregate.TransactionByDetails,
	*repo_pool_async.Job,
	error) {

	transactionModel := &models.Transaction{
		AccountID:       req.AccountId,
		Amount:          req.Amount,
		TransactionType: req.TransactionType,
	}

	err := d.persistentRepo.Create(ctx, transactionModel, tx)
	if err != nil {
		return nil, nil, err
	}

	// then this create detail transaction for pool handler will update redis
	// if have response update async before 300ms

	details := &aggregate.TransactionByDetails{
		Id:              transactionModel.ID,
		UserId:          req.UserId,
		AccountId:       req.AccountId,
		Amount:          req.Amount,
		TransactionType: req.TransactionType,
		Bank:            req.BankType,
		CreatedAt:       transactionModel.CreatedAt.String(),
	}

	_ = details.FormatDateHCM()
	asyncUpdate := d.pool.PushPriority(ctx, func(ctx context.Context) {
		_ = d.cacheRepo.Set(ctx, details)
	})
	return details, asyncUpdate, nil
}
