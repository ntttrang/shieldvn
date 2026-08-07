package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"shieldvn-backend/internal/store"
)

func main() {
	projectID := flag.String("project", "", "GCP Project ID (default uses env inference)")
	spreadsheetID := flag.String("sheet", "", "Google Sheet ID containing the Tier 1 blacklist")
	sheetRange := flag.String("range", "Sheet1!A:D", "The range to read from the sheet")
	flag.Parse()

	if *spreadsheetID == "" {
		slog.Error("missing required -sheet flag (Google Sheet ID)")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize Sheets Store
	sheetsStore, err := store.NewSheetsStore(ctx)
	if err != nil {
		slog.Error("failed to init sheets store", "error", err)
		os.Exit(1)
	}

	// Initialize Firestore Store
	firestoreStore, err := store.NewFirestoreStore(ctx, *projectID)
	if err != nil {
		slog.Error("failed to init firestore store", "error", err)
		os.Exit(1)
	}
	defer firestoreStore.Close()

	slog.Info("Reading entities from Google Sheets...", "sheet", *spreadsheetID, "range", *sheetRange)
	entities, err := sheetsStore.ReadTier1Entities(ctx, *spreadsheetID, *sheetRange)
	if err != nil {
		slog.Error("failed to read from sheets", "error", err)
		os.Exit(1)
	}

	slog.Info(fmt.Sprintf("Found %d valid entities. Upserting to Firestore...", len(entities)))
	
	successCount := 0
	for _, entity := range entities {
		err := firestoreStore.UpsertTier1Entity(ctx, entity.EntityType, entity.Value, entity.Source, entity.Note)
		if err != nil {
			slog.Error("failed to upsert entity", "type", entity.EntityType, "value", entity.Value, "error", err)
		} else {
			successCount++
		}
	}

	slog.Info(fmt.Sprintf("Successfully upserted %d/%d entities into Firestore.", successCount, len(entities)))
}
