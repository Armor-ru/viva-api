package catalog_test

import (
	"path/filepath"
	"testing"

	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
)

func TestLoadSafekid(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "catalog")
	store, err := catalog.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	store.SetDefaultLang("ru")

	steps, err := store.Steps("SAFEKID", "mo.1")
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 2 {
		t.Fatalf("mo.1 steps: got %d want 2", len(steps))
	}
	if steps[0].Command != "new" || steps[1].Command != "notify" {
		t.Fatalf("unexpected steps: %+v", steps)
	}

	text, err := store.RenderNotify("SAFEKID", "welcome_trial", "ru", map[string]interface{}{
		"ExternalID": "SAFEKID",
	})
	if err != nil {
		t.Fatal(err)
	}
	if text == "" {
		t.Fatal("expected non-empty welcome text")
	}

	_, _, ok := store.ProductByShortNumber("1020")
	if !ok {
		t.Fatal("expected short number 1020")
	}
}
