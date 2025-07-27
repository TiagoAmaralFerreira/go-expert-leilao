# Sistema de Leilões com Fechamento Automático

Este projeto implementa um sistema de leilões em Go com funcionalidade de fechamento automático baseado em tempo configurável.

## Funcionalidades

- ✅ Criação de leilões
- ✅ Sistema de lances (bids)
- ✅ Fechamento automático de leilões
- ✅ Validação de leilões ativos/inativos
- ✅ API REST para todas as operações
- ✅ Persistência em MongoDB
- ✅ Concorrência com goroutines

## Arquitetura

O projeto segue a arquitetura hexagonal (Clean Architecture) com as seguintes camadas:

- **Entity**: Entidades de domínio (Auction, Bid, User)
- **UseCase**: Casos de uso da aplicação
- **Infra**: Implementações de infraestrutura (banco de dados, API)
- **Configuration**: Configurações globais (logger, database)

## Nova Funcionalidade: Fechamento Automático

### Implementação

A funcionalidade de fechamento automático foi implementada no arquivo `internal/infra/database/auction/create_auction.go` com os seguintes componentes:

1. **Goroutine de Verificação**: Executa periodicamente para verificar leilões vencidos
2. **Maps de Controle**: Mantém o status e tempo de fim dos leilões em memória
3. **Mutex para Concorrência**: Garante thread-safety nas operações
4. **Configuração via Variáveis de Ambiente**: Permite ajustar intervalos

### Variáveis de Ambiente

- `AUCTION_INTERVAL`: Duração do leilão (ex: "5m", "1h")
- `AUCTION_CHECK_INTERVAL`: Intervalo de verificação de leilões vencidos (ex: "30s", "1m")

## Como Executar

### Pré-requisitos

- Docker e Docker Compose
- Go 1.20 ou superior

### 1. Configuração do Ambiente

Crie um arquivo `.env` na pasta `cmd/auction/` com o seguinte conteúdo:

```env
MONGODB_URL=mongodb://mongodb:27017
MONGODB_DB=auction_db
AUCTION_INTERVAL=5m
AUCTION_CHECK_INTERVAL=30s
```

### 2. Execução com Docker

```bash
# Construir e executar os containers
docker-compose up --build

# Executar em background
docker-compose up -d --build
```

### 3. Execução Local

```bash
# Instalar dependências
go mod tidy

# Executar a aplicação
go run cmd/auction/main.go
```

## API Endpoints

### Leilões

- `GET /auction` - Listar leilões
- `GET /auction/:auctionId` - Buscar leilão por ID
- `POST /auction` - Criar novo leilão
- `GET /auction/winner/:auctionId` - Buscar lance vencedor

### Lances

- `POST /bid` - Criar novo lance
- `GET /bid/:auctionId` - Listar lances de um leilão

### Usuários

- `GET /user/:userId` - Buscar usuário por ID

## Testes

### Executar Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes específicos
go test ./internal/infra/database/auction -v

# Executar testes com cobertura
go test ./... -cover
```

### Testes de Fechamento Automático

O arquivo `internal/infra/database/auction/create_auction_test.go` contém testes que validam:

1. **TestAuctionAutoClose**: Verifica se leilões são fechados automaticamente após o tempo configurado
2. **TestAuctionManualClose**: Testa o fechamento manual de leilões
3. **TestGetAuctionInterval**: Valida a leitura das variáveis de ambiente
4. **TestGetCheckInterval**: Testa a configuração do intervalo de verificação

## Exemplo de Uso

### 1. Criar um Leilão

```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15",
    "category": "Electronics",
    "description": "iPhone 15 Pro Max 256GB",
    "condition": 1
  }'
```

### 2. Fazer um Lance

```bash
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "auction_id": "auction_id_here",
    "amount": 5000.00
  }'
```

### 3. Verificar Status do Leilão

```bash
curl http://localhost:8080/auction/auction_id_here
```

## Monitoramento

A aplicação registra logs para:

- Criação de leilões
- Fechamento automático de leilões
- Erros de operação
- Graceful shutdown

## Graceful Shutdown

A aplicação implementa graceful shutdown que:

1. Aguarda sinais de interrupção (SIGINT, SIGTERM)
2. Para a goroutine de verificação de leilões
3. Fecha conexões com o banco de dados
4. Encerra a aplicação de forma limpa

## Concorrência

O sistema utiliza goroutines e mutexes para garantir:

- Thread-safety nas operações de leilão
- Verificação assíncrona de leilões vencidos
- Processamento concorrente de lances
- Sincronização adequada dos dados em memória

## Estrutura de Dados

### Auction Entity

```go
type Auction struct {
    Id          string
    ProductName string
    Category    string
    Description string
    Condition   ProductCondition
    Status      AuctionStatus
    Timestamp   time.Time
}
```

### Status do Leilão

- `Active`: Leilão em andamento
- `Completed`: Leilão finalizado

## Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Implemente as mudanças
4. Adicione testes
5. Faça commit das mudanças
6. Abra um Pull Request 
