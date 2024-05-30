package redis

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/models"
	"money_forward_code_challenge/internal/domain/transaction/repo"
	data_provider_conversion "money_forward_code_challenge/internal/infrastructure/data-provider/data-provider-conversion"

	"github.com/go-redis/redis/v8"
)

type redisUserCacheRepoImpl struct {
	db               *redis.Client
	accountDetailKey string
	logger           *zap.Logger
}

func NewRedisUserCacheRepo(db *redis.Client, logger *zap.Logger) repo.UserCacheRepo {
	return &redisUserCacheRepoImpl{
		db: db, accountDetailKey: "accounts",
		logger: logger,
	}
}
func (r *redisUserCacheRepoImpl) CreateUser(ctx context.Context, user *models.User) error {
	//TODO implement me
	panic("implement me")
}

func (r *redisUserCacheRepoImpl) SetAccount(ctx context.Context, details *aggregate.AccountByDetails) error {
	keyId := fmt.Sprintf("%s", details.Id)
	buf, err := data_provider_conversion.SerializeGOB[*aggregate.AccountByDetails](details)
	if err != nil {
		return err
	}

	err = r.db.HSet(ctx, r.accountDetailKey, keyId, buf.String()).Err()

	return err
}

func (r *redisUserCacheRepoImpl) DeleteAccountById(ctx context.Context, id uint32) error {
	keyId := fmt.Sprintf("%s", id)

	return r.db.HDel(ctx, r.accountDetailKey, keyId).Err()
}

func (r *redisUserCacheRepoImpl) GetUserById(ctx context.Context, user_id uint32) (*models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (r *redisUserCacheRepoImpl) GetAccountByAccountId(ctx context.Context, account_id uint32) (*aggregate.AccountByDetails, error) {
	keyId := fmt.Sprintf("%s", account_id)

	bufString := r.db.HGet(ctx, r.accountDetailKey, keyId).Val()
	if bufString == "" {
		return nil, fmt.Errorf("account not found")
	}
	accountDetail, err := data_provider_conversion.DeserializeGOB[*aggregate.AccountByDetails](&bufString)
	if err != nil {
		return nil, err
	}

	return accountDetail, nil
}
