#!/bin/bash

echo "=== Teste do Sistema de Leilões com Fechamento Automático ==="
echo ""

# Configurar variáveis de ambiente para teste rápido
export AUCTION_INTERVAL="10s"
export AUCTION_CHECK_INTERVAL="2s"

echo "Configuração:"
echo "- Duração do leilão: $AUCTION_INTERVAL"
echo "- Intervalo de verificação: $AUCTION_CHECK_INTERVAL"
echo ""

# Iniciar MongoDB (se não estiver rodando)
echo "Iniciando MongoDB..."
docker-compose up -d mongodb

# Aguardar MongoDB estar pronto
echo "Aguardando MongoDB estar pronto..."
sleep 5

# Iniciar aplicação em background
echo "Iniciando aplicação..."
go run cmd/auction/main.go &
APP_PID=$!

# Aguardar aplicação estar pronta
echo "Aguardando aplicação estar pronta..."
sleep 3

echo ""
echo "=== Teste 1: Criar um leilão ==="
AUCTION_RESPONSE=$(curl -s -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15 Pro",
    "category": "Electronics",
    "description": "iPhone 15 Pro Max 256GB Space Black",
    "condition": 1
  }')

echo "Resposta da criação do leilão:"
echo $AUCTION_RESPONSE
echo ""

# Extrair ID do leilão
AUCTION_ID=$(echo $AUCTION_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "ID do leilão criado: $AUCTION_ID"
echo ""

echo "=== Teste 2: Verificar status inicial do leilão ==="
curl -s http://localhost:8080/auction/$AUCTION_ID | jq .
echo ""

echo "=== Teste 3: Fazer alguns lances ==="
curl -s -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user1\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 5000.00
  }" | jq .

curl -s -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user2\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 5500.00
  }" | jq .

echo ""
echo "=== Teste 4: Aguardar fechamento automático ==="
echo "Aguardando $AUCTION_INTERVAL para o leilão fechar automaticamente..."
sleep 15

echo ""
echo "=== Teste 5: Verificar status final do leilão ==="
curl -s http://localhost:8080/auction/$AUCTION_ID | jq .

echo ""
echo "=== Teste 6: Tentar fazer lance em leilão fechado ==="
curl -s -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user3\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 6000.00
  }" | jq .

echo ""
echo "=== Teste 7: Verificar lance vencedor ==="
curl -s http://localhost:8080/auction/winner/$AUCTION_ID | jq .

echo ""
echo "=== Limpeza ==="
# Parar aplicação
kill $APP_PID
echo "Aplicação parada"

# Parar MongoDB
docker-compose down
echo "MongoDB parado"

echo ""
echo "=== Teste concluído ===" 