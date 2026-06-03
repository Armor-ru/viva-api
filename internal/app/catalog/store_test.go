package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalog_LoadAndGetProduct(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "..", "catalog", "safekid.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "safekid.json"), src, 0644); err != nil {
		t.Fatal(err)
	}

	cat := NewCatalog()
	if err := cat.Load(dir); err != nil {
		t.Fatal(err)
	}
	_ = cat.SetDefaultLang("ru")

	p, err := cat.GetProductByShortNumber("1020")
	if err != nil {
		t.Fatal(err)
	}
	if p.GetExternalID() != "SAFEKID" {
		t.Fatalf("externalId = %q", p.GetExternalID())
	}

	text := p.GetNotify("welcome_trial", map[string]interface{}{
		"Link": "https://example.com/dl",
	}, "ru")
	if text == "" || !strings.Contains(text, "https://example.com/dl") {
		t.Fatalf("welcome_trial = %q", text)
	}
}

func TestCatalog_UnknownCommand_Message13_Russian(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "..", "catalog", "safekid.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "safekid.json"), src, 0644); err != nil {
		t.Fatal(err)
	}
	cat := NewCatalog()
	if err := cat.Load(dir); err != nil {
		t.Fatal(err)
	}
	p, err := cat.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	text := p.GetNotify("unknown_command", map[string]interface{}{}, "ru")
	if !strings.Contains(text, "для подключения услуги") {
		t.Fatalf("unknown_command = %q", text)
	}
}

func TestCatalog_LanguageChanged_Message10_Russian(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "..", "catalog", "safekid.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "safekid.json"), src, 0644); err != nil {
		t.Fatal(err)
	}
	cat := NewCatalog()
	if err := cat.Load(dir); err != nil {
		t.Fatal(err)
	}
	p, err := cat.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	text := p.GetNotify("language_changed", map[string]interface{}{}, "ru")
	if !strings.Contains(text, "изменен на русский") {
		t.Fatalf("language_changed ru = %q", text)
	}
}

func TestCatalog_AlreadyDeactivated_Message9_Russian(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "..", "catalog", "safekid.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "safekid.json"), src, 0644); err != nil {
		t.Fatal(err)
	}
	cat := NewCatalog()
	if err := cat.Load(dir); err != nil {
		t.Fatal(err)
	}
	p, err := cat.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	text := p.GetNotify("already_deactivated", map[string]interface{}{}, "ru")
	if !strings.Contains(text, "услуга уже отключена") {
		t.Fatalf("already_deactivated = %q", text)
	}
}

func TestCatalog_ServiceDeactivated_Message6_Russian(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "..", "catalog", "safekid.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "safekid.json"), src, 0644); err != nil {
		t.Fatal(err)
	}
	cat := NewCatalog()
	if err := cat.Load(dir); err != nil {
		t.Fatal(err)
	}
	p, err := cat.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	text := p.GetNotify("service_deactivated", map[string]interface{}{}, "ru")
	if !strings.Contains(text, "Услуга деактивирована") {
		t.Fatalf("service_deactivated = %q", text)
	}
}

func TestCatalog_WelcomePaid_Message4_Russian(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "..", "catalog", "safekid.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "safekid.json"), src, 0644); err != nil {
		t.Fatal(err)
	}
	cat := NewCatalog()
	if err := cat.Load(dir); err != nil {
		t.Fatal(err)
	}
	p, err := cat.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	text := p.GetNotify("welcome_paid", map[string]interface{}{}, "ru")
	if !strings.Contains(text, "Платная версия услуги Kaspersky Safe Kids подключена") {
		t.Fatalf("welcome_paid = %q", text)
	}
}

func TestCatalog_TrialExpires_Message5_Russian(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "..", "catalog", "safekid.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "safekid.json"), src, 0644); err != nil {
		t.Fatal(err)
	}
	cat := NewCatalog()
	if err := cat.Load(dir); err != nil {
		t.Fatal(err)
	}
	p, err := cat.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	text := p.GetNotify("trial_expires", map[string]interface{}{"ExpiresAt": "10.06.2026 12:00"}, "ru")
	if !strings.Contains(text, "10.06.2026 12:00") || !strings.Contains(text, "пробный период истекает") {
		t.Fatalf("trial_expires = %q", text)
	}
}

func TestCatalog_AlreadyActive_Message8_Russian(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src, err := os.ReadFile(filepath.Join("..", "..", "..", "catalog", "safekid.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "safekid.json"), src, 0644); err != nil {
		t.Fatal(err)
	}
	cat := NewCatalog()
	if err := cat.Load(dir); err != nil {
		t.Fatal(err)
	}
	p, err := cat.GetProductByExternalId("SAFEKID")
	if err != nil {
		t.Fatal(err)
	}
	want := "Уважаемый абонент, услуга уже подключена. Инфо:"
	if got := p.GetNotify("already_active", nil, "ru"); got != want {
		t.Fatalf("already_active = %q", got)
	}
}

