from flask import Flask, request, jsonify
import undetected_chromedriver as uc
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
import time
import logging

app = Flask(__name__)

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class BrowserScraper:
    def __init__(self):
        self.driver = None

    def get_driver(self):
        if self.driver is None:
            options = uc.ChromeOptions()
            options.add_argument("--no-sandbox")
            options.add_argument("--disable-dev-shm-usage")
            options.add_argument("--disable-gpu")
            options.add_argument("--disable-images")
            options.add_argument("--disable-css")
            options.add_argument("--headless")
            options.add_argument("--window-size=1920,1080")
            options.add_argument(
                "--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
            )

            # Block CSS and images
            options.add_argument("--disable-javascript-images")
            options.add_argument("--disable-plugins")
            options.add_argument("--disable-extensions")
            options.add_argument("--disable-web-security")
            options.add_argument("--disable-features=VizDisplayCompositor")

            prefs = {
                "profile.managed_default_content_settings.images": 2,  # Block images
                "profile.default_content_setting_values.images": 2,  # Block images
                "profile.managed_default_content_settings.stylesheets": 2,  # Block CSS
                "profile.default_content_setting_values.stylesheets": 2,  # Block CSS
                "profile.managed_default_content_settings.media_stream": 2,  # Block media
                "profile.default_content_setting_values.media_stream": 2,  # Block media
            }
            options.add_experimental_option("prefs", prefs)

            self.driver = uc.Chrome(options=options, version_main=146)
            self.driver.implicitly_wait(10)
        return self.driver

    def scrape(self, url, wait_time=15):
        driver = self.get_driver()

        try:
            logger.info(f"Scraping URL: {url}")

            # Check if this is an RSS feed
            is_rss = any(
                rss_indicator in url.lower() for rss_indicator in ["rss", "feed", "xml"]
            )

            if is_rss:
                # For RSS feeds, get the raw response content
                driver.get(url)
                time.sleep(wait_time)

                # Get the raw response content using Chrome DevTools
                raw_content = driver.execute_script(
                    """
                    return new Promise((resolve) => {
                        const xhr = new XMLHttpRequest();
                        xhr.open('GET', arguments[0], true);
                        xhr.setRequestHeader('Accept', 'application/rss+xml, application/xml, text/xml, */*');
                        xhr.onload = function() {
                            resolve(xhr.responseText);
                        };
                        xhr.send();
                    });
                """,
                    url,
                )

                # Wait a bit for the XHR to complete
                time.sleep(5)

                if raw_content and len(raw_content) > 100:
                    return {"success": True, "html": raw_content, "url": url}

            # For regular pages, use browser with network interception
            driver.get(url)

            # Set up network interception to block CSS and images
            driver.execute_cdp_cmd("Network.setBypassServiceWorker", {"bypass": True})
            driver.execute_cdp_cmd("Network.enable", {})

            # Block CSS and image requests
            driver.execute_script("""
                // Intercept and block CSS and image requests
                const originalFetch = window.fetch;
                window.fetch = function(url, options) {
                    const urlStr = url.toString().toLowerCase();
                    if (urlStr.includes('.css') || urlStr.includes('.jpg') || urlStr.includes('.jpeg') || 
                        urlStr.includes('.png') || urlStr.includes('.gif') || urlStr.includes('.webp') ||
                        urlStr.includes('.svg') || urlStr.includes('image/') || urlStr.includes('stylesheet')) {
                        return Promise.reject(new Error('Blocked resource'));
                    }
                    return originalFetch.apply(this, arguments);
                };
                
                // Block CSS and image elements from loading
                const observer = new MutationObserver(function(mutations) {
                    mutations.forEach(function(mutation) {
                        mutation.addedNodes.forEach(function(node) {
                            if (node.nodeType === 1) { // Element node
                                if (node.tagName === 'LINK' && node.rel === 'stylesheet') {
                                    node.disabled = true;
                                    node.remove();
                                }
                                if (node.tagName === 'IMG') {
                                    node.style.display = 'none';
                                    node.src = '';
                                }
                                if (node.tagName === 'STYLE') {
                                    node.disabled = true;
                                    node.remove();
                                }
                            }
                        });
                    });
                });
                
                observer.observe(document, {
                    childList: true,
                    subtree: true
                });
                
                // Block existing CSS and images
                document.querySelectorAll('link[rel="stylesheet"]').forEach(link => {
                    link.disabled = true;
                    link.remove();
                });
                
                document.querySelectorAll('img').forEach(img => {
                    img.style.display = 'none';
                    img.src = '';
                });
                
                document.querySelectorAll('style').forEach(style => {
                    style.disabled = true;
                    style.remove();
                });
            """)

            # Wait for page to load
            time.sleep(wait_time)

            # Get the HTML
            html = driver.page_source

            return {"success": True, "html": html, "url": url}

        except Exception as e:
            logger.error(f"Error scraping {url}: {str(e)}")
            return {"success": False, "error": str(e), "url": url}


# Initialize scraper
scraper = BrowserScraper()


@app.route("/")
def home():
    return jsonify(
        {
            "message": "Web Scraping API",
            "endpoint": "/scrape",
            "method": "GET",
            "params": {
                "url": "URL to scrape (required)",
                "wait_time": "Seconds to wait for page load (default: 3)",
            },
            "example": "/scrape?url=https://example.com&wait_time=5",
        }
    )


@app.route("/scrape", methods=["GET"])
def scrape():
    try:
        url = request.args.get("url")
        wait_time = int(request.args.get("wait_time", 10))

        if not url:
            return jsonify({"error": "URL parameter is required"}), 400

        result = scraper.scrape(url, wait_time)

        if result["success"]:
            return jsonify(result)
        else:
            return jsonify(result), 500

    except Exception as e:
        logger.error(f"API error: {str(e)}")
        return jsonify({"error": str(e)}), 500


@app.route("/health")
def health():
    return jsonify({"status": "healthy", "browser_ready": scraper.driver is not None})


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8000, debug=False)
