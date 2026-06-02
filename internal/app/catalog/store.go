package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.dev.armlab.pro/armor/viva-api/internal/app/pipeline"
	"gopkg.in/yaml.v3"
)

// Store is a file-backed product catalog (implements pipeline.Catalog).
type Store struct {
	byShort    map[string]Product
	byExternal map[string]Product
	defaultLang string
}

func Load(dir string) (*Store, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, fmt.Errorf("catalog directory is empty")
	}
	if _, err := os.Stat(dir); err != nil {
		return nil, fmt.Errorf("stat catalog dir: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read catalog dir: %w", err)
	}

	c := &Store{
		byShort:    make(map[string]Product),
		byExternal: make(map[string]Product),
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if !strings.HasSuffix(name, ".json") && !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		product, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		if err := validate(path, &product); err != nil {
			return nil, err
		}
		shortKey := normalizeKey(product.ShortNumber)
		extKey := normalizeKey(product.ExternalID)
		if _, ok := c.byShort[shortKey]; ok {
			return nil, fmt.Errorf("duplicate shortNumber %q", product.ShortNumber)
		}
		if _, ok := c.byExternal[extKey]; ok {
			return nil, fmt.Errorf("duplicate externalId %q", product.ExternalID)
		}
		c.byShort[shortKey] = product
		c.byExternal[extKey] = product
	}

	if len(c.byShort) == 0 {
		return nil, fmt.Errorf("catalog dir %q has no products", dir)
	}
	return c, nil
}

func (c *Store) SetDefaultLang(lang string) {
	c.defaultLang = normalizeLang(lang)
}

func (c *Store) Steps(productCode, scenarioKey string) ([]pipeline.Step, error) {
	if c == nil {
		return nil, fmt.Errorf("catalog is not loaded")
	}
	product, ok := c.byExternal[normalizeKey(productCode)]
	if !ok {
		return nil, fmt.Errorf("product %q not found", productCode)
	}
	steps, ok := findScenario(product.Scenarios, scenarioKey)
	if !ok || len(steps) == 0 {
		return nil, fmt.Errorf("scenario %q not found for product %q", scenarioKey, productCode)
	}
	return toPipelineSteps(steps), nil
}

func findScenario(scenarios map[string][]Step, key string) ([]Step, bool) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, false
	}
	if steps, ok := scenarios[key]; ok {
		return steps, true
	}
	for k, steps := range scenarios {
		if strings.EqualFold(k, key) {
			return steps, true
		}
	}
	return nil, false
}

func (c *Store) RenderNotify(productCode, tplKey, lang string, data map[string]interface{}) (string, error) {
	if c == nil {
		return "", fmt.Errorf("catalog is not loaded")
	}
	product, ok := c.byExternal[normalizeKey(productCode)]
	if !ok {
		return "", fmt.Errorf("product %q not found", productCode)
	}
	language := c.ResolveLang(productCode, lang)
	tpl := resolveNotification(product, language, tplKey)
	if tpl == "" {
		return "", fmt.Errorf("template %q not found for lang %q", tplKey, language)
	}
	return renderTemplate(tpl, data)
}

func (c *Store) ResolveLang(productCode, langHint string) string {
	if c == nil {
		return ""
	}
	if l := normalizeLang(langHint); l != "" {
		return l
	}
	if product, ok := c.byExternal[normalizeKey(productCode)]; ok {
		if l := normalizeLang(product.DefaultLanguage); l != "" {
			return l
		}
	}
	return normalizeLang(c.defaultLang)
}

func (c *Store) ProductByShortNumber(short string) (productCode, shortNumber string, ok bool) {
	if c == nil {
		return "", "", false
	}
	product, found := c.byShort[normalizeKey(short)]
	if !found {
		return "", "", false
	}
	return product.ExternalID, product.ShortNumber, true
}

func (c *Store) ProductByExternalID(externalID string) (productCode, shortNumber string, ok bool) {
	if c == nil {
		return "", "", false
	}
	product, found := c.byExternal[normalizeKey(externalID)]
	if !found {
		return "", "", false
	}
	return product.ExternalID, product.ShortNumber, true
}

func loadFile(path string) (Product, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Product{}, err
	}
	var product Product
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		err = json.Unmarshal(raw, &product)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(raw, &product)
	default:
		return Product{}, fmt.Errorf("unsupported catalog file %q", path)
	}
	if err != nil {
		return Product{}, fmt.Errorf("decode %q: %w", path, err)
	}
	return product, nil
}

func validate(path string, p *Product) error {
	p.ShortNumber = strings.TrimSpace(p.ShortNumber)
	p.ExternalID = strings.TrimSpace(p.ExternalID)
	if p.ShortNumber == "" {
		return fmt.Errorf("%s: shortNumber required", path)
	}
	if p.ExternalID == "" {
		return fmt.Errorf("%s: externalId required", path)
	}
	if len(p.Scenarios) == 0 {
		return fmt.Errorf("%s: scenarios required", path)
	}
	for name, steps := range p.Scenarios {
		if normalizeKey(name) == "" {
			return fmt.Errorf("%s: empty scenario key", path)
		}
		if len(steps) == 0 {
			return fmt.Errorf("%s: scenario %q has no steps", path, name)
		}
		for i, step := range steps {
			step.Command = normalizeKey(step.Command)
			if step.Command == "" {
				return fmt.Errorf("%s: scenario %q step %d: command required", path, name, i)
			}
			switch step.Command {
			case "new", "renew", "cancel":
			case "notify":
				if strings.TrimSpace(castToString(step.Params["tpl"])) == "" {
					return fmt.Errorf("%s: scenario %q step %d: notify requires params.tpl", path, name, i)
				}
			default:
				return fmt.Errorf("%s: scenario %q: unsupported command %q", path, name, step.Command)
			}
		}
	}
	return nil
}

func castToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func toPipelineSteps(steps []Step) []pipeline.Step {
	out := make([]pipeline.Step, len(steps))
	for i, s := range steps {
		out[i] = pipeline.Step{Command: s.Command, Params: s.Params}
	}
	return out
}

func normalizeKey(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func normalizeLang(language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "arm", "hy":
		return "arm"
	case "rus", "ru":
		return "ru"
	case "eng", "en":
		return "en"
	default:
		return ""
	}
}

func resolveNotification(p Product, lang, key string) string {
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
