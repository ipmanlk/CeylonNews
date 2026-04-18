# Source Configuration Specification v2.0

Each source is defined in a `.toml` file that describes a **pipeline** for discovering, extracting, validating, and transforming news articles.

## Source Identification

Each source **must** have a unique, stable identifier (`id`) that never changes, separate from the human-readable display name (`name`).

```toml
id = "daily-mirror"    # Stable identifier (kebab-case, never changes)
name = "Daily Mirror"  # Display name (can change anytime)
```

### ID Format (Kebab-Case)

The `id` field **must** follow these rules:

- **Required**: Must be present and non-empty
- **Start with letter**: Must begin with a lowercase letter (`a-z`)
- **Allowed characters**: Only lowercase letters (`a-z`), numbers (`0-9`), and hyphens (`-`)
- **No consecutive hyphens**: `--` is not allowed
- **No leading/trailing hyphens**: Cannot start or end with a hyphen
- **Format regex**: `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`

**Valid examples:**
- `bbc`
- `daily-mirror`
- `hiru-news`
- `news-lk`
- `lanka-deepa-123`

**Invalid examples:**
- `BBC` (uppercase not allowed)
- `daily_mirror` (underscore not allowed)
- `hiru--news` (consecutive hyphens)
- `-hiru` (leading hyphen)
- `hiru-` (trailing hyphen)
- `123-news` (must start with letter)
- `hiru news` (spaces not allowed)

### Validation Behavior

At startup, the system validates all source configurations:
- Sources with **invalid or missing IDs** are **skipped** with an error log
- Sources with **duplicate IDs** cause the application to fail startup
- The `id` is used for all internal references (database, API filters)
- The `name` is only used for display purposes in API responses

### Migration Notes

When changing a source's display name:
- **Safe**: Update the `name` field - existing articles will show the new name immediately
- **Never change**: The `id` field - this would break all existing article references

If you must change an ID, treat it as creating a new source and deprecating the old one.

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

# Link selectors (object format - one per entry)
[[languages.discovery.html.link_selectors]]
link = "a.article-link"

[[languages.discovery.html.link_selectors]]
link = "div.news a"

# With title extraction for early validation:
[[languages.discovery.html.link_selectors]]
link = "h2 a"
title = "parent:h2"  # Extract title from parent h2 element

[[languages.discovery.html.link_selectors]]
link = ".card a"
title = "sibling:h3"  # Extract title from sibling h3 element

# URL transformation pipeline (executed in order)
[[languages.discovery.html.url_rules]]
type = "filter_prefix"
value = "/news/"
match_policy = "any"  # "any" or "all" (default: "any")

[[languages.discovery.html.url_rules]]
type = "prepend"
value = "https://example.com"
condition = "if_relative"  # "if_relative", "if_protocol_relative", "always"
```

#### Link Selectors

All link selectors use the object format:

```toml
[[languages.discovery.html.link_selectors]]
link = "h2 a"
title = "parent:h2"  # Optional: enables early validation
```

**Title selector prefixes:**
- `"self"` or `"."`: Title is the link text itself
- `"parent:<selector>"`: Title is found in the parent element matching `<selector>`
- `"sibling:<selector>"`: Title is found in a sibling element matching `<selector>`
- `"container:<container_selector> <title_selector>"`: Title is found by `<title_selector>` within the closest `<container_selector>`

**Examples:**
```toml
# Title is the link text itself
[[languages.discovery.html.link_selectors]]
link = "h2 a"
title = "self"

# Title is in the parent h2 element
[[languages.discovery.html.link_selectors]]
link = "h2 a"
title = "parent:h2"

# Title is in a sibling element
[[languages.discovery.html.link_selectors]]
link = ".card a.read-more"
title = "sibling:h3"

# Title is in a specific container (find closest article, then h2 within it)
[[languages.discovery.html.link_selectors]]
link = "article.item a"
title = "container:article.item h2"
```

#### URL Rule Types

**Filtering Rules** (remove URLs):
- `filter_prefix`: Keep URLs starting with value
- `filter_not_prefix`: Remove URLs starting with value
- `filter_contains`: Keep URLs containing value
- `filter_not_contains`: Remove URLs containing value
- `filter_regex`: Keep URLs matching pattern

**Filter Match Policy** (for multiple filters of same type):
- `match_policy = "any"` (default): Keep URL if it matches ANY rule of this type (OR logic)
- `match_policy = "all"`: Keep URL only if it matches ALL rules of this type (AND logic)

**Transformation Rules** (modify URLs):
- `prepend`: Add value to start
  - `condition = "if_relative"`: Only if starts with `/`
  - `condition = "if_protocol_relative"`: Only if starts with `//`
  - `condition = "always"`: Always prepend
