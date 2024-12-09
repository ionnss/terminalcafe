#!/bin/bash

echo "🔄 Iniciando atualização do Terminal Café..."

# Verifica se está rodando como root
if [ "$EUID" -ne 0 ]; then 
    echo "❌ Por favor, execute como root (sudo)"
    exit 1
fi

# Backup antes da atualização
echo "📦 Criando backup..."
BACKUP_FILE="backup-$(date +%F-%H%M).tar.gz"
tar -czf "$BACKUP_FILE" .env server.key products/menu.md
echo "✅ Backup criado: $BACKUP_FILE"

# Pull das últimas alterações
echo "⬇️ Baixando atualizações..."
git pull

# Atualiza dependências do Docker
echo "🐳 Atualizando containers..."
docker-compose down
docker-compose pull
docker-compose build --no-cache

# Reinicia os serviços
echo "🚀 Reiniciando serviços..."
docker-compose up -d

# Verifica status
echo "🔍 Verificando status..."
docker-compose ps

echo "✨ Atualização concluída!"
echo "📝 Logs disponíveis em: docker-compose logs -f" 