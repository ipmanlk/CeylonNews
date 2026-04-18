package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"ipmanlk/cnapi/internal/fetcher"

	"github.com/BurntSushi/toml"
)

// Kebab-case validation regex: starts with letter, followed by letters/numbers/hyphens
// No consecutive hyphens, no leading/trailing hyphens
var kebabCaseRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// ValidateSourceID checks if the source ID follows kebab-case format
func ValidateSourceID(id string) error {
	if id == "" {
		return fmt.Errorf("source id is required")
	}

	if !kebabCaseRegex.MatchString(id) {
		return fmt.Errorf("source id %q must be kebab-case: lowercase letters, numbers, and single hyphens only; must start with a letter", id)
	}

	return nil
}

// URLRule represents a single URL transformation rule
type URLRule struct {
	Type        string `toml:"type"`
	Value       string `toml:"value"`
	Pattern     string `toml:"pattern"`
	Replacement string `toml:"replacement"`
	Condition   string `toml:"condition"`
	MatchPolicy string `toml:"match_policy"`
}

// LinkSelector represents a link selector with optional title extraction
type LinkSelector struct {
	Link  string `toml:"link"`
	Title string `toml:"title"`
}

// HTMLDiscoveryConfig holds HTML-specific discovery settings
type HTMLDiscoveryConfig struct {
	LinkSelectors []LinkSelector `toml:"link_selectors"`
	URLRules      []URLRule      `toml:"url_rules"`
}

// DiscoveryConfig defines how to discover article URLs
type DiscoveryConfig struct {
	Type    string              `toml:"type"`
	URL     string              `toml:"url"`
	Browser bool                `toml:"browser"`
	HTML    HTMLDiscoveryConfig `toml:"html"`
}

// ContentConfig defines field-specific extraction selectors
type ContentConfig struct {
	ScopeSelector string `toml:"scope_selector"`
	TitleSelector string `toml:"title_selector"`
	BodySelector  string `toml:"body_selector"`
	ImageSelector string `toml:"image_selector"`
	DateSelector  string `toml:"date_selector"`
	PruneSelector string `toml:"prune_selector"`
}

// ExtractionConfig defines how to extract article content
type ExtractionConfig struct {
	Browser bool                  `toml:"browser"`
	Content fetcher.ContentConfig `toml:"content"`
}

// ValidationRule defines a single validation check
type ValidationRule struct {
	Field         string `toml:"field"`
	Type          string `toml:"type"`
	Value         string `toml:"value"`
	Pattern       string `toml:"pattern"`
	CaseSensitive bool   `toml:"case_sensitive"`
}

// ValidationConfig holds all validation rules
type ValidationConfig struct {
	Skip    []ValidationRule `toml:"skip"`
	Require []ValidationRule `toml:"require"`
}

// ReplaceRule defines a text replacement operation
type ReplaceRule struct {
	Field         string `toml:"field"`
	Pattern       string `toml:"pattern"`
	Replacement   string `toml:"replacement"`
	Regex         bool   `toml:"regex"`
	CaseSensitive bool   `toml:"case_sensitive"`
}

// NormalizeConfig defines normalization operations per field
type NormalizeConfig struct {
	Title []string `toml:"title"`
	Body  []string `toml:"body"`
}

// TransformationConfig holds all transformation rules
type TransformationConfig struct {
	Replace   []ReplaceRule   `toml:"replace"`
	Normalize NormalizeConfig `toml:"normalize"`
}

// LanguageConfig defines a complete pipeline for one language
type LanguageConfig struct {
	Language       string               `toml:"language"`
	MaxItems       int                  `toml:"max_items"`
	Shared         string               `toml:"shared"`
	Discovery      DiscoveryConfig      `toml:"discovery"`
	Extraction     ExtractionConfig     `toml:"extraction"`
	Validation     ValidationConfig     `toml:"validation"`
	Transformation TransformationConfig `toml:"transformation"`
}

