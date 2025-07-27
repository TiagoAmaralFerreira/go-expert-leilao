# Implementação do Fechamento Automático de Leilões

## Resumo da Implementação

Esta implementação adiciona funcionalidade de fechamento automático de leilões ao sistema existente, utilizando goroutines e concorrência para garantir performance e escalabilidade.

## Arquivos Modificados

### 1. `internal/infra/database/auction/create_auction.go`

**Principais mudanças:**

- **Estrutura do Repository**: Adicionados campos para controle de concorrência:
  ```go
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
  ```

- **Método UpdateAuctionStatus**: Novo método para atualizar status do leilão:
  ```go
  func (ar *AuctionRepository) UpdateAuctionStatus(
      ctx context.Context,
      auctionId string,
      status auction_entity.AuctionStatus) *internal_error.InternalError
  ```

- **Goroutine de Verificação**: Implementada em `startAuctionCloser()`:
  ```go
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
  ```

- **Lógica de Verificação**: Implementada em `checkAndCloseExpiredAuctions()`:
  ```go
  func (ar *AuctionRepository) checkAndCloseExpiredAuctions() {
      ctx := context.Background()
      now := time.Now()

      // Copia os maps para evitar deadlock
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
  ```

### 2. `internal/entity/auction_entity/auction_entity.go`

**Mudança:**
- Adicionado método `UpdateAuctionStatus` à interface `AuctionRepositoryInterface`

### 3. `cmd/auction/main.go`

**Mudanças:**
- Implementado graceful shutdown
- Adicionado tratamento de sinais (SIGINT, SIGTERM)
- Garantia de que a goroutine seja parada adequadamente

## Variáveis de Ambiente

### Configuração

- `AUCTION_INTERVAL`: Duração do leilão (padrão: 5m)
- `AUCTION_CHECK_INTERVAL`: Intervalo de verificação de leilões vencidos (padrão: 30s)

### Exemplos

```env
# Configuração do MongoDB
MONGODB_URL=mongodb://mongodb:27017
MONGODB_DB=auction_db

# Leilão de 10 minutos com verificação a cada 30 segundos
AUCTION_INTERVAL=10m
AUCTION_CHECK_INTERVAL=30s

# Leilão de 1 hora com verificação a cada 2 minutos
AUCTION_INTERVAL=1h
AUCTION_CHECK_INTERVAL=2m

# Leilão de 5 minutos com verificação a cada 10 segundos (para testes)
AUCTION_INTERVAL=5m
AUCTION_CHECK_INTERVAL=10s
```

## Concorrência e Thread-Safety

### Estratégias Implementadas

1. **Mutex para Maps**: Proteção contra race conditions nos maps de controle
2. **Cópia de Maps**: Evita deadlock durante verificação
3. **Channel para Stop**: Comunicação segura para parar goroutines
4. **WaitGroup**: Garante que todas as goroutines sejam finalizadas

### Pontos de Atenção

- Os maps são copiados durante a verificação para evitar deadlock
- Mutexes são usados apenas quando necessário
- Graceful shutdown garante limpeza adequada dos recursos

## Testes Implementados

### Arquivo: `internal/infra/database/auction/create_auction_test.go`

1. **TestGetAuctionInterval**: Valida leitura da variável de ambiente
2. **TestGetCheckInterval**: Valida leitura da variável de ambiente
3. **TestAuctionStatusManagement**: Testa gerenciamento de status em memória
4. **TestAuctionExpirationLogic**: Testa lógica de expiração

### Execução dos Testes

```bash
# Executar todos os testes
go test ./internal/infra/database/auction -v

# Executar com cobertura
go test ./internal/infra/database/auction -cover
```

## Fluxo de Funcionamento

### 1. Criação do Leilão
```
CreateAuction() → Adiciona aos maps → Inicia contagem regressiva
```

### 2. Verificação Periódica
```
Goroutine → Ticker → checkAndCloseExpiredAuctions() → UpdateAuctionStatus()
```

### 3. Fechamento Automático
```
Leilão expirado → Status = Completed → Logs registrados
```

### 4. Graceful Shutdown
```
SIGINT/SIGTERM → stopChan → Goroutine para → WaitGroup.Wait()
```

## Monitoramento e Logs

### Logs Implementados

- **Criação de leilão**: Log de sucesso
- **Fechamento automático**: Log informativo com ID do leilão
- **Erro no fechamento**: Log de erro com detalhes
- **Graceful shutdown**: Log de início e fim do processo

### Exemplo de Logs

```
INFO: Closing expired auction: 123e4567-e89b-12d3-a456-426614174000
INFO: Shutting down gracefully...
INFO: Server stopped
```

## Performance e Escalabilidade

### Otimizações Implementadas

1. **Verificação em Lote**: Todos os leilões são verificados de uma vez
2. **Maps em Memória**: Acesso rápido sem consultas ao banco
3. **Intervalo Configurável**: Permite ajuste baseado na carga
4. **Cópia Eficiente**: Minimiza tempo de lock dos mutexes

### Considerações

- **Memória**: Maps crescem com o número de leilões ativos
- **CPU**: Verificação periódica consome recursos
- **Banco**: Apenas atualizações quando necessário

## Troubleshooting

### Problemas Comuns

1. **Leilões não fecham**: Verificar variáveis de ambiente
2. **Alto uso de CPU**: Aumentar `AUCTION_CHECK_INTERVAL`
3. **Vazamento de memória**: Verificar se `Stop()` é chamado
4. **Race conditions**: Verificar uso correto dos mutexes

### Debug

```bash
# Verificar logs
docker-compose logs app

# Verificar variáveis de ambiente
docker-compose exec app env | grep AUCTION

# Testar endpoint manualmente
curl http://localhost:8080/auction/{auction_id}
```

## Próximos Passos

### Melhorias Sugeridas

1. **Persistência de Configuração**: Salvar configurações no banco
2. **Métricas**: Adicionar métricas de performance
3. **Notificações**: Notificar usuários sobre fechamento
4. **Cache Distribuído**: Usar Redis para múltiplas instâncias
5. **Testes de Integração**: Testes com MongoDB real 