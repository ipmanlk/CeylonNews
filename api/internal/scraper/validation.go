package scraper

import (
	"regexp"
	"strings"

	"ipmanlk/cnapi/internal/model"
)

// ValidationEngine handles article validation
type ValidationEngine struct {
	config ValidationConfig
}

// NewValidationEngine creates a new validation engine
func NewValidationEngine(config ValidationConfig) *ValidationEngine {
	return &ValidationEngine{config: config}
}

// Validate checks if an article passes all validation rules
// Returns (shouldKeep bool, reason string)
func (v *ValidationEngine) Validate(article *model.ScrapedArticle) (bool, string) {
	// Check skip rules first - if any match, skip the article
	for _, rule := range v.config.Skip {
		if v.evaluateRule(article, rule) {
			return false, "skip rule matched: " + rule.Type + " on " + rule.Field
		}
	}

	// Check require rules - all must pass
	for _, rule := range v.config.Require {
		if !v.evaluateRule(article, rule) {
			return false, "require rule failed: " + rule.Type + " on " + rule.Field
		}
	}

	return true, ""
}

// ValidateTitle checks only title against validation rules
// Used for early validation before full extraction
func (v *ValidationEngine) ValidateTitle(title string) (bool, string) {
	// Check skip rules
	for _, rule := range v.config.Skip {
		if rule.Field == "title" || rule.Field == "any" {
			if v.evaluateField(title, rule) {
				return false, "skip rule matched on title"
			}
		}
	}

	// Check require rules
	for _, rule := range v.config.Require {
		if rule.Field == "title" || rule.Field == "any" {
			if !v.evaluateField(title, rule) {
				return false, "require rule failed on title"
			}
		}
	}

	return true, ""
}

func (v *ValidationEngine) evaluateRule(article *model.ScrapedArticle, rule ValidationRule) bool {
	switch rule.Field {
	case "title":
		return v.evaluateField(article.Title, rule)
	case "body":
		return v.evaluateField(article.ContentText, rule)
	case "any":
		return v.evaluateField(article.Title, rule) || v.evaluateField(article.ContentText, rule)
	default:
		return false
	}
}

func (v *ValidationEngine) evaluateField(value string, rule ValidationRule) bool {
	switch rule.Type {
	case "contains":
		return v.stringContains(value, rule.Value, rule.CaseSensitive)
	case "not_contains":
		return !v.stringContains(value, rule.Value, rule.CaseSensitive)
	case "regex":
		return v.matchesRegex(value, rule.Pattern)
	case "equals":
		return v.stringEquals(value, rule.Value, rule.CaseSensitive)
	case "not_equals":
		return !v.stringEquals(value, rule.Value, rule.CaseSensitive)
	case "min_length":
		return len(value) >= parseIntOrZero(rule.Value)
	case "max_length":
		return len(value) <= parseIntOrZero(rule.Value)
	default:
		return false
	}
}

func (v *ValidationEngine) stringContains(s, substr string, caseSensitive bool) bool {
	if substr == "" {
		return false
	}
	if caseSensitive {
		return strings.Contains(s, substr)
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func (v *ValidationEngine) stringEquals(s, other string, caseSensitive bool) bool {
	if caseSensitive {
		return s == other
	}
	return strings.EqualFold(s, other)
}

func (v *ValidationEngine) matchesRegex(s, pattern string) bool {
	if pattern == "" {
		return false
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

func parseIntOrZero(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}
