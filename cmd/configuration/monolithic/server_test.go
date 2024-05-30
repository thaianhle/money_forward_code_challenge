package monolithic

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	server := gin.Default()
	apiGroup := server.Group("/api")
	userGroup := apiGroup.Group("/users/:user_id")
	transactionGroup := userGroup.Group("/transactions")
	transactionGroup.GET("/", func(context *gin.Context) {
		type Query struct {
			AccountId uint32 `form:"account_id" binding:"min=1"`
			Limit     uint32 `form:"limit"`
			Offset    uint32 `form:"offset"`
		}

		var query Query
		err := context.BindQuery(&query)
		accountQueryValue := context.Query("account_id")
		if err != nil && accountQueryValue != "" {
			context.JSON(http.StatusBadRequest, &gin.H{
				"error": err.Error(),
			})
			return
		}
		context.JSON(http.StatusOK, &gin.H{
			"query_options": query,
		})
	})
	server.Run()
}
