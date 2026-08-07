package store

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Tier1Entity struct {
	EntityType string
	Value      string
	Source     string
	Note       string
}

// SheetsStore handles reading from Google Sheets.
type SheetsStore struct {
	client *sheets.Service
}

// NewSheetsStore initializes a new SheetsStore.
func NewSheetsStore(ctx context.Context) (*SheetsStore, error) {
	// If the sheet is public, we may not even need credentials, but using default credentials is safe.
	srv, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsReadonlyScope))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %w", err)
	}

	return &SheetsStore{client: srv}, nil
}

// ReadTier1Entities reads the blacklist entities from a given sheet ID and range.
// Assumes columns are: [entity_type, entity_value, source, note]
func (s *SheetsStore) ReadTier1Entities(ctx context.Context, spreadsheetId, readRange string) ([]Tier1Entity, error) {
	resp, err := s.client.Spreadsheets.Values.Get(spreadsheetId, readRange).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %w", err)
	}

	var entities []Tier1Entity
	for i, row := range resp.Values {
		// Skip header row if i == 0, assuming first row is header
		if i == 0 {
			continue
		}
		
		if len(row) >= 2 {
			entityType := fmt.Sprintf("%v", row[0])
			value := fmt.Sprintf("%v", row[1])
			source := ""
			note := ""
			if len(row) >= 3 {
				source = fmt.Sprintf("%v", row[2])
			}
			if len(row) >= 4 {
				note = fmt.Sprintf("%v", row[3])
			}
			
			if entityType != "" && value != "" {
				entities = append(entities, Tier1Entity{
					EntityType: entityType,
					Value:      value,
					Source:     source,
					Note:       note,
				})
			}
		}
	}
	return entities, nil
}
