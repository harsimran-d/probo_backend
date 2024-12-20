package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"probo_backend/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var SigningSecret string

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	postgresUrl := os.Getenv("POSTGRES_URL")
	if postgresUrl == "" {
		log.Fatal("error loading postgresUrl")
	}
	pool, err := pgxpool.New(context.Background(), postgresUrl)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer pool.Close()

	userHandlers := handlers.NewUserHandler(pool)
	proboHandler := handlers.NewProboHandler()

	router := gin.Default()
	router.Use(RejectTrailingSpaces())
	router.POST("/signup", userHandlers.Signup)
	router.POST("/signin", userHandlers.Signin)

	router.POST("/user/create/:userId", proboHandler.CreateUser)
	router.POST("/market/create/:marketId", proboHandler.CreateMarket)
	router.GET("/orderbook/:marketId", proboHandler.GetOrderBook)
	router.GET("/balances/inr", proboHandler.GetInrBalances)
	router.GET("/balances/stock", proboHandler.GetStockBalances)
	router.POST("/reset", proboHandler.Reset)

	router.Use(authMiddleware())
	router.Group("/")
	{
		router.POST("", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"message": "hello world"})
		})
	}
	router.Run(":3000")
}
