package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/tbauriedel/go-nland-liveticker/internal/config"
	"github.com/tbauriedel/go-nland-liveticker/internal/database"
	"github.com/tbauriedel/go-nland-liveticker/internal/model"
	"github.com/tbauriedel/go-nland-liveticker/internal/scraper"
	"github.com/tbauriedel/go-nland-liveticker/internal/telegram"
)

func main() {
	// Get config from environment. Defaults are applied in case variables are missing
	conf := config.GetFromEnv()

	// Init logger. We use debug by default
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	logger.Info("Starting go-nland-liveticker")

	// sqlite connection
	db := database.NewSQLiteDatabase(conf.DatabaseDSN, logger)
	if db.ConnectErr != nil {
		logger.Error("cant connect to database", "error", db.ConnectErr)
		os.Exit(1)
	}

	err := db.ImportSchema(context.Background())
	if err != nil {
		logger.Error("cant import schema", "error", err)
		os.Exit(1)
	}

	// Telegram instance
	bot, err := telegram.NewBotInstance(conf.TelegramBotId, logger)
	if err != nil {
		logger.Error("cant create telegram bot instance", "error", err)
		os.Exit(1)
	}

	// Init scraper
	logger.Info("Creating scraper")
	s := scraper.NewScraper()
	s.Collector.AllowURLRevisit = true

	logger.Info("Registering scraper")
	s.Register()

	logger.Info("Starting endless scraper")

	// endless loop
	for {
		// Wait before scraping. Done here because of more than one continue in loop
		time.Sleep(5 * time.Second)

		lastOperationFromScraper, err := s.ScrapeOperations()
		if err != nil {
			logger.Error("scraping failed. Waiting 5 seconds before new try", "error", err)
			time.Sleep(5 * time.Second)

			continue
		}

		// context for db select
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		lastOperationFromDB, err := db.GetLastInsertedOperation(ctx)
		if err != nil {
			logger.Error("scraping failed", "error", err)
			continue
		}

		if lastOperationFromDB == nil {
			logger.Debug("No operation found in database yet")

			handleOperation(bot, conf.TelegramChatId, lastOperationFromScraper, logger, db)
			continue
		}

		if lastOperationFromDB.GetIdentifier() == lastOperationFromScraper.GetIdentifier() {
			logger.Debug("Latest found operation already in database")
			continue
		}

		logger.Info("New operation found", "operation", lastOperationFromScraper.GetIdentifier())

		// Send new operation to telegram and insert into database
		handleOperation(bot, conf.TelegramChatId, lastOperationFromScraper, logger, db)
	}
}

func handleOperation(t telegram.Bot, chatID int64, o model.Operation, logger *slog.Logger, db database.Database) {
	sent := false
	for !sent {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info(fmt.Sprintf("recovered from: %v\nWill wait 1 second before retry", r))
					time.Sleep(1 * time.Second)
				}
			}()

			// Build message
			message := "\U0001F692 " + o.Units + "\n" +
				"\U0001F4A5 " + strings.Replace(o.Report, "\n", " ", -1) + "\n" +
				"\U0001F4CD " + strings.Replace(o.Location, "\n", "", -1) + " (" + o.District + ")\n" +
				"\U0001F551 " + o.Time.Format("02.01.2006 15:04")

			err := t.SendMessage(chatID, message)
			if err != nil {
				panic(fmt.Errorf("panic. failed to send operation: %#v %w", o, err))
			}

			sent = true

			logger.Debug("Sent operation to telegram", "operation", o.GetIdentifier())
		}()
	}

	inserted := false
	for !inserted {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Info(fmt.Sprintf("recovered from: %v\nWill wait 1 second before retry", r))
					time.Sleep(1 * time.Second)
				}
			}()

			// save to database
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := db.InsertOperation(ctx, &o)
			if err != nil {
				panic(fmt.Errorf("panic. failed to insert operation: %#v. %w", o, err))
			}

			inserted = true
		}()
	}
}
