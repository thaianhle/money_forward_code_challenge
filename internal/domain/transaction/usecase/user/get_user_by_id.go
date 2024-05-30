package user

import (
	"context"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/domain/transaction/models"
	"money_forward_code_challenge/internal/domain/transaction/repo"
)

type GetUserByIdReq struct {
	UserId uint32
}

type GetUserById[TxType any] interface {
	Execute(ctx context.Context, req *GetUserByIdReq) (*models.User, error)
}

type GetUserByIdUseCase[TxType any] struct {
	persistentRepo repo.UserRepo[TxType]
	cacheRepo      repo.UserCacheRepo
	logger         *zap.Logger
}

func NewGetUserByIdUseCase[TxType any](persistentRepo repo.UserRepo[TxType], cacheRepo repo.UserCacheRepo, logger *zap.Logger) GetUserById[TxType] {
	return &GetUserByIdUseCase[TxType]{
		persistentRepo: persistentRepo, cacheRepo: cacheRepo,
		logger: logger,
	}
}

func (u *GetUserByIdUseCase[TxType]) Execute(ctx context.Context, req *GetUserByIdReq) (*models.User, error) {
	return u.persistentRepo.GetUserById(ctx, req.UserId)
}
