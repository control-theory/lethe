package graph

import (
	"context"
	"database/sql"
)

// CheckDuplicate checks if a node with the same kind+name already exists,
// or if there's a likely FTS match.
func (s *Store) CheckDuplicate(ctx context.Context, kind NodeKind, name, content string) (*DedupResult, error) {
	// Strategy 1: exact kind+name match (case-insensitive)
	existing, err := s.GetNodeByName(ctx, kind, name)
	if err == nil && existing != nil {
		return &DedupResult{
			IsDuplicate: true,
			ExistingID:  existing.ID,
			Confidence:  "exact",
		}, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Strategy 2: FTS similarity on name + content
	searchText := name
	if content != "" {
		searchText = name + " " + content
	}
	results, err := s.Search(ctx, searchText, SearchOpts{Kind: &kind, Limit: 1, IncludeFaded: false})
	if err != nil {
		// FTS search failure is non-fatal for dedup; treat as no match
		return &DedupResult{Confidence: "none"}, nil
	}
	if len(results) > 0 {
		// If the top result has the same name (case-insensitive), treat as likely dup
		if equalsIgnoreCase(results[0].Name, name) {
			return &DedupResult{
				IsDuplicate: true,
				ExistingID:  results[0].ID,
				Confidence:  "likely",
			}, nil
		}
	}

	return &DedupResult{Confidence: "none"}, nil
}

func equalsIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
