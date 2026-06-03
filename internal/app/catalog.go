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

type Product struct {
	ShortNumber     string                 `json:"shortNumber"`
	ExternalId      string                 `json:"externalId"`
	DefaultLanguage string                 `json:"defaultLanguage"`
	Notifications   map[string]interface{} `json:"notifications"`
	catalogLang     string                 `json:"-"`
}

func (p *Product) GetNotify(key string, data map[string]interface{}, lang ...string) string {
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
	if p.Notifications != nil {
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

type Catalog struct {
	products    []*Product
	defaultLang string
}

func NewCatalog() *Catalog {
	return &Catalog{
		products: make([]*Product, 0),
	}
}

func (c *Catalog) Load(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read catalog directory: %w", err)
	}

	c.products = make([]*Product, 0)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read catalog file %q: %w", path, err)
		}
		var item Product
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

		item.catalogLang = c.defaultLang
		p := &item
		c.products = append(c.products, p)
	}

	if len(c.products) == 0 {
		return fmt.Errorf("catalog directory %q has no products", dir)
	}
	return nil
}

func (c *Catalog) SetDefaultLang(lang string) error {
	c.defaultLang = strings.TrimSpace(lang)
	for _, p := range c.products {
		p.catalogLang = c.defaultLang
	}
	return nil
}

func (c *Catalog) GetProductByShortNumber(id string) (*Product, error) {
	id = strings.TrimSpace(id)
	for _, p := range c.products {
		if p.ShortNumber == id {
			return p, nil
		}
	}
	return nil, fmt.Errorf("product by short number %q not found", id)
}

func (c *Catalog) GetProductByExternalId(id string) (*Product, error) {
	id = strings.TrimSpace(id)
	for _, p := range c.products {
		if p.ExternalId == id {
			return p, nil
		}
	}
	return nil, fmt.Errorf("product by external id %q not found", id)
}
