// Package main est le point d'entrée de l'application StreamFlix Backend.
//
// Ce package initialise la configuration, le logging structuré, le routeur HTTP Gin
// et démarre le serveur avec un graceful shutdown pour gérer proprement les arrêts.
//
// Flux de démarrage :
//  1. Chargement de la configuration depuis les variables d'environnement (.env.dev)
//  2. Validation des tokens API obligatoires (TMDB, Real-Debrid)
//  3. Configuration du logging structuré JSON via slog
//  4. Création du routeur Gin avec tous les middleware et routes
//  5. Démarrage du serveur HTTP avec timeouts configurés
//  6. Écoute des signaux SIGINT/SIGTERM pour un arrêt gracieux
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"StreamflixBackend/internal/config"
	router "StreamflixBackend/internal/http"
)

// main initialise et démarre le serveur HTTP StreamFlix.
//
// Le serveur écoute sur le port configuré via la variable d'environnement PORT
// (défaut : 2000) et s'arrête proprement à la réception d'un signal SIGINT ou SIGTERM
// avec un timeout de 30 secondes pour terminer les requêtes en cours.
func main() {
	// Chargement de la configuration depuis .env.dev et les variables d'environnement
	cfg := config.Load()
	cfg.Validate()

	// Configuration du logging structuré en JSON pour la production
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	})
	slog.SetDefault(slog.New(logHandler))

	// Construction du routeur Gin avec middleware, routes API v1 et routes legacy
	engine := router.NewRouter(cfg)

	// Création du serveur HTTP avec des timeouts de sécurité contre les connexions lentes
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,  // Timeout de lecture du body de la requête
		WriteTimeout: 30 * time.Second,  // Timeout d'écriture de la réponse
		IdleTimeout:  60 * time.Second,  // Timeout pour les connexions keep-alive inactives
	}

	// Démarrage du serveur dans une goroutine séparée pour ne pas bloquer le main
	go func() {
		slog.Info("server starting", slog.String("port", cfg.Port), slog.String("mode", cfg.GinMode))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown : écoute des signaux d'arrêt système (Ctrl+C ou kill)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutdown signal received", slog.String("signal", sig.String()))

	// Contexte avec timeout de 30s pour laisser le temps aux requêtes en cours de se terminer
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}
