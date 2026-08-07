package store

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FirestoreStore handles interactions with Cloud Firestore.
type FirestoreStore struct {
	client *firestore.Client
}

// NewFirestoreStore initializes a new FirestoreStore.
// It assumes GOOGLE_APPLICATION_CREDENTIALS is set in the environment or a service account JSON is provided.
func NewFirestoreStore(ctx context.Context, projectID string) (*FirestoreStore, error) {
	// If projectID is empty, the client will attempt to infer it from the environment.
	if projectID == "" {
		projectID = firestore.DetectProjectID
	}
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return &FirestoreStore{
		client: client,
	}, nil
}

// Close closes the Firestore client.
func (s *FirestoreStore) Close() error {
	return s.client.Close()
}

// IsListed checks if an entity is in the Tier-1 blacklist.
func (s *FirestoreStore) IsListed(ctx context.Context, entityType, value string) (bool, string, error) {
	// We use a composite key for the document ID: {entityType}_{value}
	docID := fmt.Sprintf("%s_%s", entityType, value)

	doc, err := s.client.Collection("blacklist_tier1").Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, "", nil
		}
		return false, "", fmt.Errorf("firestore get error: %w", err)
	}

	// Entity is found
	source := "unknown"
	if s, ok := doc.Data()["source"].(string); ok {
		source = s
	}

	return true, source, nil
}

// UpsertTier1Entity inserts or updates an entity in the Tier-1 blacklist.
func (s *FirestoreStore) UpsertTier1Entity(ctx context.Context, entityType, value, source, note string) error {
	docID := fmt.Sprintf("%s_%s", entityType, value)

	_, err := s.client.Collection("blacklist_tier1").Doc(docID).Set(ctx, map[string]interface{}{
		"entity_type": entityType,
		"entity_value": value,
		"source":      source,
		"note":        note,
	})
	if err != nil {
		return fmt.Errorf("firestore set error: %w", err)
	}
	return nil
}
