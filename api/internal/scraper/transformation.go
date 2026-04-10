package scraper

import (
	"html"
	"regexp"
	"strings"

	"ipmanlk/cnapi/internal/model"
)

// TransformationEngine handles article content transformation
type TransformationEngine struct {
	config TransformationConfig
}

// NewTransformationEngine creates a new transformation engine
func NewTransformationEngine(config TransformationConfig) *TransformationEngine {
	return &TransformationEngine{config: config}
}

// Transform applies all transformation rules to an article
func (t *TransformationEngine) Transform(article *model.ScrapedArticle) {
	// Apply replace rules
	for _, rule := range t.config.Replace {
		switch rule.Field {
		case "title":
			article.Title = t.applyReplace(article.Title, rule)
		case "body":
			article.ContentText = t.applyReplace(article.ContentText, rule)
			article.ContentHTML = t.applyReplace(article.ContentHTML, rule)
		}
	}

	// Apply normalization
	article.Title = t.applyNormalization(article.Title, t.config.Normalize.Title)
	article.ContentText = t.applyNormalization(article.ContentText, t.config.Normalize.Body)
	article.ContentHTML = t.applyNormalization(article.ContentHTML, t.config.Normalize.Body)
}

func (t *TransformationEngine) applyReplace(text string, rule ReplaceRule) string {
	if rule.Pattern == "" {
		return text
	}

	if rule.Regex {
		return t.applyRegexReplace(text, rule.Pattern, rule.Replacement, rule.CaseSensitive)
	}
	return t.applyStringReplace(text, rule.Pattern, rule.Replacement, rule.CaseSensitive)
}

func (t *TransformationEngine) applyStringReplace(s, old, new string, caseSensitive bool) string {
	if caseSensitive {
		return strings.ReplaceAll(s, old, new)
	}
	return replaceAllCaseInsensitive(s, old, new)
}

func (t *TransformationEngine) applyRegexReplace(s, pattern, replacement string, caseSensitive bool) string {
	flags := ""
	if !caseSensitive {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + pattern)
	if err != nil {
		return s
	}
	return re.ReplaceAllString(s, replacement)
}

func (t *TransformationEngine) applyNormalization(text string, operations []string) string {
	for _, op := range operations {
		switch op {
		case "trim":
			text = strings.TrimSpace(text)
		case "collapse_spaces":
			text = collapseSpaces(text)
		case "collapse_newlines":
			text = collapseNewlines(text)
		case "remove_empty_paragraphs":
			text = removeEmptyParagraphs(text)
		case "decode_html_entities":
			text = html.UnescapeString(text)
		}
	}
	return text
}

func replaceAllCaseInsensitive(s, old, new string) string {
	if old == "" {
		return s
	}
	lowerS := strings.ToLower(s)
	lowerOld := strings.ToLower(old)
	var result strings.Builder
	start := 0
	for {
		idx := strings.Index(lowerS[start:], lowerOld)
		if idx == -1 {
			result.WriteString(s[start:])
			break
		}
		idx += start
		result.WriteString(s[start:idx])
		result.WriteString(new)
		start = idx + len(old)
	}
	return result.String()
}

func collapseSpaces(s string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(` +`)
	return re.ReplaceAllString(s, " ")
}

func collapseNewlines(s string) string {
	// Replace 3+ newlines with 2 newlines
	re := regexp.MustCompile(`\n{3,}`)
	return re.ReplaceAllString(s, "\n\n")
}

func removeEmptyParagraphs(s string) string {
	// Remove <p></p> and <p> </p>
	re := regexp.MustCompile(`<p>\s*</p>`)
	return re.ReplaceAllString(s, "")
}
