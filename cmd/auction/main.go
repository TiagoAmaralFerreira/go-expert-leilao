package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/TiagoAmaralFerreira/go-expert-leilao/configuration/database/mongodb"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/infra/api/web/controller/auction_controller"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/infra/api/web/controller/bid_controller"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/infra/api/web/controller/user_controller"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/infra/database/auction"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/infra/database/bid"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/infra/database/user"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/usecase/auction_usecase"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/usecase/bid_usecase"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/usecase/user_usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load("cmd/auction/.env"); err != nil {
		log.Fatal("Error trying to load env variables")
		return
	}

	databaseConnection, err := mongodb.NewMongoDBConnection(ctx)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	router := gin.Default()

	userController, bidController, auctionsController, auctionRepository := initDependencies(databaseConnection)

	router.GET("/auction", auctionsController.FindAuctions)
	router.GET("/auction/:auctionId", auctionsController.FindAuctionById)
	router.POST("/auction", auctionsController.CreateAuction)
	router.GET("/auction/winner/:auctionId", auctionsController.FindWinningBidByAuctionId)
	router.POST("/bid", bidController.CreateBid)
	router.GET("/bid/:auctionId", bidController.FindBidByAuctionId)
	router.GET("/user/:userId", userController.FindUserById)

	// Configuração para graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		router.Run(":8080")
	}()

	// Aguarda sinal de interrupção
	<-sigChan
	log.Println("Shutting down gracefully...")

	// Para o repositório de leilões
	auctionRepository.Stop()

	log.Println("Server stopped")
}

func initDependencies(database *mongo.Database) (
	userController *user_controller.UserController,
	bidController *bid_controller.BidController,
	auctionController *auction_controller.AuctionController,
	auctionRepository *auction.AuctionRepository) {

	auctionRepository = auction.NewAuctionRepository(database)
	bidRepository := bid.NewBidRepository(database, auctionRepository)
	userRepository := user.NewUserRepository(database)

	userController = user_controller.NewUserController(
		user_usecase.NewUserUseCase(userRepository))
	auctionController = auction_controller.NewAuctionController(
		auction_usecase.NewAuctionUseCase(auctionRepository, bidRepository))
	bidController = bid_controller.NewBidController(bid_usecase.NewBidUseCase(bidRepository))

	return
}
