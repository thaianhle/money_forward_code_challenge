package mysql

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"log"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/models"
	"money_forward_code_challenge/internal/domain/transaction/repo"

	"gorm.io/gorm"
)

type mysqlTransactionRepoImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewMysqlTransactionRepo(db *gorm.DB, logger *zap.Logger) repo.TransactionRepo[*gorm.DB] {
	return &mysqlTransactionRepoImpl{
		db:     db,
		logger: logger,
	}
}

func (r *mysqlTransactionRepoImpl) GetById(ctx context.Context, id uint32) (*aggregate.TransactionByDetails, error) {
	var transaction aggregate.TransactionByDetails
	err := r.db.WithContext(ctx).
		Select(models.TRANSACTIONCOLUMN_ID,
			models.TRANSACTIONCOLUMN_CREATED_AT,
			models.TRANSACTIONCOLUMN_TRANSACTION_TYPE,
			models.TRANSACTIONCOLUMN_AMOUNT,
			models.TRANSACTIONCOLUMN_ACCOUNT_ID,
			models.ACCOUNTCOLUMN_BANK,
			models.ACCOUNTCOLUMN_USER_ID,
		).
		Joins(fmt.Sprintf("INNER JOIN %s ON %s = %s",
			models.ACCOUNTTABLE,
			models.TRANSACTIONCOLUMN_ACCOUNT_ID, // transactions.account_id
			models.ACCOUNTCOLUMN_ID),            // accounts.id
		).Where(fmt.Sprintf("%s = ?", models.TRANSACTIONCOLUMN_ID), id).
		Table(models.TRANSACTIONTABLE).
		Find(&transaction).Error

	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *mysqlTransactionRepoImpl) GetByUserId(ctx context.Context, user_id uint32, query *repo.Query) ([]*aggregate.TransactionByDetails, error) {
	var transactions []*aggregate.TransactionByDetails
	log.Println("[repo] GetByUserId: ", user_id)
	builder := r.db.WithContext(ctx).
		Select(models.TRANSACTIONCOLUMN_ID,
			models.TRANSACTIONCOLUMN_CREATED_AT,
			models.TRANSACTIONCOLUMN_TRANSACTION_TYPE,
			models.TRANSACTIONCOLUMN_AMOUNT,
			models.TRANSACTIONCOLUMN_ACCOUNT_ID,
			models.ACCOUNTCOLUMN_BANK,
			models.ACCOUNTCOLUMN_USER_ID,
		).
		Joins(fmt.Sprintf("INNER JOIN %s ON %s = %s",
			models.ACCOUNTTABLE,
			models.TRANSACTIONCOLUMN_ACCOUNT_ID, // transactions.account_id
			models.ACCOUNTCOLUMN_ID),            // accounts.id
		).
		Limit(query.Limit).
		Offset(query.Offset).
		Where(fmt.Sprintf("%s = ? AND %s = ?", models.ACCOUNTCOLUMN_USER_ID, models.TRANSACTIONCOLUMN_DELETED), user_id, false).
		Table(models.TRANSACTIONTABLE)
	if query.SortBy != "" {
		builder.Order(query.SortBy + " " + query.Order)
	} else {
		builder.Order("transactions.id DESC ")
	}

	err := builder.Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(transactions); i++ {
		err := transactions[i].FormatDateHCM()
		if err != nil {
			fmt.Println("parse error layout: ", err)
		}
	}
	return transactions, nil
}

func (r *mysqlTransactionRepoImpl) GetByAccountId(ctx context.Context, account_id uint32, query *repo.Query) ([]*aggregate.TransactionByDetails, error) {
	var transactions []*aggregate.TransactionByDetails
	builder := r.db.WithContext(ctx).
		Select("transactions.*", "accounts.bank_type", "accounts.balance", "accounts.user_id").
		Select(models.TRANSACTIONCOLUMN_ID,
			models.TRANSACTIONCOLUMN_CREATED_AT,
			models.TRANSACTIONCOLUMN_TRANSACTION_TYPE,
			models.TRANSACTIONCOLUMN_AMOUNT,
			models.TRANSACTIONCOLUMN_ACCOUNT_ID,
			models.ACCOUNTCOLUMN_BANK,
			models.ACCOUNTCOLUMN_USER_ID,
		).
		Joins(fmt.Sprintf("INNER JOIN %s ON %s = %s",
			models.ACCOUNTTABLE,
			models.TRANSACTIONCOLUMN_ACCOUNT_ID,
			models.ACCOUNTCOLUMN_ID)).
		Where(fmt.Sprintf("%s = ? AND %s = ?", models.ACCOUNTCOLUMN_ID, models.TRANSACTIONCOLUMN_DELETED), account_id, false).
		Order(models.TRANSACTIONCOLUMN_ID + " DESC")
	if query.Offset != 0 && query.Limit != 0 {
		builder.Limit(query.Limit).Offset(query.Offset)
	} else {
		builder.Limit(10).Offset(0)
	}

	err := builder.Table("transactions").
		Find(&transactions).Error

	for _, transactionDetail := range transactions {
		err := transactionDetail.FormatDateHCM()
		fmt.Println(err)
	}
	return transactions, err
}

func (r *mysqlTransactionRepoImpl) Create(ctx context.Context, transaction *models.Transaction, tx *gorm.DB) error {
	txDB := r.db
	if tx != nil {
		txDB = tx
	}
	return txDB.WithContext(ctx).Create(transaction).Error
}

func (r *mysqlTransactionRepoImpl) Update(ctx context.Context, transaction *models.Transaction, tx *gorm.DB) error {
	txDB := r.db
	if tx != nil {
		txDB = tx
	}
	return txDB.WithContext(ctx).Save(transaction).Error
}

func (r *mysqlTransactionRepoImpl) Delete(ctx context.Context, transaction *models.Transaction, tx *gorm.DB) error {
	r.logger.Info("[MYSQLTransactionRepo-DELETE-TRANSACTION]", zap.Any("transaction", transaction))
	txDB := r.db
	if tx != nil {
		txDB = tx
	}

	err := txDB.WithContext(ctx).Table(models.TRANSACTIONTABLE).
		Where(fmt.Sprintf("%s = ?", models.TRANSACTIONCOLUMN_ID), transaction.ID).
		Update(models.TRANSACTIONCOLUMN_DELETED, true).Error
	if err != nil {
		r.logger.Error("[MYSQLTransactionRepo-DELETE-TRANSACTION]", zap.Any("transaction", transaction))
		return err
	}

	r.logger.Info("[MYSQLTransactionRepo-DELETE-TRANSACTION]", zap.Any("transaction", transaction))
	return nil
}

func (r *mysqlTransactionRepoImpl) BeginTx() *gorm.DB {
	return r.db.Begin()
}
