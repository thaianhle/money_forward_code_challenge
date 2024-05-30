package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/repo"
	data_provider_conversion "money_forward_code_challenge/internal/infrastructure/data-provider/data-provider-conversion"
)

type redisTransactionCacheRepoImpl struct {
	client               *redis.Client
	transactionDetailKey string
	logger               *zap.Logger
}

func (t *redisTransactionCacheRepoImpl) Set(ctx context.Context, details *aggregate.TransactionByDetails) error {
	KeyId := fmt.Sprintf("%s", details.Id)
	buf, err := data_provider_conversion.SerializeGOB[*aggregate.TransactionByDetails](details)

	if err != nil {
		return err
	}

	return t.client.HSet(ctx, t.transactionDetailKey, KeyId, buf.String()).Err()
}

func (t *redisTransactionCacheRepoImpl) GetById(ctx context.Context, id uint32) (*aggregate.TransactionByDetails, error) {
	keyId := fmt.Sprintf("%s", id)

	bufString := t.client.HGet(ctx, t.transactionDetailKey, keyId).Val()
	transactionDetail, err := data_provider_conversion.DeserializeGOB[*aggregate.TransactionByDetails](&bufString)
	if err != nil {
		return nil, err
	}

	return transactionDetail, nil
}

func (t *redisTransactionCacheRepoImpl) Delete(ctx context.Context, transactionId uint32) error {
	id := fmt.Sprintf("%s", transactionId)

	return t.client.HDel(ctx, t.transactionDetailKey, id).Err()
}

func (t *redisTransactionCacheRepoImpl) GetByAccountId(ctx context.Context, accountId uint32, query *repo.Query) ([]*aggregate.TransactionByDetails, error) {
	//TODO implement me
	// no implement because it very change frequently in persistent DB
	return nil, nil
}

func NewRedisTransactionCacheRepo(client *redis.Client, logger *zap.Logger) repo.TransactionCacheRepo {
	return &redisTransactionCacheRepoImpl{
		client:               client,
		transactionDetailKey: "transactions_detail",
		logger:               logger,
	}
}
