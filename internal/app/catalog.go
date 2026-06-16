package viva_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"git.dev.armlab.pro/armor/sds-go/pkg/errs"
	"git.dev.armlab.pro/armor/sds-go/pkg/tplext"
)

type Product struct {
	ShortNumber       string                 `json:"shortNumber"`
	ExternalId        string                 `json:"externalId"`
	LandingConfirmURL string                 `json:"landingConfirmUrl"`
	Notifications     map[string]interface{} `json:"notifications"`
}

func (p *Product) GetNotify(key string, data map[string]interface{}, lang ...string) string {
	l := ""
	if len(lang) > 0 {
		l = strings.TrimSpace(lang[0])
	}

	text := ""
	if p.Notifications != nil {
		if l != "" {
			if byLang, ok := p.Notifications[l].(map[string]interface{}); ok {
				text, _ = byLang[key].(string)
			}
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

	tpl, err := template.New("notify").Funcs(tplext.Funcs).Parse(text)
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
	products []*Product
}

func NewCatalog() *Catalog {
	return &Catalog{
		products: make([]*Product, 0),
	}
}

func (c *Catalog) Load(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return errs.WrapWithFields(
			fmt.Errorf("read catalog directory failed, %w", err),
			map[string]interface{}{"dir": dir},
		)
	}

	c.products = make([]*Product, 0)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		raw, err := os.ReadFile(path)
		if err != nil {
			return errs.WrapWithFields(
				fmt.Errorf("read catalog file failed, %w", err),
				map[string]interface{}{"path": path},
			)
		}

		var item Product
		if err := json.Unmarshal(raw, &item); err != nil {
			return errs.WrapWithFields(
				fmt.Errorf("decode catalog file failed, %w", err),
				map[string]interface{}{"path": path},
			)
		}

		item.ShortNumber = strings.TrimSpace(item.ShortNumber)
		item.ExternalId = strings.TrimSpace(item.ExternalId)
		item.LandingConfirmURL = strings.TrimSpace(item.LandingConfirmURL)

		if item.ShortNumber == "" {
			return errs.WrapWithFields(
				fmt.Errorf("catalog product shortNumber is required"),
				map[string]interface{}{"path": path},
			)
		}

		if item.ExternalId == "" {
			return errs.WrapWithFields(
				fmt.Errorf("catalog product externalId is required"),
				map[string]interface{}{"path": path},
			)
		}

		for _, p := range c.products {
			if p.ShortNumber == item.ShortNumber {
				return errs.WrapWithFields(
					fmt.Errorf("duplicate catalog product shortNumber"),
					map[string]interface{}{
						"path":        path,
						"shortNumber": item.ShortNumber,
					},
				)
			}

			if p.ExternalId == item.ExternalId {
				return errs.WrapWithFields(
					fmt.Errorf("duplicate catalog product externalId"),
					map[string]interface{}{
						"path":       path,
						"externalId": item.ExternalId,
					},
				)
			}
		}

		p := &item
		c.products = append(c.products, p)
	}

	if len(c.products) == 0 {
		return errs.WrapWithFields(
			fmt.Errorf("catalog directory has no products"),
			map[string]interface{}{"dir": dir},
		)
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

	return nil, errs.WrapWithFields(
		fmt.Errorf("product by short number not found"),
		map[string]interface{}{"shortNumber": id},
	)
}

func (c *Catalog) GetProductByExternalId(id string) (*Product, error) {
	id = strings.TrimSpace(id)

	for _, p := range c.products {
		if p.ExternalId == id {
			return p, nil
		}
	}

	return nil, errs.WrapWithFields(
		fmt.Errorf("product by external id not found"),
		map[string]interface{}{"externalId": id},
	)
}
