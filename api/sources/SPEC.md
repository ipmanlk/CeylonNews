# Source Configuration Specification v2.0

Each source is defined in a `.toml` file that describes a **pipeline** for discovering, extracting, validating, and transforming news articles.

## Pipeline Stages

```
1. DISCOVERY      → Find article URLs (from RSS feed or HTML listing)
2. EXTRACTION     → Extract content fields (title, body, image, date)
3. VALIDATION     → Check if article meets criteria (skip if fails)
4. TRANSFORMATION → Clean/modify extracted content
5. STORAGE        → Save to database
```

**Key Principles:**
- Each language is **independently configured** (different URLs, selectors, rules)
- Validation runs **early** when possible (before full article fetch)
- URL transformation is **composable** (chain multiple rules)
- Rules are **explicit** and **self-documenting**

---

## Configuration Structure

### Top Level

```toml
name = "Source Name"

# Optional: Shared configuration template
[shared.discovery]
# ... discovery config

[shared.extraction]
# ... extraction config

# Per-language configurations
[[languages]]
language = "en"
max_items = 5
shared = "default"  # optional: reference shared config

[languages.discovery]
# ... overrides/extends shared config
```

### Language Configuration

Each `[[languages]]` block defines a complete pipeline for one language.

```toml
[[languages]]
language = "en"  # "en", "si", or "ta"
max_items = 5     # Maximum articles to scrape per run
shared = "default"  # Optional: inherit from [shared.<name>]
```

---

## 1. Discovery Phase

**Purpose:** Find article URLs to scrape

### RSS Discovery

```toml
[languages.discovery]
type = "rss"
url = "https://example.com/feed.xml"
browser = false  # Use headless browser if feed requires JS
```

### HTML Discovery

```toml
[languages.discovery]
type = "html"
url = "https://example.com/news"
browser = false

[languages.discovery.html]
# CSS selectors to extract links (evaluated in order, combined)
link_selectors = ["a.article-link", "div.news a"]

# Optional: Extract title from listing page (enables early validation)
title_selector = "h2.article-title"

# Optional: Extract publish date from listing
date_selector = "time.published"

# URL transformation pipeline (executed in order)
[[languages.discovery.html.url_rules]]
type = "filter_prefix"
value = "/news/"
mode = "any"  # "any" or "all" (default: "any")

[[languages.discovery.html.url_rules]]
type = "prepend"
value = "https://example.com"
condition = "if_relative"  # "if_relative", "if_protocol_relative", "always"
```

#### URL Rule Types

**Filtering Rules** (remove URLs):
- `filter_prefix`: Keep URLs starting with value
- `filter_not_prefix`: Remove URLs starting with value
- `filter_contains`: Keep URLs containing value
- `filter_not_contains`: Remove URLs containing value
- `filter_regex`: Keep URLs matching pattern

**Filter Mode** (for multiple filters of same type):
- `mode = "any"` (default): Keep URL if it matches ANY rule
- `mode = "all"`: Keep URL only if it matches ALL rules

**Transformation Rules** (modify URLs):
- `prepend`: Add value to start
  - `condition = "if_relative"`: Only if starts with `/`
  - `condition = "if_protocol_relative"`: Only if starts with `//`
  - `condition = "always"`: Always prepend
- `append`: Add value to end
- `regex_replace`: Replace `pattern` with `replacement`
- `normalize`: Standard cleanup (lowercase protocol, remove trailing slash, etc.)

**Execution Order:** Rules execute sequentially as defined. Filtering happens first, then transformations.

---

## 2. Extraction Phase

**Purpose:** Extract article content from URL

```toml
[languages.extraction]
browser = false  # Use headless browser if article requires JS

[languages.extraction.content]
# Option A: Scope trafilatura to specific element
scope_selector = ".article-content"

# Option B: Override individual field extraction
title_selector = "h1.headline"
body_selector = ".article-body"
image_selector = "img.featured-image"
date_selector = "time[datetime]"

# If no selectors specified, trafilatura auto-extracts from full page
```

**Field Priority:**
1. If field selector specified → use selector
2. Else if scope_selector specified → trafilatura within scope
3. Else → trafilatura on full page
4. For RSS: title/date from feed item takes precedence

---

## 3. Validation Phase

**Purpose:** Filter out unwanted articles

**Timing:**
- If title extracted during Discovery → validate immediately, skip full fetch if fails
- Otherwise → validate after Extraction

```toml
[languages.validation]

# Skip article if ANY rule matches
[[languages.validation.skip]]
field = "title"       # "title", "body", "any"
type = "contains"
value = "Error"
case_sensitive = false

[[languages.validation.skip]]
field = "body"
type = "regex"
pattern = "paywall|subscription required"
case_sensitive = false

# Article must pass ALL require rules
[[languages.validation.require]]
field = "title"
type = "min_length"
value = 10

[[languages.validation.require]]
field = "body"
type = "not_contains"
value = "advertisement"
case_sensitive = false
```

### Validation Rule Types

