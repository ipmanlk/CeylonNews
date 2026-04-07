package scraper

import "strings"

type ListingConfig struct {
	Type    string `toml:"type"`
	URL     string `toml:"url"`
	Browser bool   `toml:"browser"`

	// HTML listing fields only
	Selectors []string `toml:"selectors"`
	URLPrefix string   `toml:"url_prefix"`
	BaseURL   string   `toml:"base_url"`
}

func (lc ListingConfig) ResolveLinks(links []string) []string {
	if lc.BaseURL == "" {
		return links
	}
	for i, link := range links {
		if strings.HasPrefix(link, "/") {
			links[i] = lc.BaseURL + link
		}
	}
	return links
}

type ArticleConfig struct {
	Browser  bool   `toml:"browser"`
	Selector string `toml:"selector"`
}

func (ac ArticleConfig) NeedsBrowser() bool {
	return ac.Browser || ac.Selector != ""
}

type LanguageConfig struct {
	Language string        `toml:"language"`
	MaxItems int           `toml:"max_items"`
	Listing  ListingConfig `toml:"listing"`
	Article  ArticleConfig `toml:"article"`
}

type SkipRule struct {
	Contains      string `toml:"contains"`
	CaseSensitive bool   `toml:"case_sensitive"`
}

type ReplaceRule struct {
	Pattern       string `toml:"pattern"`
	With          string `toml:"with"`
	CaseSensitive bool   `toml:"case_sensitive"`
}

type TitleTransform struct {
	Skip    []SkipRule    `toml:"skip"`
	Replace []ReplaceRule `toml:"replace"`
}

type Config struct {
	Name           string           `toml:"name"`
	Languages      []LanguageConfig `toml:"languages"`
	TitleTransform TitleTransform   `toml:"title_transform"`
}
