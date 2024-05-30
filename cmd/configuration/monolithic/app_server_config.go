package monolithic

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"money_forward_code_challenge/internal/domain/transaction/models"
	"os"
)

type AppConfigServer struct {
	logger      *zap.Logger
	gormDB      *gorm.DB
	redisDB     *redis.Client
	Environment string
	server      *gin.Engine
}

func (a *AppConfigServer) CreateGormMysqlDB() error {
	a.logger.Info("[AppConfigServer-CreateGormMysqlDB]", zap.String("Environment", a.Environment))
	dbName := os.Getenv("MYSQL_DATABASE")
	dbUser := os.Getenv("MYSQL_USER")
	dbPass := os.Getenv("MYSQL_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPass, dbHost, dbPort, dbName)

	a.logger.Info("[AppConfigServer-CreateGormMysqlDB]", zap.String("EnvironmentMysqlAddr", dsn))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	a.gormDB = db
	return nil
}

func (a *AppConfigServer) SetLogger(logger *zap.Logger) {
	a.logger = logger
}
func (a *AppConfigServer) CreateRedisDB() error {
	a.logger.Info("[AppConfigServer-CreateRedisDB]", zap.String("ConnectRedisDB", "Success"))
	redisAddr := os.Getenv("REDIS_ADDR")
	a.logger.Info("[AppConfigServer-CreateRedisDB]", zap.String("EnvironmentRedisAddr", redisAddr))
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
	})

	_, err := client.Ping(client.Context()).Result()
	if err != nil {
		a.logger.Error(err.Error())
		return err
	}

	a.redisDB = client
	return nil
}

func (a *AppConfigServer) InitDB() {
	err := a.gormDB.AutoMigrate(&models.User{}, &models.Account{}, &models.Transaction{})
	if err != nil {
		a.logger.Error(err.Error())
	}

	user := &models.User{
		ID:        1,
		FirstName: "THAI",
		LastName:  "LE",
	}
	err = a.gormDB.Create(user).Error

	if err != nil {
		a.logger.Error(err.Error())
		panic(err)
	}
	accounts := []*models.Account{
		&models.Account{
			ID:      1,
			UserId:  user.ID,
			Balance: 10000,
			Bank:    "VIB",
		},
		&models.Account{
			ID:      2,
			UserId:  user.ID,
			Balance: 500000,
			Bank:    "ACB",
		},

		&models.Account{
			ID:      3,
			UserId:  user.ID,
			Balance: 300000,
			Bank:    "VCB",
		},
	}

	err = a.gormDB.CreateInBatches(accounts, 1).Error
	if err != nil {
		a.logger.Error(err.Error())
		panic(err)
	}
	a.logger.Info("[AppConfigServer-InitDB]", zap.String("InitDB", "Success"))
}

func InitAppConfigServer() {

	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	appServerConfig := &AppConfigServer{
		Environment: os.Getenv("ENVIRONMENT"), // from docker-compose
	}
	appServerConfig.SetLogger(zapLogger)
	err = appServerConfig.CreateGormMysqlDB()
	if err != nil {
		panic(err)
	}
	err = appServerConfig.CreateRedisDB()

	if err != nil {
		panic(err)
	}

	appServerConfig.server = gin.Default()
	apiGroup := appServerConfig.server.Group("/api")
	userGroup := apiGroup.Group("/users/:id")
	transactionGroup := userGroup.Group("/transactions")
	InitTransactionRouter(appServerConfig.logger, transactionGroup, appServerConfig)
	appServerConfig.server.Run(":8080")
}
