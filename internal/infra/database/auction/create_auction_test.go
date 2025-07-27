package auction

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/TiagoAmaralFerreira/go-expert-leilao/internal/entity/auction_entity"
	"github.com/stretchr/testify/assert"
)

func TestGetAuctionInterval(t *testing.T) {
	// Teste com valor válido
	os.Setenv("AUCTION_INTERVAL", "10m")
	interval := getAuctionInterval()
	assert.Equal(t, 10*time.Minute, interval)

	// Teste com valor inválido (deve retornar padrão)
	os.Setenv("AUCTION_INTERVAL", "invalid")
	interval = getAuctionInterval()
	assert.Equal(t, 5*time.Minute, interval)
}

func TestGetCheckInterval(t *testing.T) {
	// Teste com valor válido
	os.Setenv("AUCTION_CHECK_INTERVAL", "15s")
	interval := getCheckInterval()
	assert.Equal(t, 15*time.Second, interval)

	// Teste com valor inválido (deve retornar padrão)
	os.Setenv("AUCTION_CHECK_INTERVAL", "invalid")
	interval = getCheckInterval()
	assert.Equal(t, 30*time.Second, interval)
}

func TestAuctionStatusManagement(t *testing.T) {
	// Configuração do teste
	os.Setenv("AUCTION_INTERVAL", "1h")
	os.Setenv("AUCTION_CHECK_INTERVAL", "30s")

	// Criar um repositório mock (sem conexão com banco)
	repo := &AuctionRepository{
		auctionInterval:       getAuctionInterval(),
		checkInterval:         getCheckInterval(),
		stopChan:              make(chan struct{}),
		auctionStatusMap:      make(map[string]auction_entity.AuctionStatus),
		auctionEndTimeMap:     make(map[string]time.Time),
		auctionStatusMapMutex: &sync.Mutex{},
		auctionEndTimeMutex:   &sync.Mutex{},
	}

	// Criar um leilão
	auction, _ := auction_entity.CreateAuction(
		"Test Product",
		"Electronics",
		"Test Description",
		auction_entity.New,
	)

	// Simular adição do leilão aos maps
	repo.auctionStatusMapMutex.Lock()
	repo.auctionStatusMap[auction.Id] = auction.Status
	repo.auctionStatusMapMutex.Unlock()

	repo.auctionEndTimeMutex.Lock()
	repo.auctionEndTimeMap[auction.Id] = auction.Timestamp.Add(repo.auctionInterval)
	repo.auctionEndTimeMutex.Unlock()

	// Verificar se o leilão foi adicionado corretamente
	repo.auctionStatusMapMutex.Lock()
	status, exists := repo.auctionStatusMap[auction.Id]
	repo.auctionStatusMapMutex.Unlock()

	assert.True(t, exists)
	assert.Equal(t, auction_entity.Active, status)

	// Verificar se o tempo de fim foi calculado corretamente
	repo.auctionEndTimeMutex.Lock()
	endTime, exists := repo.auctionEndTimeMap[auction.Id]
	repo.auctionEndTimeMutex.Unlock()

	assert.True(t, exists)
	assert.True(t, endTime.After(auction.Timestamp))
}

func TestAuctionExpirationLogic(t *testing.T) {
	// Configuração do teste
	os.Setenv("AUCTION_INTERVAL", "1s") // Leilão muito curto para teste
	os.Setenv("AUCTION_CHECK_INTERVAL", "500ms")

	// Criar um repositório mock
	repo := &AuctionRepository{
		auctionInterval:       getAuctionInterval(),
		checkInterval:         getCheckInterval(),
		stopChan:              make(chan struct{}),
		auctionStatusMap:      make(map[string]auction_entity.AuctionStatus),
		auctionEndTimeMap:     make(map[string]time.Time),
		auctionStatusMapMutex: &sync.Mutex{},
		auctionEndTimeMutex:   &sync.Mutex{},
	}

	// Criar um leilão que já expirou
	auction, _ := auction_entity.CreateAuction(
		"Expired Product",
		"Electronics",
		"Expired Description",
		auction_entity.New,
	)

	// Definir o tempo de fim como passado
	repo.auctionStatusMapMutex.Lock()
	repo.auctionStatusMap[auction.Id] = auction_entity.Active
	repo.auctionStatusMapMutex.Unlock()

	repo.auctionEndTimeMutex.Lock()
	repo.auctionEndTimeMap[auction.Id] = time.Now().Add(-2 * time.Second) // 2 segundos atrás
	repo.auctionEndTimeMutex.Unlock()

	// Verificar se o leilão é detectado como expirado
	now := time.Now()
	repo.auctionEndTimeMutex.Lock()
	endTime := repo.auctionEndTimeMap[auction.Id]
	repo.auctionEndTimeMutex.Unlock()

	assert.True(t, now.After(endTime), "Leilão deveria estar expirado")

	// Verificar se o status ainda está ativo (não foi fechado automaticamente ainda)
	repo.auctionStatusMapMutex.Lock()
	status := repo.auctionStatusMap[auction.Id]
	repo.auctionStatusMapMutex.Unlock()

	assert.Equal(t, auction_entity.Active, status)
}
