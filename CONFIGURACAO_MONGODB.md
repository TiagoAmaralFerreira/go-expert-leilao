# Configuração do MongoDB

## 📍 Onde Configurar a Variável MONGODB_DB

### 1. Arquivo de Configuração Principal

**Localização**: `cmd/auction/.env`

**Conteúdo necessário**:
```env
MONGODB_URL=mongodb://mongodb:27017
MONGODB_DB=auction_db
AUCTION_INTERVAL=5m
AUCTION_CHECK_INTERVAL=30s
```

### 2. Como Criar o Arquivo .env

```bash
# Na raiz do projeto
cp env.example cmd/auction/.env
```

### 3. Variáveis de Ambiente do MongoDB

| Variável | Descrição | Valor Padrão | Exemplo |
|----------|-----------|--------------|---------|
| `MONGODB_URL` | URL de conexão com o MongoDB | `mongodb://mongodb:27017` | `mongodb://localhost:27017` |
| `MONGODB_DB` | Nome do banco de dados | `auction_db` | `auctions` |

### 4. Configurações Recomendadas

#### Para Desenvolvimento Local
```env
MONGODB_URL=mongodb://localhost:27017
MONGODB_DB=auction_db
AUCTION_INTERVAL=5m
AUCTION_CHECK_INTERVAL=30s
```

#### Para Docker
```env
MONGODB_URL=mongodb://mongodb:27017
MONGODB_DB=auction_db
AUCTION_INTERVAL=5m
AUCTION_CHECK_INTERVAL=30s
```

#### Para Testes
```env
MONGODB_URL=mongodb://localhost:27017
MONGODB_DB=test_auction_db
AUCTION_INTERVAL=1m
AUCTION_CHECK_INTERVAL=10s
```

### 5. Verificação da Configuração

#### Verificar se o arquivo .env existe:
```bash
ls -la cmd/auction/.env
```

#### Verificar conteúdo do arquivo:
```bash
cat cmd/auction/.env
```

#### Testar conexão com MongoDB:
```bash
# Com Docker
docker-compose up -d mongodb

# Verificar se o MongoDB está rodando
docker-compose ps
```

### 6. Estrutura do Banco de Dados

Quando a aplicação iniciar, ela criará automaticamente:

- **Database**: `auction_db` (ou o valor de `MONGODB_DB`)
- **Collections**:
  - `auctions` - Leilões
  - `bids` - Lances
  - `users` - Usuários

### 7. Comandos Úteis do MongoDB

#### Conectar ao MongoDB:
```bash
# Via Docker
docker exec -it mongodb mongosh

# Local
mongosh
```

#### Listar bancos de dados:
```javascript
show dbs
```

#### Usar o banco de dados:
```javascript
use auction_db
```

#### Listar collections:
```javascript
show collections
```

#### Ver dados dos leilões:
```javascript
db.auctions.find()
```

### 8. Troubleshooting

#### Problema: "Error trying to connect to mongodb database"
**Solução**: Verificar se o MongoDB está rodando
```bash
docker-compose up -d mongodb
```

#### Problema: "Error trying to ping mongodb database"
**Solução**: Verificar a URL de conexão
```env
MONGODB_URL=mongodb://mongodb:27017
```

#### Problema: "Database not found"
**Solução**: Verificar o nome do banco de dados
```env
MONGODB_DB=auction_db
```

### 9. Exemplo Completo de Configuração

Crie o arquivo `cmd/auction/.env` com o seguinte conteúdo:

```env
# Configuração do MongoDB
MONGODB_URL=mongodb://mongodb:27017
MONGODB_DB=auction_db

# Configuração dos Leilões
AUCTION_INTERVAL=5m
AUCTION_CHECK_INTERVAL=30s
```

### 10. Execução

Após configurar o arquivo `.env`, execute:

```bash
# Com Docker
docker-compose up --build

# Local
go run cmd/auction/main.go
```

A aplicação irá:
1. Conectar ao MongoDB usando as variáveis configuradas
2. Criar o banco de dados automaticamente
3. Criar as collections necessárias
4. Iniciar o sistema de leilões com fechamento automático 