**Skip Rules** (any match → discard article):
- `contains`: Field contains substring
- `not_contains`: Field does not contain substring
- `regex`: Field matches regex pattern
- `equals`: Field exactly equals value
- `not_equals`: Field does not equal value
- `min_length`: Field shorter than value
- `max_length`: Field longer than value

**Require Rules** (all must pass → keep article):
Same types as skip rules, but all must be true.

**Field Options:**
- `title`: Check only title field
- `body`: Check only body content
- `any`: Check if pattern appears in either title or body

---

## 4. Transformation Phase

**Purpose:** Clean and normalize extracted content

```toml
[languages.transformation]

# Text replacement rules
[[languages.transformation.replace]]
field = "title"  # "title" or "body"
pattern = " - Breaking News.*$"
replacement = ""
regex = false
case_sensitive = false

[[languages.transformation.replace]]
field = "body"
pattern = "<div class=\"ad\">.*?</div>"
replacement = ""
regex = true

# Standard normalization operations
[languages.transformation.normalize]
title = ["trim", "collapse_spaces"]
body = ["trim", "collapse_spaces", "remove_empty_paragraphs"]
```

### Transformation Options

**Replace Rules:**
- `field`: "title" or "body"
- `pattern`: Text to find (string or regex if `regex = true`)
- `replacement`: Text to replace with
- `regex`: Treat pattern as regex (default: false)
- `case_sensitive`: Case-sensitive matching (default: false)

**Normalize Operations:**
- `trim`: Remove leading/trailing whitespace
- `collapse_spaces`: Convert multiple spaces to single space
- `collapse_newlines`: Convert multiple newlines to single newline
- `remove_empty_paragraphs`: Remove `<p></p>` or `<p> </p>`
- `decode_html_entities`: Convert `&amp;` → `&`, etc.

---

## Shared Configuration

For sources where multiple languages use similar config:

```toml
name = "Example Source"

# Define reusable templates
[shared.discovery.default]
type = "html"
browser = false

[shared.discovery.default.html]
link_selectors = ["a.article-link"]

[[shared.discovery.default.html.url_rules]]
type = "filter_prefix"
value = "/news/"

# Use in languages
[[languages]]
language = "en"
max_items = 5
shared = "default"  # Inherits from [shared.discovery.default]

[languages.discovery]
url = "https://example.com/en/news"  # Override specific fields

[[languages]]
language = "si"
max_items = 5
shared = "default"

[languages.discovery]
url = "https://example.com/si/news"  # Only URL differs
```

**Inheritance Rules:**
- Fields in `[languages.*]` override shared config
- Arrays are **replaced**, not merged
- Nested objects are **deep merged**

---

## Complete Examples

### Example 1: Simple RSS Source

```toml
name = "Lankadeepa"

[[languages]]
language = "si"
max_items = 5

[languages.discovery]
type = "rss"
url = "https://www.lankadeepa.lk/rss/latest_news/1"
```

### Example 2: RSS with Validation & Transformation

```toml
name = "Daily Mirror"

[[languages]]
language = "en"
max_items = 5

[languages.discovery]
type = "rss"
url = "https://www.dailymirror.lk/rss/todays_headlines/419"

[languages.extraction]
browser = true

[languages.validation]
[[languages.validation.skip]]
field = "title"
type = "contains"
value = "An Error Was"
case_sensitive = false

[languages.transformation]
[[languages.transformation.replace]]
field = "title"
pattern = " - Breaking News | Daily Mirror"
replacement = ""
case_sensitive = false

[languages.transformation.normalize]
title = ["trim", "collapse_spaces"]
```

### Example 3: HTML with Different URL Rules per Language

```toml
name = "BBC"

[[languages]]
language = "en"
max_items = 5

[languages.discovery]
type = "html"
url = "https://www.bbc.com/news/topics/cywd23g0gxgt"

[languages.discovery.html]
link_selectors = ["a"]

[[languages.discovery.html.url_rules]]
type = "filter_prefix"
value = "/news/articles/"

[[languages.discovery.html.url_rules]]
type = "prepend"
value = "https://www.bbc.com"
condition = "if_relative"

[[languages]]
language = "si"
max_items = 5

[languages.discovery]
type = "html"
url = "https://www.bbc.com/sinhala/topics/cg7267dz901t"

[languages.discovery.html]
link_selectors = ["a"]

# Different filter - absolute URL
[[languages.discovery.html.url_rules]]
type = "filter_prefix"
value = "https://www.bbc.com/sinhala/articles/"

[[languages]]
language = "ta"
max_items = 5

[languages.discovery]
type = "html"
url = "https://www.bbc.com/tamil/topics/cz74k7p3qw7t"

[languages.discovery.html]
link_selectors = ["a"]

[[languages.discovery.html.url_rules]]
type = "filter_prefix"
value = "https://www.bbc.com/tamil/articles/"
```

### Example 4: Complex Multi-Stage with Shared Config

