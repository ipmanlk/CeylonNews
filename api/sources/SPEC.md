# Source Configuration Specification

Each source is a `.toml` file in the `sources/` directory.

## Fields

### Top-level

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Display name for the source |
| `title_transform` | table | no | Rules for filtering/rewriting titles (see below) |
| `[[languages]]` | array | yes | One entry per supported language |

### `[[languages]]`

| Field | Type | Required | Description |
|---|---|---|---|
| `language` | string | yes | Language code: `en`, `si`, or `ta` |
| `max_items` | int | yes | Max articles to scrape per run |
| `[languages.listing]` | table | yes | How to retrieve the article list |
| `[languages.article]` | table | no | How to fetch each article (defaults: HTTP, no selector) |

### `[languages.listing]`

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | yes | `"rss"` or `"html"` |
| `url` | string | yes | Feed URL (rss) or listing page URL (html) |
| `browser` | bool | no (default: false) | Use headless browser to fetch the listing |
| `selectors` | []string | html only | CSS selectors for extracting article links |
| `url_prefix` | string | html only | Keep only links starting with this prefix |
| `base_url` | string | html only | Prepend to relative links (those starting with `/`) |

### `[languages.article]`

| Field | Type | Required | Description |
|---|---|---|---|
| `browser` | bool | no (default: false) | Use headless browser to fetch each article page |
| `selector` | string | no | Scope content extraction to this CSS element |

### `[title_transform]`

| Field | Description |
|---|---|
| `skip` | Array of `{ contains, case_sensitive }` — skip articles whose title matches |
| `replace` | Array of `{ pattern, with, case_sensitive }` — rewrite matching text in titles |

## Quick Reference

| Need | Config |
|---|---|
| Standard RSS | `listing.type = "rss"` |
| RSS feed needs JS to load | `listing.browser = true` |
| Article page needs JS | `article.browser = true` |
| Article content in specific element | `article.selector = ".css-selector"` |
| HTML listing page | `listing.type = "html"` + `selectors` + `url_prefix` |
| HTML listing page needs JS | `listing.type = "html"` + `listing.browser = true` |
| Relative article links | `listing.base_url = "https://..."` |

## Examples

### Simple RSS Source

```toml
name = "Lankadeepa"

[[languages]]
language = "si"
max_items = 5

[languages.listing]
type = "rss"
url = "https://www.lankadeepa.lk/rss/latest_news/1"
```

### RSS with Browser for Articles

```toml
name = "Daily Mirror"

[title_transform]
skip = [{ contains = "An Error Was", case_sensitive = false }]
replace = [{ pattern = " - Breaking News | Daily Mirror", with = "", case_sensitive = false }]

[[languages]]
language = "en"
max_items = 5

[languages.listing]
type = "rss"
url = "https://www.dailymirror.lk/rss/todays_headlines/419"

[languages.article]
browser = true
```

### RSS with Content Selector

```toml
name = "The Island"

[[languages]]
language = "en"
max_items = 5

[languages.listing]
type = "rss"
url = "https://island.lk/feed/"

[languages.article]
browser = true
selector = ".mvp-post-soc-out.right.relative"
```

### HTML Listing

```toml
name = "Hiru News"

[[languages]]
language = "si"
max_items = 5

[languages.listing]
type = "html"
url = "https://www.hirunews.lk"
selectors = ["a.card-featured", "a.card-v1"]
url_prefix = "https://hirunews.lk/"
```

### Multi-Language Source

```toml
name = "BBC"

[[languages]]
language = "en"
max_items = 5

[languages.listing]
type = "html"
url = "https://www.bbc.com/news/topics/cywd23g0gxgt"
selectors = ["a"]
url_prefix = "/news/articles/"
base_url = "https://www.bbc.com"

[[languages]]
language = "si"
max_items = 5

[languages.listing]
type = "html"
url = "https://www.bbc.com/sinhala/topics/cg7267dz901t"
selectors = ["a"]
url_prefix = "https://www.bbc.com/sinhala/articles/"
```

### Browser for Both Listing and Articles

```toml
name = "News.lk"

[[languages]]
language = "en"
max_items = 2

[languages.listing]
type = "html"
url = "https://news.lk/news/"
selectors = ["article.item h2 a"]
url_prefix = "/news/"
base_url = "https://news.lk"
browser = true

[languages.article]
browser = true
```
