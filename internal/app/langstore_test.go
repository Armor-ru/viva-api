package viva_api

import (
	"testing"
	"time"
)

func newLangStoreViva() *Viva {
	return &Viva{langStore: make(LangStore)}
}

func TestSetLang_GetLang(t *testing.T) {
	t.Parallel()

	v := newLangStoreViva()
	v.SetLang("37477600552", "en")
	if got := v.GetLang("37477600552"); got != "en" {
		t.Fatalf("GetLang() = %q, want en", got)
	}
}

func TestSetLang_ignoresUnknownLang(t *testing.T) {
	t.Parallel()

	v := newLangStoreViva()
	v.SetLang("37477600552", "hy")
	if got := v.GetLang("37477600552"); got != "" {
		t.Fatalf("GetLang() = %q, want empty", got)
	}
}

func TestGetLang_normalizesPhone(t *testing.T) {
	t.Parallel()

	v := newLangStoreViva()
	v.SetLang("37477600552", "arm")
	if got := v.GetLang("+37477600552"); got != "arm" {
		t.Fatalf("GetLang(+phone) = %q, want arm", got)
	}
	if got := v.GetLang(" 37477600552 "); got != "arm" {
		t.Fatalf("GetLang(trimmed phone) = %q, want arm", got)
	}
}

func TestSetLang_ignoresEmptyInput(t *testing.T) {
	t.Parallel()

	v := newLangStoreViva()
	v.SetLang("", "en")
	v.SetLang("37477600552", "")
	v.SetLang("  ", "ru")
	if len(v.langStore) != 0 {
		t.Fatalf("langStore len = %d, want 0", len(v.langStore))
	}
}

func TestGetLang_missingOrExpired(t *testing.T) {
	t.Parallel()

	v := newLangStoreViva()
	if got := v.GetLang("37477600552"); got != "" {
		t.Fatalf("GetLang(missing) = %q, want empty", got)
	}

	v.langStore["37477600552"] = langEntry{
		lang:      "en",
		expiresAt: time.Now().Add(-time.Hour),
	}
	if got := v.GetLang("37477600552"); got != "" {
		t.Fatalf("GetLang(expired) = %q, want empty", got)
	}
}

func TestSetLang_overwritesPrevious(t *testing.T) {
	t.Parallel()

	v := newLangStoreViva()
	v.SetLang("37477600552", "en")
	v.SetLang("37477600552", "ru")
	if got := v.GetLang("37477600552"); got != "ru" {
		t.Fatalf("GetLang() = %q, want ru", got)
	}
}
