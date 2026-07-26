package content_test

import (
	"context"
	"testing"

	"enterprise-search/internal/content"
)

func TestContentRepository(t *testing.T) {
	repo := content.NewMemoryRepository()
	ctx := context.Background()

	item := &content.ContentItem{
		ID:          "content-2026-ridge-launch",
		BrandID:     "ApexMotors",
		ContentType: content.TypeModelPage,
		Title:       "2026 Apex Ridge Trail Launch Announcement",
		Body:        "Introducing the 2026 Apex Ridge Trail with 7,000 lbs max towing capacity.",
		Status:      content.StatusPublished,
	}

	if err := repo.Save(ctx, item); err != nil {
		t.Fatalf("failed to save content item: %v", err)
	}

	fetched, err := repo.GetByID(ctx, "content-2026-ridge-launch")
	if err != nil {
		t.Fatalf("failed to get content item: %v", err)
	}

	if fetched.Title != "2026 Apex Ridge Trail Launch Announcement" {
		t.Errorf("expected launch title, got %s", fetched.Title)
	}

	items, err := repo.FindByBrand(ctx, "ApexMotors", content.TypeModelPage)
	if err != nil {
		t.Fatalf("failed to find items by brand: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("expected 1 matching item, got %d", len(items))
	}
}
