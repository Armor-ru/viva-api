package viva_api

import (
	"testing"
	"time"
)

func TestLangStore_SetGetWithTTL(t *testing.T) {
	t.Parallel()

	store := NewLangStore("ru", time.Hour)
	phone := "37477600552"
	store.Set(phone, "en")

	if got := store.Get(phone); got != "en" {
		t.Fatalf("Get = %q, want en", got)
	}
	if got := store.DefaultLang(); got != "ru" {
		t.Fatalf("DefaultLang = %q", got)
	}
}

func TestLangStore_GetEmptyWithoutSet(t *testing.T) {
	t.Parallel()

	store := NewLangStore("ru")
	if got := store.Get("37477600552"); got != "" {
		t.Fatalf("Get without Set = %q, want empty", got)
	}
}

func TestResolveLang_PrefersStoredOverProductDefault(t *testing.T) {
	t.Parallel()

	v := &viva{langStore: NewLangStore("ru")}
	phone := "37477600552"
	v.langStore.Set(phone, "en")

	if got := v.resolveLang(phone, "ru"); got != "en" {
		t.Fatalf("resolveLang = %q, want en", got)
	}
}

func TestResolveLang_FallsBackToProductDefault(t *testing.T) {
	t.Parallel()

	v := &viva{langStore: NewLangStore("ru")}
	if got := v.resolveLang("37477600552", "ru"); got != "ru" {
		t.Fatalf("resolveLang = %q", got)
	}
}

func TestAppConfig_ResolvedLangPreferenceTTL(t *testing.T) {
	t.Parallel()

	if got := (AppConfig{}).ResolvedLangPreferenceTTL(); got != defaultLangPreferenceTTL {
		t.Fatalf("default ttl = %v", got)
	}
	if got := (AppConfig{LangPreferenceTTL: "48h"}).ResolvedLangPreferenceTTL(); got != 48*time.Hour {
		t.Fatalf("48h ttl = %v", got)
	}
}
