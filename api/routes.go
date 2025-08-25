package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	taskRoutes := r.Group("/tasks")
	{
		taskRoutes.POST("", AddTask)
		taskRoutes.GET("", GetTasks)
		taskRoutes.PATCH("/:id", MarkTaskDone)
		taskRoutes.DELETE("/:id", DeleteTask)
		r.GET("/metrices", GetMetrics)
	}
}
