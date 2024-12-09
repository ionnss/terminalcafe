#!/bin/bash

# Verifica se o arquivo .env existe
if [ ! -f .env ]; then
    echo "Arquivo .env não encontrado. Criando um novo..."
    cp .env.example .env
    echo "Por favor, configure as variáveis no arquivo .env"
    exit 1
fi

# Gera chave SSH se não existir
if [ ! -f server.key ]; then
    echo "Gerando chave SSH..."
    ssh-keygen -t rsa -f server.key -N ""
fi

# Inicia os containers
docker-compose up -d

echo "Terminal Café está rodando!"
echo "Para conectar: ssh -p 2222 localhost" 