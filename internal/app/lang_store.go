package viva_api

import (
	"strings"
	"sync"
	"time"
)

// TTL по умолчанию для языковых предпочтений абонента (сценарий 7).
const defaultLangPreferenceTTL = 90 * 24 * time.Hour

// LangStore хранит языковые предпочтения абонента с истечением срока.
type LangStore interface {
	// Get возвращает сохранённый язык для телефона или "" если нет или истёк TTL.
	Get(phone string) string
	Set(phone, lang string)
	DefaultLang() string
}

type langEntry struct {
	lang      string
	expiresAt time.Time
}

type memoryLangStore struct {
	mu          sync.RWMutex
	byPhone     map[string]langEntry
	defaultLang string
	ttl         time.Duration
}

// NewLangStore создаёт in-memory хранилище. ttl переопределяет значение по умолчанию (90 дней); 0 — дефолт.
func NewLangStore(defaultLang string, ttl ...time.Duration) LangStore {
	prefTTL := defaultLangPreferenceTTL
	if len(ttl) > 0 && ttl[0] > 0 {
		prefTTL = ttl[0]
	}
	return &memoryLangStore{
		byPhone:     make(map[string]langEntry),
		defaultLang: normalizeLang(defaultLang),
		ttl:         prefTTL,
	}
}

func (s *memoryLangStore) Get(phone string) string {
	phone = normalizePhone(phone)
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.byPhone[phone]
	if !ok {
		return ""
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return ""
	}
	return entry.lang
}

func (s *memoryLangStore) Set(phone, lang string) {
	phone = normalizePhone(phone)
	lang = normalizeLang(lang)
	if phone == "" || lang == "" {
		return
	}
	s.mu.Lock()
	s.byPhone[phone] = langEntry{
		lang:      lang,
		expiresAt: time.Now().Add(s.ttl),
	}
	s.mu.Unlock()
}

func (s *memoryLangStore) DefaultLang() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.defaultLang
}

// resolveLang выбирает язык исходящих СМС: сохранённый → fallback (дефолт продукта) → дефолт приложения.
func (s *viva) resolveLang(phone, fallback string) string {
	if s.langStore != nil {
		if lang := s.langStore.Get(phone); lang != "" {
			return lang
		}
	}
	if lang := normalizeLang(fallback); lang != "" {
		return lang
	}
	if s.langStore != nil {
		return s.langStore.DefaultLang()
	}
	return ""
}

func normalizeLang(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "rus", "ru":
		return "ru"
	case "eng", "en":
		return "en"
	case "arm", "hy":
		return "arm"
	default:
		return s
	}
}
