package mysql

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/models"
	"money_forward_code_challenge/internal/domain/transaction/repo"

	"gorm.io/gorm"
)

type mysqlUserRepoImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func (m *mysqlUserRepoImpl) CreateUser(ctx context.Context, user_model *models.User, tx *gorm.DB) error {
	defaultTx := m.db
	if tx != nil {
		defaultTx = tx
	}
	return defaultTx.WithContext(ctx).Create(user_model).Error
}

func (m *mysqlUserRepoImpl) CreateAccount(ctx context.Context, account_model *models.Account, tx *gorm.DB) error {
	defaultTx := m.db
	if tx != nil {
		defaultTx = tx
	}
	return defaultTx.WithContext(ctx).Create(account_model).Error
}

func (m *mysqlUserRepoImpl) UpdateAccount(ctx context.Context, account_model *models.Account, tx *gorm.DB) error {
	defaultTx := m.db
	if tx != nil {
		defaultTx = tx
	}
	return defaultTx.WithContext(ctx).Save(account_model).Error
}

func (m *mysqlUserRepoImpl) DeleteAccountById(ctx context.Context, account_id uint32, tx *gorm.DB) error {
	defaultTx := m.db
	if tx != nil {
		defaultTx = tx
	}
	return defaultTx.WithContext(ctx).Where("id = ?", account_id).Delete(&models.Account{}).Error
}

func (m *mysqlUserRepoImpl) GetUserById(ctx context.Context, user_id uint32) (*models.User, error) {
	var userModel models.User
	m.logger.Info("[MYSQLUserRepo-GET-USER]", zap.Uint32("user_id", user_id))
	err := m.db.WithContext(ctx).
		Select(models.USERCOLUMN_ID).
		Table(models.USERTABLE).
		Where(fmt.Sprintf("%s = ?", models.USERCOLUMN_ID), user_id).
		Find(&userModel).Error
	if err != nil {
		m.logger.Info("[MYSQLUserRepo-GET-USER]", zap.String("Error", err.Error()))
		return nil, err
	}

	if userModel.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	m.logger.Info("[MYSQLUserRepo-GET-USER]", zap.Any("Found", userModel))
	return &userModel, err
}

func (m *mysqlUserRepoImpl) GetAccountByAccountId(ctx context.Context, account_id uint32) (*aggregate.AccountByDetails, error) {
	var account aggregate.AccountByDetails
	err := m.db.WithContext(ctx).
		Select(models.ACCOUNTCOLUMN_ID,
			models.ACCOUNTCOLUMN_USER_ID,
			models.ACCOUNTCOLUMN_BALANCE,
			models.ACCOUNTCOLUMN_CREATED_AT,
			models.ACCOUNTCOLUMN_UPDATED_AT,
			models.ACCOUNTCOLUMN_BANK,
		).
		Where(fmt.Sprintf("%s = ?", models.ACCOUNTCOLUMN_ID), account_id).
		Table("accounts").
		Find(&account).Error
	if err != nil {
		m.logger.Info("[MYSQLUserRepo-GET-ACCOUNT]", zap.String("Error", err.Error()))
		return nil, err
	}

	if account.Id == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &account, nil
}

func (m *mysqlUserRepoImpl) BeginTx() *gorm.DB {
	return m.db.Begin()
}

func (m *mysqlUserRepoImpl) UpdateBalance(ctx context.Context, account_id uint32, new_balance float32, tx *gorm.DB) error {
	err := m.db.WithContext(ctx).
		Table(models.ACCOUNTTABLE).
		Where(fmt.Sprintf("%s = ?", models.ACCOUNTCOLUMN_ID), account_id).
		Update(models.ACCOUNTCOLUMN_BALANCE, new_balance).Error

	if err != nil {
		return err
	}

	return nil
}

func NewMysqlUserRepo(db *gorm.DB, logger *zap.Logger) repo.UserRepo[*gorm.DB] {
	return &mysqlUserRepoImpl{
		db:     db,
		logger: logger,
	}
}
