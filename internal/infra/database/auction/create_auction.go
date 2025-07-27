package auction

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/TiagoAmaralFerreira/go-expert-leilao/configuration/logger"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/entity/auction_entity"
	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/internal_error"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}

type AuctionRepository struct {
	Collection            *mongo.Collection
	auctionInterval       time.Duration
	checkInterval         time.Duration
	stopChan              chan struct{}
	wg                    sync.WaitGroup
	auctionStatusMap      map[string]auction_entity.AuctionStatus
	auctionEndTimeMap     map[string]time.Time
	auctionStatusMapMutex *sync.Mutex
	auctionEndTimeMutex   *sync.Mutex
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	repo := &AuctionRepository{
		Collection:            database.Collection("auctions"),
		auctionInterval:       getAuctionInterval(),
		checkInterval:         getCheckInterval(),
		stopChan:              make(chan struct{}),
		auctionStatusMap:      make(map[string]auction_entity.AuctionStatus),
		auctionEndTimeMap:     make(map[string]time.Time),
		auctionStatusMapMutex: &sync.Mutex{},
		auctionEndTimeMutex:   &sync.Mutex{},
	}

	// Inicia a goroutine para verificar leilões vencidos
	repo.startAuctionCloser()

	return repo
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	// Adiciona o leilão aos maps para controle
	ar.auctionStatusMapMutex.Lock()
	ar.auctionStatusMap[auctionEntity.Id] = auctionEntity.Status
	ar.auctionStatusMapMutex.Unlock()

	ar.auctionEndTimeMutex.Lock()
	ar.auctionEndTimeMap[auctionEntity.Id] = auctionEntity.Timestamp.Add(ar.auctionInterval)
	ar.auctionEndTimeMutex.Unlock()

	return nil
}

func (ar *AuctionRepository) UpdateAuctionStatus(
	ctx context.Context,
	auctionId string,
	status auction_entity.AuctionStatus) *internal_error.InternalError {
	filter := bson.M{"_id": auctionId}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Error trying to update auction status", err)
		return internal_error.NewInternalServerError("Error trying to update auction status")
	}

	// Atualiza o map local
	ar.auctionStatusMapMutex.Lock()
	ar.auctionStatusMap[auctionId] = status
	ar.auctionStatusMapMutex.Unlock()

	return nil
}

func (ar *AuctionRepository) startAuctionCloser() {
	ar.wg.Add(1)
	go func() {
		defer ar.wg.Done()
		ticker := time.NewTicker(ar.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ar.checkAndCloseExpiredAuctions()
			case <-ar.stopChan:
				return
			}
		}
	}()
}

func (ar *AuctionRepository) checkAndCloseExpiredAuctions() {
	ctx := context.Background()
	now := time.Now()

	ar.auctionEndTimeMutex.Lock()
	auctionEndTimes := make(map[string]time.Time)
	for k, v := range ar.auctionEndTimeMap {
		auctionEndTimes[k] = v
	}
	ar.auctionEndTimeMutex.Unlock()

	ar.auctionStatusMapMutex.Lock()
	auctionStatuses := make(map[string]auction_entity.AuctionStatus)
	for k, v := range ar.auctionStatusMap {
		auctionStatuses[k] = v
	}
	ar.auctionStatusMapMutex.Unlock()

	for auctionId, endTime := range auctionEndTimes {
		if now.After(endTime) {
			status, exists := auctionStatuses[auctionId]
			if exists && status == auction_entity.Active {
				logger.Info("Closing expired auction: " + auctionId)
				if err := ar.UpdateAuctionStatus(ctx, auctionId, auction_entity.Completed); err != nil {
					logger.Error("Error closing auction: "+auctionId, err)
				}
			}
		}
	}
}

func (ar *AuctionRepository) Stop() {
	close(ar.stopChan)
	ar.wg.Wait()
}

func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5
	}
	return duration
}

func getCheckInterval() time.Duration {
	checkInterval := os.Getenv("AUCTION_CHECK_INTERVAL")
	duration, err := time.ParseDuration(checkInterval)
	if err != nil {
		return time.Second * 30
	}
	return duration
}
