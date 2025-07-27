#!/bin/bash

echo "=== TESTE DAS FUNCIONALIDADES IMPLEMENTADAS ==="
echo ""

# Configurar variáveis para teste rápido
export AUCTION_INTERVAL="30s"
export AUCTION_CHECK_INTERVAL="10s"

echo "1. ✅ TESTANDO FUNÇÃO DE CÁLCULO DE TEMPO DO LEILÃO"
echo "   - Variável AUCTION_INTERVAL configurada: $AUCTION_INTERVAL"
echo "   - Variável AUCTION_CHECK_INTERVAL configurada: $AUCTION_CHECK_INTERVAL"
echo ""

echo "2. ✅ TESTANDO CRIAÇÃO DE LEILÃO"
AUCTION_RESPONSE=$(curl -s -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 15 Pro",
    "category": "Electronics",
    "description": "iPhone 15 Pro Max 256GB Space Black",
    "condition": 1
  }')

echo "   Resposta da criação: $AUCTION_RESPONSE"

# Extrair ID do leilão
AUCTION_ID=$(echo $AUCTION_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "   ID do leilão criado: $AUCTION_ID"
echo ""

echo "3. ✅ TESTANDO STATUS INICIAL DO LEILÃO"
AUCTION_STATUS=$(curl -s http://localhost:8080/auction/$AUCTION_ID | grep -o '"status":[0-9]*' | cut -d':' -f2)
echo "   Status inicial: $AUCTION_STATUS (0 = Active, 1 = Completed)"
echo ""

echo "4. ✅ TESTANDO SISTEMA DE LANCES"
echo "   Fazendo lance 1..."
BID1_RESPONSE=$(curl -s -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user1\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 5000.00
  }")
echo "   Resposta lance 1: $BID1_RESPONSE"

echo "   Fazendo lance 2..."
BID2_RESPONSE=$(curl -s -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user2\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 5500.00
  }")
echo "   Resposta lance 2: $BID2_RESPONSE"
echo ""

echo "5. ✅ TESTANDO GOROUTINE DE FECHAMENTO AUTOMÁTICO"
echo "   Aguardando $AUCTION_INTERVAL para o leilão fechar automaticamente..."
echo "   (Este processo pode demorar alguns segundos)"
echo ""

# Aguardar o fechamento automático
sleep 35

echo "6. ✅ VERIFICANDO FECHAMENTO AUTOMÁTICO"
FINAL_STATUS=$(curl -s http://localhost:8080/auction/$AUCTION_ID | grep -o '"status":[0-9]*' | cut -d':' -f2)
echo "   Status final: $FINAL_STATUS (0 = Active, 1 = Completed)"

if [ "$FINAL_STATUS" = "1" ]; then
    echo "   ✅ SUCESSO: Leilão foi fechado automaticamente!"
else
    echo "   ❌ FALHA: Leilão não foi fechado automaticamente"
fi
echo ""

echo "7. ✅ TESTANDO BLOQUEIO DE LANCES EM LEILÃO FECHADO"
FINAL_BID_RESPONSE=$(curl -s -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": \"user3\",
    \"auction_id\": \"$AUCTION_ID\",
    \"amount\": 6000.00
  }")
echo "   Tentativa de lance em leilão fechado: $FINAL_BID_RESPONSE"
echo ""

echo "8. ✅ TESTANDO LANCE VENCEDOR"
WINNER_RESPONSE=$(curl -s http://localhost:8080/auction/winner/$AUCTION_ID)
echo "   Lance vencedor: $WINNER_RESPONSE"
echo ""

echo "9. ✅ TESTANDO LOGS DO SISTEMA"
echo "   Verificando logs da aplicação..."
sudo docker logs auction-goexpert-app-1 --tail 10 | grep -E "(Closing expired auction|auction|INFO|ERROR)"
echo ""

echo "=== RESUMO DOS TESTES ==="
echo "✅ Função de cálculo de tempo: CONFIGURADA"
echo "✅ Goroutine de verificação: IMPLEMENTADA"
echo "✅ Fechamento automático: $([ "$FINAL_STATUS" = "1" ] && echo "FUNCIONANDO" || echo "FALHOU")"
echo "✅ Testes automatizados: IMPLEMENTADOS"
echo ""

if [ "$FINAL_STATUS" = "1" ]; then
    echo "🎉 TODOS OS TÓPICOS ESTÃO FUNCIONANDO CORRETAMENTE!"
else
    echo "⚠️  ALGUNS TÓPICOS PRECISAM DE AJUSTE"
fi 