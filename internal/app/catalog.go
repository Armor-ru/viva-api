package viva_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Product interface {
	GetNotify(key string, data map[string]interface{}, lang ...string) string
}

type Catalog interface {
	Load(dir string) error
	GetProductByShortNumber(id string) (Product, error)
	GetProductByExternalId(id string) (Product, error)
	SetDefaultLang(lang string) error
}

type catalog struct {
	products    []*product
	byShort     map[string]*product
	byExternal  map[string]*product
	defaultLang string
}

type product struct {
	ShortNumber     string                 `json:"shortNumber"`
	ExternalId      string                 `json:"externalId"`
	DefaultLanguage string                 `json:"defaultLanguage"`
	Notifications   map[string]interface{} `json:"notifications"`
	catalogLang     string                 `json:"-"`
}

func NewCatalog() Catalog {
	return &catalog{
		products:   make([]*product, 0),
		byShort:    make(map[string]*product),
		byExternal: make(map[string]*product),
	}
}

func (c *catalog) Load(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read catalog directory: %w", err)
	}

	c.products = make([]*product, 0)
	c.byShort = make(map[string]*product)
	c.byExternal = make(map[string]*product)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read catalog file %q: %w", path, err)
		}
		var item product
		if err := json.Unmarshal(raw, &item); err != nil {
			return fmt.Errorf("decode catalog file %q: %w", path, err)
		}

		item.ShortNumber = strings.TrimSpace(item.ShortNumber)
		item.ExternalId = strings.TrimSpace(item.ExternalId)
		item.DefaultLanguage = strings.TrimSpace(item.DefaultLanguage)
		if item.ShortNumber == "" {
			return fmt.Errorf("invalid catalog file %q: shortNumber is required", path)
		}
		if item.ExternalId == "" {
			return fmt.Errorf("invalid catalog file %q: externalId is required", path)
		}
		if _, exists := c.byShort[item.ShortNumber]; exists {
			return fmt.Errorf("duplicate shortNumber %q in catalog", item.ShortNumber)
		}
		if _, exists := c.byExternal[item.ExternalId]; exists {
			return fmt.Errorf("duplicate externalId %q in catalog", item.ExternalId)
		}
		item.catalogLang = c.defaultLang
		p := &item
		c.products = append(c.products, p)
		c.byShort[p.ShortNumber] = p
		c.byExternal[p.ExternalId] = p
	}

	if len(c.products) == 0 {
		return fmt.Errorf("catalog directory %q has no products", dir)
	}
	return nil
}

func (c *catalog) SetDefaultLang(lang string) error {
	c.defaultLang = strings.TrimSpace(lang)
	for i := range c.products {
		c.products[i].catalogLang = c.defaultLang
	}
	return nil
}

func (c *catalog) GetProductByShortNumber(id string) (Product, error) {
	p, ok := c.byShort[strings.TrimSpace(id)]
	if !ok {
		return nil, fmt.Errorf("product by short number %q not found", id)
	}
	return p, nil
}

func (c *catalog) GetProductByExternalId(id string) (Product, error) {
	p, ok := c.byExternal[strings.TrimSpace(id)]
	if !ok {
		return nil, fmt.Errorf("product by external id %q not found", id)
	}
	return p, nil
}

func (p *product) GetNotify(key string, data map[string]interface{}, lang ...string) string {
	l := ""
	if len(lang) > 0 {
		l = strings.TrimSpace(lang[0])
	}
	if l == "" {
		l = p.DefaultLanguage
	}
	if l == "" {
		l = p.catalogLang
	}

	text := ""
	if byLang, ok := p.Notifications[l].(map[string]interface{}); ok {
		text, _ = byLang[key].(string)
	}
	if text == "" {
		for _, v := range p.Notifications {
			if byLang, ok := v.(map[string]interface{}); ok {
				if t, ok := byLang[key].(string); ok && t != "" {
					text = t
					break
				}
			}
		}
	}
	if text == "" {
		return ""
	}

	tpl, err := template.New("notify").Parse(text)
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}
