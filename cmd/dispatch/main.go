// Stockyard Dispatch — Email list and newsletter.
// Manage subscribers, send campaigns via your SMTP, track opens. Self-hosted.
// Single binary, embedded SQLite, zero external dependencies.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"github.com/stockyard-dev/stockyard-dispatch/internal/server"
	"github.com/stockyard-dev/stockyard-dispatch/internal/store"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Printf("dispatch %s\n", version)
		os.Exit(0)
	}
	if len(os.Args) > 1 && (os.Args[1] == "--health" || os.Args[1] == "health") {
		fmt.Println("ok")
		os.Exit(0)
	}

	log.SetFlags(log.Ltime | log.Lshortfile)

	retentionDays := 30
	if r := os.Getenv("RETENTION_DAYS"); r != "" {
		if n, err := strconv.Atoi(r); err == nil && n > 0 {
			retentionDays = n
		}
	}

	port := 8900
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	smtpCfg := server.SMTPConfig{
		Host: os.Getenv("SMTP_HOST"),
		Port: os.Getenv("SMTP_PORT"),
		User: os.Getenv("SMTP_USER"),
		Pass: os.Getenv("SMTP_PASS"),
		From: os.Getenv("SMTP_FROM"),
	}
	if smtpCfg.Port == "" {
		smtpCfg.Port = "587"
	}

		limits := server.DefaultLimits()
	if limits.RetentionDays > retentionDays {
		retentionDays = limits.RetentionDays
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	smtpStatus := "not configured"
	if smtpCfg.Host != "" {
		smtpStatus = smtpCfg.Host + ":" + smtpCfg.Port
	}

	log.Printf("")
	log.Printf("  Stockyard Dispatch %s", version)
	log.Printf("  API:            http://localhost:%d/api/lists", port)
	log.Printf("  Subscribe:      POST http://localhost:%d/subscribe/{list_id}", port)
	log.Printf("  SMTP:           %s", smtpStatus)
	log.Printf("  Retention:      %d days", retentionDays)
	log.Printf("  Dashboard:      http://localhost:%d/ui", port)
	log.Printf("")

	go func() {
		for {
			time.Sleep(6 * time.Hour)
			n, err := db.Cleanup(retentionDays)
			if err != nil {
				log.Printf("[cleanup] error: %v", err)
			} else if n > 0 {
				log.Printf("[cleanup] deleted %d old send records", n)
			}
		}
	}()

	srv := server.New(db, port, limits, smtpCfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
