package main

import (
    "student-to-do-list/backend/api"
    "student-to-do-list/backend/config"
    "student-to-do-list/backend/scheduler"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "log"
)

func main() {
    config.ConnectDatabase()    

    r := gin.Default()

    // Add CORS Middleware
    r.Use(cors.Default())

    api.SetupRoutes(r)

    go scheduler.InitScheduler()

    if err := r.Run(":8080"); err != nil {
        log.Fatalf("Failed to run server: %v", err)
    }
}
