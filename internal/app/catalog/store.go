package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NewCatalog — пустой файловый каталог (реализация catalog.Catalog).
func NewCatalog() Catalog {
	return &catalog{
		Products:   make([]product, 0),
		byShort:    make(map[string]*product),
		byExternal: make(map[string]*product),
	}
}

func (c *catalog) Load(dir string) error {
	loaded, err := loadDir(dir)
	if err != nil {
		return err
	}
	c.Products = loaded.Products
	c.byShort = loaded.byShort
	c.byExternal = loaded.byExternal
	return nil
}

func (c *catalog) SetDefaultLang(lang string) error {
	c.defaultLang = normalizeLang(lang)
	return nil
}

func (c *catalog) GetProductByShortNumber(id string) (Product, error) {
	p, ok := c.byShort[normalizeKey(id)]
	if !ok {
		return nil, fmt.Errorf("product by short number %q not found", id)
	}
	return p, nil
}

func (c *catalog) GetProductByExternalId(id string) (Product, error) {
	p, ok := c.byExternal[normalizeKey(id)]
	if !ok {
		return nil, fmt.Errorf("product by external id %q not found", id)
	}
	return p, nil
}

func (p *product) GetExternalID() string  { return p.ExternalId }
func (p *product) GetShortNumber() string { return p.ShortNumber }
func (p *product) GetDefaultLanguage() string {
	return p.DefaultLanguage
}

func (p *product) GetNotify(key string, data map[string]interface{}, lang ...string) string {
	text := p.resolveNotification(resolveLangArg(lang...), key)
	if text == "" {
		return ""
	}
	rendered, err := renderTemplate(text, data)
	if err != nil {
		return ""
	}
	return rendered
}

func (p *product) resolveNotification(language, key string) string {
	lang := normalizeKey(language)
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if lang != "" {
		if byLang, ok := p.Notifications[lang]; ok {
			if text := strings.TrimSpace(byLang[key]); text != "" {
				return text
			}
		}
	}
	def := normalizeLang(p.DefaultLanguage)
	if def == "" {
		def = normalizeLang(p.catalogLang)
	}
	if def != "" {
		if byLang, ok := p.Notifications[def]; ok {
			if text := strings.TrimSpace(byLang[key]); text != "" {
				return text
			}
		}
	}
	for _, byLang := range p.Notifications {
		if text := strings.TrimSpace(byLang[key]); text != "" {
			return text
		}
	}
	return ""
}

func resolveLangArg(lang ...string) string {
	if len(lang) == 0 {
		return ""
	}
	return normalizeLang(lang[0])
}

func loadDir(dir string) (*catalog, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read catalog directory: %w", err)
	}

	c := &catalog{
		Products:   make([]product, 0),
		byShort:    make(map[string]*product),
		byExternal: make(map[string]*product),
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		prod, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		shortKey := normalizeKey(prod.ShortNumber)
		extKey := normalizeKey(prod.ExternalId)
		if _, exists := c.byShort[shortKey]; exists {
			return nil, fmt.Errorf("duplicate shortNumber %q in catalog", prod.ShortNumber)
		}
		if _, exists := c.byExternal[extKey]; exists {
			return nil, fmt.Errorf("duplicate externalId %q in catalog", prod.ExternalId)
		}
		c.Products = append(c.Products, prod)
		p := &c.Products[len(c.Products)-1]
		c.byShort[shortKey] = p
		c.byExternal[extKey] = p
	}

	if len(c.Products) == 0 {
		return nil, fmt.Errorf("catalog directory %q has no products", dir)
	}
	return c, nil
}

func loadFile(path string) (product, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return product{}, fmt.Errorf("read catalog file %q: %w", path, err)
	}
	var raw ProductRecord
	if err := json.Unmarshal(content, &raw); err != nil {
		return product{}, fmt.Errorf("decode catalog file %q: %w", path, err)
	}
	shortNumber := normalizeKey(raw.ShortNumber)
	externalID := strings.TrimSpace(raw.ExternalID)
	if shortNumber == "" {
		return product{}, fmt.Errorf("invalid catalog file %q: shortNumber is required", path)
	}
	if externalID == "" {
		return product{}, fmt.Errorf("invalid catalog file %q: externalId is required", path)
	}
	return product{
		ShortNumber:     shortNumber,
		ExternalId:      externalID,
		DefaultLanguage: strings.TrimSpace(raw.DefaultLanguage),
		Notifications:   raw.Notifications,
	}, nil
}

func normalizeKey(s string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.ToLower(s), "+"))
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