// SharedConfig holds reusable configuration templates
type SharedConfig struct {
	Discovery      map[string]DiscoveryConfig      `toml:"discovery"`
	Extraction     map[string]ExtractionConfig     `toml:"extraction"`
	Validation     map[string]ValidationConfig     `toml:"validation"`
	Transformation map[string]TransformationConfig `toml:"transformation"`
}

// Config is the root configuration structure
type Config struct {
	ID        string           `toml:"id"`
	Name      string           `toml:"name"`
	Shared    SharedConfig     `toml:"shared"`
	Languages []LanguageConfig `toml:"languages"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if err := ValidateSourceID(c.ID); err != nil {
		return err
	}

	if c.Name == "" {
		return fmt.Errorf("source name is required")
	}

	if len(c.Languages) == 0 {
		return fmt.Errorf("at least one language configuration is required")
	}

	return nil
}

// LoadConfig loads a Config from TOML file and validates it
func LoadConfig(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Apply shared config inheritance
	for i := range cfg.Languages {
		if cfg.Languages[i].Shared != "" {
			cfg.Languages[i] = mergeWithShared(cfg.Languages[i], cfg.Shared)
		}
	}

	return &cfg, nil
}

// mergeWithShared applies shared config to language config
func mergeWithShared(lang LanguageConfig, shared SharedConfig) LanguageConfig {
	if lang.Shared == "" {
		return lang
	}

	// Merge discovery config
	if sharedDiscovery, ok := shared.Discovery[lang.Shared]; ok {
		lang.Discovery = mergeDiscoveryConfig(lang.Discovery, sharedDiscovery)
	}

	// Merge extraction config
	if sharedExtraction, ok := shared.Extraction[lang.Shared]; ok {
		lang.Extraction = mergeExtractionConfig(lang.Extraction, sharedExtraction)
	}

	// Merge validation config
	if sharedValidation, ok := shared.Validation[lang.Shared]; ok {
		lang.Validation = mergeValidationConfig(lang.Validation, sharedValidation)
	}

	// Merge transformation config
	if sharedTransformation, ok := shared.Transformation[lang.Shared]; ok {
		lang.Transformation = mergeTransformationConfig(lang.Transformation, sharedTransformation)
	}

	return lang
}

func mergeDiscoveryConfig(specific, shared DiscoveryConfig) DiscoveryConfig {
	if specific.Type == "" {
		specific.Type = shared.Type
	}
	if specific.URL == "" {
		specific.URL = shared.URL
	}
	// Browser defaults to false, so we only inherit if explicitly set in shared
	// (This is a design choice - could also inherit always)

	// Merge HTML config
	specific.HTML = mergeHTMLDiscoveryConfig(specific.HTML, shared.HTML)

	return specific
}

func mergeHTMLDiscoveryConfig(specific, shared HTMLDiscoveryConfig) HTMLDiscoveryConfig {
	// Arrays are replaced, not merged
	if len(specific.LinkSelectors) == 0 {
		specific.LinkSelectors = shared.LinkSelectors
	}

	// URL rules: append specific rules after shared rules
	if len(specific.URLRules) > 0 {
		specific.URLRules = append(shared.URLRules, specific.URLRules...)
	} else {
		specific.URLRules = shared.URLRules
	}

	return specific
}

func mergeExtractionConfig(specific, shared ExtractionConfig) ExtractionConfig {
	if !specific.Browser {
		specific.Browser = shared.Browser
	}

	// Merge content config
	if specific.Content.ScopeSelector == "" {
		specific.Content.ScopeSelector = shared.Content.ScopeSelector
	}
	if specific.Content.TitleSelector == "" {
		specific.Content.TitleSelector = shared.Content.TitleSelector
	}
	if specific.Content.BodySelector == "" {
		specific.Content.BodySelector = shared.Content.BodySelector
	}
	if specific.Content.ImageSelector == "" {
		specific.Content.ImageSelector = shared.Content.ImageSelector
	}
	if specific.Content.DateSelector == "" {
		specific.Content.DateSelector = shared.Content.DateSelector
	}
	if specific.Content.PruneSelector == "" {
		specific.Content.PruneSelector = shared.Content.PruneSelector
	}

	return specific
}

func mergeValidationConfig(specific, shared ValidationConfig) ValidationConfig {
	// Append specific rules after shared rules
	specific.Skip = append(shared.Skip, specific.Skip...)
	specific.Require = append(shared.Require, specific.Require...)
	return specific
}

func mergeTransformationConfig(specific, shared TransformationConfig) TransformationConfig {
	// Append specific rules after shared rules
	specific.Replace = append(shared.Replace, specific.Replace...)

	// Merge normalize config
	if len(specific.Normalize.Title) == 0 {
		specific.Normalize.Title = shared.Normalize.Title
	}
	if len(specific.Normalize.Body) == 0 {
		specific.Normalize.Body = shared.Normalize.Body
	}

	return specific
}

// ApplyURLRules executes the URL transformation pipeline
func ApplyURLRules(urls []string, rules []URLRule) []string {
	result := urls

	// Separate filtering and transformation rules
	var filterRules, transformRules []URLRule
	for _, rule := range rules {
		if isFilterRule(rule.Type) {
			filterRules = append(filterRules, rule)
		} else {
			transformRules = append(transformRules, rule)
		}
	}

	// Apply filtering rules by type with match_policy support
	result = applyFilterRulesByType(result, filterRules)

	// Apply transformation rules sequentially
	for _, rule := range transformRules {
		result = applyTransformRule(result, rule)
	}

	return result
}

// applyFilterRulesByType groups filter rules by type and applies match_policy
func applyFilterRulesByType(urls []string, rules []URLRule) []string {
	if len(rules) == 0 {
		return urls
	}

	// Group rules by type
	rulesByType := make(map[string][]URLRule)
	for _, rule := range rules {
		rulesByType[rule.Type] = append(rulesByType[rule.Type], rule)
	}

	result := urls
	for _, ruleGroup := range rulesByType {
		result = applyFilterRuleGroup(result, ruleGroup)
	}

	return result
}

// applyFilterRuleGroup applies a group of filter rules with match_policy
// match_policy = "any" (default): URL passes if it matches ANY rule in the group (OR logic)
// match_policy = "all": URL passes only if it matches ALL rules in the group (AND logic)
func applyFilterRuleGroup(urls []string, rules []URLRule) []string {
	if len(rules) == 0 {
		return urls
	}

	// Determine match policy - default to "any" if not specified or invalid
	matchPolicy := "any"
	if rules[0].MatchPolicy == "all" {
		matchPolicy = "all"
	}

	if matchPolicy == "any" {
		// OR logic: URL passes if it matches ANY rule
		return applyFilterRulesAny(urls, rules)
	}

	// AND logic: URL passes only if it matches ALL rules
	return applyFilterRulesAll(urls, rules)
}

// applyFilterRulesAny applies OR logic - URL passes if it matches ANY rule
func applyFilterRulesAny(urls []string, rules []URLRule) []string {
	if len(rules) == 0 {
		return urls
	}

	// Collect all URLs that match at least one rule
	matched := make(map[string]bool)
	for _, rule := range rules {
		matchingURLs := getMatchingURLs(urls, rule)
		for _, url := range matchingURLs {
			matched[url] = true
		}
	}

	// Return only matched URLs, preserving original order
	var result []string
	for _, url := range urls {
		if matched[url] {
			result = append(result, url)
		}
	}
	return result
}

// applyFilterRulesAll applies AND logic - URL passes only if it matches ALL rules
func applyFilterRulesAll(urls []string, rules []URLRule) []string {
	if len(rules) == 0 {
		return urls
	}

	result := urls
	for _, rule := range rules {
		result = applySingleFilterRule(result, rule)
		if len(result) == 0 {
			break // Early exit if no URLs remain
		}
	}
	return result
}

// getMatchingURLs returns URLs that match a filter rule
func getMatchingURLs(urls []string, rule URLRule) []string {
	switch rule.Type {
	case "filter_prefix":
		return filterByPrefix(urls, rule.Value, "")
	case "filter_not_prefix":
		return filterNotByPrefix(urls, rule.Value)
	case "filter_contains":
		return filterByContains(urls, rule.Value)
	case "filter_not_contains":
		return filterNotByContains(urls, rule.Value)
	case "filter_regex":
		return filterByRegex(urls, rule.Value)
	default:
		return urls
	}
}

// applySingleFilterRule applies a single filter rule
func applySingleFilterRule(urls []string, rule URLRule) []string {
	switch rule.Type {
	case "filter_prefix":
		return filterByPrefix(urls, rule.Value, "")
	case "filter_not_prefix":
		return filterNotByPrefix(urls, rule.Value)
	case "filter_contains":
		return filterByContains(urls, rule.Value)
	case "filter_not_contains":
		return filterNotByContains(urls, rule.Value)
	case "filter_regex":
		return filterByRegex(urls, rule.Value)
	default:
		return urls
	}
}

func isFilterRule(ruleType string) bool {
	return strings.HasPrefix(ruleType, "filter_")
}

func applyTransformRule(urls []string, rule URLRule) []string {
	switch rule.Type {
	case "prepend":
		return prependToURLs(urls, rule.Value, rule.Condition)
	case "append":
		return appendToURLs(urls, rule.Value)
	case "regex_replace":
		return regexReplaceURLs(urls, rule.Pattern, rule.Replacement)
	case "normalize":
		return normalizeURLs(urls)
	default:
		return urls
	}
}

func filterByPrefix(urls []string, prefix string, matchPolicy string) []string {
	if prefix == "" {
		return urls
	}

	var result []string
	for _, url := range urls {
		if strings.HasPrefix(url, prefix) {
			result = append(result, url)
		}
	}
	return result
}

func filterNotByPrefix(urls []string, prefix string) []string {
	if prefix == "" {
		return urls
	}

	var result []string
	for _, url := range urls {
		if !strings.HasPrefix(url, prefix) {
			result = append(result, url)
		}
	}
	return result
}

func filterByContains(urls []string, value string) []string {
	if value == "" {
		return urls
	}

	var result []string
	for _, url := range urls {
		if strings.Contains(url, value) {
			result = append(result, url)
		}
	}
	return result
}

func filterNotByContains(urls []string, value string) []string {
	if value == "" {
		return urls
	}

	var result []string
	for _, url := range urls {
		if !strings.Contains(url, value) {
			result = append(result, url)
		}
	}
	return result
}

func filterByRegex(urls []string, pattern string) []string {
	if pattern == "" {
		return urls
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return urls
	}

	var result []string
	for _, url := range urls {
		if re.MatchString(url) {
			result = append(result, url)
		}
	}
	return result
}

func prependToURLs(urls []string, prefix string, condition string) []string {
	if prefix == "" {
		return urls
	}

	var result []string
	for _, url := range urls {
		switch condition {
		case "if_relative":
			if strings.HasPrefix(url, "/") {
				result = append(result, prefix+url)
			} else {
				result = append(result, url)
			}
		case "if_protocol_relative":
			if strings.HasPrefix(url, "//") {
				result = append(result, prefix+url)
			} else {
				result = append(result, url)
			}
		default: // "always" or empty
			result = append(result, prefix+url)
		}
	}
	return result
}

func appendToURLs(urls []string, suffix string) []string {
	if suffix == "" {
		return urls
	}

	var result []string
	for _, url := range urls {
		result = append(result, url+suffix)
	}
	return result
}

func regexReplaceURLs(urls []string, pattern, replacement string) []string {
	if pattern == "" {
		return urls
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return urls
	}

	var result []string
	for _, url := range urls {
		result = append(result, re.ReplaceAllString(url, replacement))
	}
	return result
}

func normalizeURLs(urls []string) []string {
	var result []string
	for _, url := range urls {
		// Remove trailing slash
		url = strings.TrimSuffix(url, "/")
		// Ensure protocol is lowercase
		if strings.HasPrefix(url, "HTTP://") {
			url = "http://" + url[7:]
		} else if strings.HasPrefix(url, "HTTPS://") {
			url = "https://" + url[8:]
		}
		result = append(result, url)
	}
	return result
}