```toml
name = "Hiru News"

# Shared config for all languages
[shared.discovery.default]
type = "html"
browser = false

[shared.discovery.default.html]
link_selectors = ["a.card-featured", "a.card-v1"]

[[shared.discovery.default.html.url_rules]]
type = "filter_prefix"
value = "https://hirunews.lk/"

[[languages]]
language = "si"
max_items = 5
shared = "default"

[languages.discovery]
url = "https://www.hirunews.lk"

[[languages]]
language = "en"
max_items = 5
shared = "default"

[languages.discovery]
url = "https://www.hirunews.lk/en/"

[[languages]]
language = "ta"
max_items = 5
shared = "default"

[languages.discovery]
url = "https://www.hirunews.lk/tm/"
```

### Example 5: Different Extraction per Language

```toml
name = "News.lk"

[[languages]]
language = "en"
max_items = 2

[languages.discovery]
type = "html"
url = "https://news.lk/news/"
browser = true

[languages.discovery.html]
link_selectors = ["article.item h2 a"]
title_selector = "article.item h2"

[[languages.discovery.html.url_rules]]
type = "filter_prefix"
value = "/news/"

[[languages.discovery.html.url_rules]]
type = "prepend"
value = "https://news.lk"
condition = "if_relative"

[languages.extraction]
browser = true

[languages.extraction.content]
scope_selector = ".article-main-en"  # English-specific structure

[languages.validation]
[[languages.validation.skip]]
field = "title"
type = "contains"
value = "Premium Members Only"

[[languages.validation.require]]
field = "body"
type = "min_length"
value = 200

[[languages]]
language = "si"
max_items = 2

[languages.discovery]
type = "html"
url = "https://sinhala.news.lk/news/"
browser = true

[languages.discovery.html]
link_selectors = ["article.item h2 a"]

[[languages.discovery.html.url_rules]]
type = "filter_prefix"
value = "/news/"

[[languages.discovery.html.url_rules]]
type = "prepend"
value = "https://sinhala.news.lk"
condition = "if_relative"

[languages.extraction]
browser = true

[languages.extraction.content]
scope_selector = ".article-main-si"  # Sinhala-specific structure (different!)

[languages.validation]
[[languages.validation.require]]
field = "body"
type = "min_length"
value = 200
```

---

## Migration Guide

### Breaking Changes from v1

1. **`[title_transform]` removed** → Use `[languages.validation]` and `[languages.transformation]`
2. **`[languages.listing]` renamed** → Now `[languages.discovery]`
3. **`[languages.article]` renamed** → Now `[languages.extraction]`
4. **`url_prefix` + `base_url` removed** → Use `[languages.discovery.html.url_rules]`
5. **`selector` in article** → Now `[languages.extraction.content.scope_selector]`

### Conversion Example

**Old (v1):**
```toml
[title_transform]
skip = [{ contains = "Error", case_sensitive = false }]
replace = [{ pattern = " - News", with = "", case_sensitive = false }]

[[languages]]
language = "en"
max_items = 5

[languages.listing]
type = "html"
url = "https://example.com"
selectors = ["a.link"]
url_prefix = "/news/"
base_url = "https://example.com"

[languages.article]
browser = true
selector = ".content"
```

**New (v2):**
```toml
[[languages]]
language = "en"
max_items = 5

[languages.discovery]
type = "html"
url = "https://example.com"

[languages.discovery.html]
link_selectors = ["a.link"]

[[languages.discovery.html.url_rules]]
type = "filter_prefix"
value = "/news/"

[[languages.discovery.html.url_rules]]
type = "prepend"
value = "https://example.com"
condition = "if_relative"

[languages.extraction]
browser = true

[languages.extraction.content]
scope_selector = ".content"

[languages.validation]
[[languages.validation.skip]]
field = "title"
type = "contains"
value = "Error"
case_sensitive = false

[languages.transformation]
[[languages.transformation.replace]]
field = "title"
pattern = " - News"
replacement = ""
case_sensitive = false
```

---

## Implementation Notes

### URL Rule Execution Order

Rules execute sequentially in the order defined:
1. All filtering rules execute first (removing URLs)
2. Then transformation rules execute (modifying remaining URLs)

### Validation Timing

1. **Early validation**: If `title_selector` is defined in `[languages.discovery.html]`, validation runs immediately after title extraction, before full article fetch
2. **Late validation**: If no early title available, validation runs after full extraction

### Shared Config Inheritance

When `shared = "name"` is specified:
1. Start with shared config values
2. Override with language-specific values
3. Arrays are **replaced** (not merged)
4. Nested tables are **deep merged**

Example:
```toml
[shared.discovery.default.html]
link_selectors = ["a"]

[[shared.discovery.default.html.url_rules]]
type = "filter_prefix"
value = "/news/"

[[languages]]
language = "en"
shared = "default"

[languages.discovery.html]
# This REPLACES shared link_selectors
link_selectors = ["a.article"]

[[languages.discovery.html.url_rules]]
# This is ADDED to shared url_rules
# (because url_rules is an array of tables, each [[...]] creates new entry)
type = "filter_contains"
value = "article"
```
