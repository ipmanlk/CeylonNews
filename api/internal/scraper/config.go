package scraper

type ListingConfig struct {
	Type    string `toml:"type"`
	URL     string `toml:"url"`
	Browser bool   `toml:"browser"`

	Selectors []string `toml:"selectors"`
	URLPrefix string   `toml:"url_prefix"`
	BaseURL   string   `toml:"base_url"`
}

type ArticleConfig struct {
	Browser  bool   `toml:"browser"`
	Selector string `toml:"selector"`
}

type LangConfig struct {
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

type SourceConfig struct {
	Name           string         `toml:"name"`
	Languages      []LangConfig   `toml:"languages"`
	TitleTransform TitleTransform `toml:"title_transform"`
}
