package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"infotecs/internal/handlers"
	"infotecs/internal/repository"
	"infotecs/internal/service"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	postgresDatabase := repository.NewPostgresRepository(os.Getenv("PQ_USERNAME"),
		os.Getenv("PQ_PASSWORD"),
		os.Getenv("PQ_HOST"),
		os.Getenv("PQ_PORT"),
		os.Getenv("PQ_DBNAME"),
	)
	err = postgresDatabase.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer postgresDatabase.Close()

	err = postgresDatabase.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	service.Service.InjectRepository(postgresDatabase)

	router := gin.Default()
	router.POST("/api/v1/wallet", handlers.PostNewWallet)
	router.POST("/api/v1/wallet/:id/send", handlers.PostSendMoney)
	router.GET("/api/v1/wallet/:id/history", handlers.GetTransactionHistory)
	router.GET("/api/v1/wallet/:id", handlers.GetWallet)
	router.Run(os.Getenv("SERVER_PORT"))
}
