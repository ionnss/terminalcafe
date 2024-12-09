package main

import (
	"log"
	"path/filepath"

	"terminal-cafe/internal/config"
	"terminal-cafe/internal/server"
	"terminal-cafe/internal/store"
)

func main() {
	// Carrega configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Inicializa store com configurações
	store := store.NewStore(cfg)

	menuPath := filepath.Join("products", "menu.md")
	if err := store.LoadProductsFromMD(menuPath); err != nil {
		log.Fatalf("Erro ao carregar menu: %v", err)
	}

	server := server.NewServer(store, 2222)
	if err := server.Start(); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