- `append`: Add value to end
- `regex_replace`: Replace `pattern` with `replacement`
- `normalize`: Standard cleanup (lowercase protocol, remove trailing slash, etc.)

**Execution Order:** URL rules execute in two phases:
1. **Filtering phase**: All filtering rules execute first, grouped by type with `match_policy` applied within each group
2. **Transformation phase**: Transformation rules execute sequentially in definition order

This means all filters run before any transformations, regardless of their order in the config file.

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

# Option C: Remove unwanted HTML elements before extraction
prune_selector = ".advertisement, .social-share, aside.related-articles"

# If no selectors specified, trafilatura auto-extracts from full page
```

**Field Priority:**

The extraction system uses the following approach:

1. **Individual field selectors** (`title_selector`, `image_selector`, `date_selector`) extract those specific fields directly from the HTML using CSS selectors
2. **Content extraction** uses either:
   - `scope_selector`: Trafilatura extracts content within this element only
   - `body_selector`: Direct extraction from this element
   - Or full page extraction if neither is specified
3. **Pruning**: If `prune_selector` is specified, matching elements are removed before content extraction
4. **RSS precedence**: For RSS discovery, title and date from the feed item take precedence over extracted values

These approaches can be combined. For example, you can use `title_selector` to extract the title directly while using `scope_selector` for the body content.

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
- `min_length`: Field length must be at least value (≥)
- `max_length`: Field length must be at most value (≤)

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

[[shared.discovery.default.html.link_selectors]]
link = "a.article-link"

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
- Arrays are **replaced**, not merged (except `url_rules` which is **appended** to shared rules)
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

[[languages.discovery.html.link_selectors]]
link = "a"

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

[[languages.discovery.html.link_selectors]]
link = "a"

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

[[languages.discovery.html.link_selectors]]
link = "a"

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

[[shared.discovery.default.html.link_selectors]]
link = "a.card-featured"

[[shared.discovery.default.html.link_selectors]]
link = "a.card-v1"

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

### Example 5: Different Extraction per Language with Early Validation

```toml
name = "News.lk"

[[languages]]
language = "en"
max_items = 2

[languages.discovery]
type = "html"
url = "https://news.lk/news/"
browser = true

# Link selector with title extraction for early validation
[[languages.discovery.html.link_selectors]]
link = "article.item h2 a"
title = "parent:h2"

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

# No title extraction for this language
[[languages.discovery.html.link_selectors]]
link = "article.item h2 a"

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

### Example 6: Remove Unwanted Content with Prune Selector

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

[languages.extraction.content]
# Remove ads, social widgets, and related articles before content extraction
prune_selector = ".advertisement, .social-share-buttons, aside.related-articles, .newsletter-signup"

[languages.validation]
[[languages.validation.skip]]
field = "title"
type = "contains"
value = "An Error Was"
case_sensitive = false
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

URL rules execute in two phases:

1. **Filtering phase**: All filtering rules execute first, grouped by type. Within each group, `match_policy` determines how multiple rules combine:
   - `match_policy = "any"` (default): URL passes if it matches ANY rule in the group (OR logic)
   - `match_policy = "all"`: URL passes only if it matches ALL rules in the group (AND logic)

2. **Transformation phase**: Transformation rules execute sequentially in the order defined

This two-phase approach means all filters run before any transformations, regardless of their order in the config file.

### Validation Timing

1. **Early validation**: If any `link_selectors` entry has a `title` field, validation runs immediately after title extraction from the listing page, before full article fetch. The title is extracted based on the `title` selector pattern (parent:, sibling:, self, or container-based).
2. **Late validation**: If no early title available, validation runs after full extraction

### Shared Config Inheritance

When `shared = "name"` is specified:
1. Start with shared config values
2. Override with language-specific values
3. Arrays are **replaced** (not merged), except `url_rules` which is **appended**
4. Nested tables are **deep merged**

### Default Values

Optional fields have the following defaults:

| Field | Default | Description |
|-------|---------|-------------|
| `browser` | `false` | Use HTTP client, not headless browser |
| `max_items` | `0` (unlimited) | No limit on articles if omitted or 0 |
| `match_policy` | `"any"` | URL passes if it matches any filter of that type |
| `condition` (prepend) | `"always"` | Always prepend the value |
| `case_sensitive` | `false` | Case-insensitive matching |
| `regex` | `false` | Treat pattern as literal string |
| `shared` | `""` (none) | No shared config inheritance |

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
