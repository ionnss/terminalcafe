FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instala dependências de build
RUN apk add --no-cache gcc musl-dev

# Copia os arquivos de módulo
COPY go.mod go.sum ./
RUN go mod download

# Copia o código fonte
COPY . .

# Compila a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o terminal-cafe ./cmd/terminal-cafe

# Imagem final
FROM alpine:latest

WORKDIR /app

# Instala openssh para o ssh-keygen
RUN apk add --no-cache openssh

# Copia o binário compilado
COPY --from=builder /app/terminal-cafe .
# Copia o menu de produtos
COPY products/menu.md ./products/

# Cria usuário e diretórios necessários
RUN adduser -D appuser && \
    mkdir -p /app/keys && \
    chown -R appuser:appuser /app

USER appuser

# Expõe a porta SSH
EXPOSE 2222

# Comando para iniciar a aplicação
CMD ["./terminal-cafe"] 