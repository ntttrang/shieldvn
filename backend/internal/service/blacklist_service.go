package service

import (
	"context"
	"fmt"
	"log/slog"

	"shieldvn-backend/internal/model"
	"shieldvn-backend/internal/store"
)

type BlacklistService struct {
	firestoreStore *store.FirestoreStore
}

func NewBlacklistService(firestoreStore *store.FirestoreStore) *BlacklistService {
	return &BlacklistService{
		firestoreStore: firestoreStore,
	}
}

// CheckTier1 checks the extracted entities against the Tier 1 blacklist.
// It returns a slice of evidence strings for any hits found.
func (s *BlacklistService) CheckTier1(ctx context.Context, entities model.ExtractedEntities) []string {
	var evidence []string

	if s.firestoreStore == nil {
		slog.Warn("Firestore store is not initialized, skipping Tier 1 blacklist check")
		return evidence
	}

	// Helper to check and append evidence
	checkEntity := func(entityType, value, vnName string) {
		if value == "" {
			return
		}
		
		isListed, source, err := s.firestoreStore.IsListed(ctx, entityType, value)
		if err != nil {
			slog.Error("failed to check tier-1 blacklist", "entityType", entityType, "value", value, "error", err)
			return
		}
		
		if isListed {
			evidenceStr := fmt.Sprintf("%s %s nằm trong danh sách đen Tier-1 (nguồn: %s)", vnName, value, source)
			evidence = append(evidence, evidenceStr)
		}
	}

	checkEntity("bank_account", entities.BankAccount, "Số tài khoản ngân hàng")
	checkEntity("phone_number", entities.PhoneNumber, "Số điện thoại")
	checkEntity("url", entities.URL, "Đường dẫn (URL)")

	return evidence
}
