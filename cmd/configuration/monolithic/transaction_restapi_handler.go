package monolithic

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/common/composite"
	"money_forward_code_challenge/internal/common/httpresponse"
	"money_forward_code_challenge/internal/domain/transaction/repo"
	"money_forward_code_challenge/internal/domain/transaction/usecase/transaction"
	transactionusecase "money_forward_code_challenge/internal/domain/transaction/usecase/transaction"
	"money_forward_code_challenge/internal/infrastructure/data-provider/mysql"
	"money_forward_code_challenge/internal/infrastructure/data-provider/redis"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var ContextUserIdKey string = "user-id"

type TransactionHandler struct {
	routerGroup     *gin.RouterGroup
	appServerConfig *AppConfigServer
	service         *TransactionService
	logger          *zap.Logger
}

func InitTransactionRouter(logger *zap.Logger, routerGroup *gin.RouterGroup, appServerConfig *AppConfigServer) {
	t := &TransactionHandler{
		routerGroup:     routerGroup,
		appServerConfig: appServerConfig,
		logger:          logger,
	}
	transactionRepoComposite := &composite.TransactionRepoComposite{
		PersistentRepo: mysql.NewMysqlTransactionRepo(appServerConfig.gormDB, logger),
		CacheRepo:      redis.NewRedisTransactionCacheRepo(appServerConfig.redisDB, logger),
	}

	userRepoComposite := &composite.UserRepoComposite{
		PersistentRepo: mysql.NewMysqlUserRepo(appServerConfig.gormDB, logger),
		CacheRepo:      redis.NewRedisUserCacheRepo(appServerConfig.redisDB, logger),
	}

	t.service = NewTransactionService(transactionRepoComposite, userRepoComposite, logger, 10)
	t.InitRouter()
}

func (t *TransactionHandler) InitRouter() {
	t.routerGroup.POST("/", t.createTransactionByUser)   // 1 api
	t.routerGroup.GET("/", t.getTransactions)            // 2 api in one
	t.routerGroup.DELETE("/", t.deleteTransactionByUser) // 1 api

	// TODO
	// I need discuss a little
	// Should Update
	// Or make one transaction new
	// because I think , transaction should be immutable
	// to use for event stream or any thing
	// have meaning about history
	// t.routerGroup.PUT("/", t.updateTransactionByUser) // 1 api
}

func (t *TransactionHandler) createTransactionByUser(ginCtx *gin.Context) {
	userIdParam, err := getUserIdURLParam(ginCtx, "id")

	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, &gin.H{
			"error": err.Error(),
		})
		return
	}

	setUserIdToContext(ginCtx, userIdParam)
	var req transaction.CreateReq
	err = ginCtx.ShouldBindJSON(&req)

	response := t.service.createTransactionByUser(ginCtx, &req)
	ginCtx.JSON(http.StatusOK, response)
}

func (t *TransactionHandler) getTransactions(ginCtx *gin.Context) {
	type QueryOption struct {
		AccountId uint32 `form:"account_id"`
		Limit     int    `form:"limit"`
		Offset    int    `form:"offset"`
	}

	userIdParam, err := getUserIdURLParam(ginCtx, "id")
	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, &gin.H{
			"error": err.Error(),
		})
		return
	}

	var queryOption QueryOption
	err = ginCtx.ShouldBindQuery(&queryOption)
	accountIdValue := ginCtx.Value("account_id")
	getByUserId := accountIdValue == nil
	// when parse query option
	// have two case
	// 1. account_id == 0 (it not valid because I get min must be 1)
	// 2. client don't enter one account_id (mean that query transaction with user_id)
	// but case 2 is passed
	if err != nil && !getByUserId {
		ginCtx.JSON(http.StatusBadRequest, &gin.H{
			"error": err.Error(),
		})
		return
	}

	if accountIdValue == "0" {
		ginCtx.JSON(http.StatusBadRequest, &gin.H{
			"error": "account_id must be greater than 0",
		})
		return
	}

	if queryOption.Limit == queryOption.Offset {
		queryOption.Limit = 10
		queryOption.Offset = 0
	}

	setUserIdToContext(ginCtx, userIdParam)
	var response *httpresponse.Response
	repoQuery := &repo.Query{
		Limit:  queryOption.Limit,
		Offset: queryOption.Offset,
	}
	if getByUserId {
		req := &transactionusecase.GetTransactionByUserIdReq{
			UserId: userIdParam,
			Query:  repoQuery,
		}

		response = t.service.getTransactionsByUserId(ginCtx, req)
	} else {
		req := &transactionusecase.GetTransactionByAccountIdReq{
			AccountId: queryOption.AccountId,
			Query:     repoQuery,
		}
		fmt.Println("GET TRANSACTIONS BY ACCOUNT ID: ", *req)
		response = t.service.getTransactionsByAccountId(ginCtx, req)
	}

	ginCtx.JSON(response.Code, response)
}

func (t *TransactionHandler) deleteTransactionByUser(ginCtx *gin.Context) {
	userIdParam, err := getUserIdURLParam(ginCtx, "id")

	if err != nil {
		ginCtx.JSON(http.StatusBadRequest, &gin.H{
			"error": err.Error(),
		})
		return
	}

	setUserIdToContext(ginCtx, userIdParam)
	var req transaction.DeleteReq
	err = ginCtx.ShouldBindJSON(&req)

	response := t.service.deleteTransactionByUser(ginCtx, &req)
	ginCtx.JSON(http.StatusOK, response)
}

func getUserIdURLParam(c *gin.Context, nameParam string) (uint32, error) {
	userIdParam := c.Param(nameParam)
	userIdInt, err := strconv.Atoi(userIdParam)
	if err != nil {
		return 0, err
	}

	if userIdInt == 0 {
		// I want match error from here
		// without process any user_id = 0
		return 0, fmt.Errorf("user_id is fake = 0, not valid")
	}
	return uint32(userIdInt), nil
}

func getTransactionIdQueryParam(c *gin.Context, nameParam string) (uint32, error) {
	transactionIdParam := c.Query(nameParam)
	transactionIdInt, err := strconv.Atoi(transactionIdParam)
	if err != nil {
		return 0, err
	}
	return uint32(transactionIdInt), nil
}

func setUserIdToContext(ginCtx *gin.Context, userId uint32) {
	ginCtx.Set(ContextUserIdKey, userId)
}

func getUserIdFromContext(ctx context.Context) uint32 {
	return ctx.Value(ContextUserIdKey).(uint32)
}
