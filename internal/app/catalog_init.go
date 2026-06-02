package viva_api

import (
	"fmt"
	"strings"

	"git.dev.armlab.pro/armor/viva-api/internal/app/catalog"
	"git.dev.armlab.pro/armor/viva-api/internal/app/pipeline"
)

func (s *Viva) loadCatalog() error {
	dir := strings.TrimSpace(s.sms.MenuDir)
	if dir == "" {
		return fmt.Errorf("sms.menuDir is required")
	}
	store, err := catalog.Load(dir)
	if err != nil {
		return err
	}
	store.SetDefaultLang(s.sms.DefaultLanguage)
	s.catalog = store
	s.engine = &pipeline.Engine{
		Catalog: store,
		Actions: vivaActions{v: s},
	}
	return nil
}
