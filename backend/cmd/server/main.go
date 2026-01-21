package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/fn-jakubkarp/coresend/internal/smtp"
	"github.com/fn-jakubkarp/coresend/internal/store"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	domain := getEnv("DOMAIN_NAME", "localhost")
	listenAddr := getEnv("SMTP_LISTEN_ADDR", ":1025")
	certPath := os.Getenv("SMTP_CERT_PATH")
	keyPath := os.Getenv("SMTP_KEY_PATH")

	emailStore := store.NewStore(redisAddr, redisPassword)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := emailStore.Ping(ctx); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", redisAddr, err)
	}
	log.Printf("Connected to Redis at %s", redisAddr)

	be := &smtp.Backend{
		Store: emailStore,
	}

	s := gosmtp.NewServer(be)
	s.Addr = listenAddr
	s.Domain = domain
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	// Inbound-only server: we accept mail from any sender without authentication
	s.AllowInsecureAuth = true

	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			log.Printf("Warning: TLS certificate failed to load (STARTTLS disabled): %v", err)
		} else {
			s.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS12,
			}
			log.Println("TLS certificates loaded successfully")
		}
	} else {
		log.Println("TLS certificates not configured, running without STARTTLS")
	}

	// Graceful shutdown on SIGINT/SIGTERM
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down SMTP server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := s.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	log.Printf("SMTP server starting on %s for domain %s", listenAddr, domain)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("SMTP server error: %v", err)
	}
}
