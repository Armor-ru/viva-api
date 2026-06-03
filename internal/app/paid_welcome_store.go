package viva_api

import (
	"strings"
	"sync"
)

// PaidWelcomeStore учитывает отправку СМС №4 (триал → платный) по абоненту и продукту.
// Чтобы не слать приветствие при каждом renew (сценарий 4, шаг 12).
type PaidWelcomeStore interface {
	AlreadySent(phone, productCode string) bool
	MarkSent(phone, productCode string)
	Clear(phone, productCode string)
}

type memoryPaidWelcomeStore struct {
	mu   sync.RWMutex
	keys map[string]struct{}
}

func NewPaidWelcomeStore() PaidWelcomeStore {
	return &memoryPaidWelcomeStore{keys: make(map[string]struct{})}
}

func paidWelcomeKey(phone, productCode string) string {
	return normalizePhone(phone) + ":" + strings.TrimSpace(productCode)
}

func (s *memoryPaidWelcomeStore) AlreadySent(phone, productCode string) bool {
	key := paidWelcomeKey(phone, productCode)
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.keys[key]
	return ok
}

func (s *memoryPaidWelcomeStore) MarkSent(phone, productCode string) {
	key := paidWelcomeKey(phone, productCode)
	if key == ":" {
		return
	}
	s.mu.Lock()
	s.keys[key] = struct{}{}
	s.mu.Unlock()
}

func (s *memoryPaidWelcomeStore) Clear(phone, productCode string) {
	key := paidWelcomeKey(phone, productCode)
	if key == ":" {
		return
	}
	s.mu.Lock()
	delete(s.keys, key)
	s.mu.Unlock()
}

func (s *viva) paidWelcomeStore() PaidWelcomeStore {
	if s.paidWelcomeSent == nil {
		s.paidWelcomeSent = NewPaidWelcomeStore()
	}
	return s.paidWelcomeSent
}
