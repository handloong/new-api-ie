package router

import (
	"html"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

// homePageHTML is the landing page served for GET /. It is a static notice
// pointing humans at the console, deliberately independent of the embedded
// dashboard build so it can be changed without rebuilding the frontend.
const homePageHTML = `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>new-api-ie · server running</title>
<style>
  :root {
    --bg: #f6f7f9;
    --card: #ffffff;
    --text: #16181d;
    --muted: #6b7280;
    --line: #e5e7eb;
    --accent: #2f6df6;
    --accent-soft: rgba(47, 109, 246, 0.10);
    --ok: #12a150;
    --ok-soft: rgba(18, 161, 80, 0.12);
  }
  @media (prefers-color-scheme: dark) {
    :root {
      --bg: #0e1116;
      --card: #151a21;
      --text: #e8eaed;
      --muted: #9aa4b2;
      --line: #232a34;
      --accent: #6f9bff;
      --accent-soft: rgba(111, 155, 255, 0.14);
      --ok: #3ddc84;
      --ok-soft: rgba(61, 220, 132, 0.14);
    }
  }
  * { box-sizing: border-box; }
  html, body { height: 100%; margin: 0; }
  body {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: var(--bg);
    color: var(--text);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", system-ui, sans-serif;
    -webkit-font-smoothing: antialiased;
  }
  body::before {
    content: "";
    position: fixed;
    inset: 0;
    background-image: radial-gradient(circle at 1px 1px, var(--line) 1px, transparent 0);
    background-size: 22px 22px;
    opacity: 0.65;
    pointer-events: none;
  }
  main {
    position: relative;
    width: 100%;
    max-width: 560px;
    background: var(--card);
    border: 1px solid var(--line);
    border-radius: 16px;
    padding: 40px 36px 26px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04), 0 12px 32px -12px rgba(0, 0, 0, 0.12);
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 5px 12px 5px 10px;
    border-radius: 999px;
    background: var(--ok-soft);
    color: var(--ok);
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.01em;
  }
  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: currentColor;
    box-shadow: 0 0 0 0 currentColor;
    animation: pulse 2s ease-out infinite;
  }
  @keyframes pulse {
    0%   { box-shadow: 0 0 0 0 currentColor; opacity: 1; }
    70%  { box-shadow: 0 0 0 7px transparent; opacity: 0.75; }
    100% { box-shadow: 0 0 0 0 transparent; opacity: 1; }
  }
  h1 {
    margin: 22px 0 10px;
    font-size: 28px;
    line-height: 1.2;
    letter-spacing: -0.02em;
  }
  h1 code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: 0.94em;
  }
  p {
    margin: 0 0 8px;
    color: var(--muted);
    font-size: 15px;
    line-height: 1.65;
  }
  .actions {
    margin-top: 28px;
    display: flex;
    align-items: center;
    gap: 14px;
    flex-wrap: wrap;
  }
  a.btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 11px 20px;
    border-radius: 10px;
    background: var(--accent);
    color: #fff;
    font-size: 15px;
    font-weight: 600;
    text-decoration: none;
    transition: transform 0.12s ease, filter 0.12s ease;
  }
  a.btn:hover { filter: brightness(1.06); transform: translateY(-1px); }
  a.btn:active { transform: translateY(0); }
  a.btn svg { width: 16px; height: 16px; }
  .hint { font-size: 13px; color: var(--muted); }
  footer {
    margin-top: 32px;
    padding-top: 16px;
    border-top: 1px solid var(--line);
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 10px 16px;
    flex-wrap: wrap;
    font-size: 12.5px;
    color: var(--muted);
  }
  footer .author {
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    color: var(--text);
    font-weight: 500;
  }
  footer .meta {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: 12px;
  }
  footer .meta .tag {
    padding: 3px 8px;
    border-radius: 6px;
    border: 1px solid var(--line);
    background: var(--accent-soft);
    color: var(--muted);
    white-space: nowrap;
  }
  @media (max-width: 480px) {
    main { padding: 32px 22px 22px; }
    h1 { font-size: 23px; }
    footer { flex-direction: column; align-items: flex-start; }
  }
</style>
</head>
<body>
<main>
  <span class="status"><span class="dot"></span>server running</span>

  <h1><code>new-api-ie</code> 服务运行中</h1>

  <p>网关已启动，API 请求可正常转发。</p>
  <p>管理控制台需要登录后访问。</p>

  <div class="actions">
    <a class="btn" href="/sign-in">
      登录控制台
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M5 12h14M13 6l6 6-6 6"/>
      </svg>
    </a>
    <span class="hint">API 端点：<code>/v1</code></span>
  </div>

  <footer>
    <span>built by <span class="author">Handloong</span></span>
    <span class="meta">
      <span class="tag">{{VERSION}}</span>
      <span class="tag">built {{BUILD_TIME}}</span>
    </span>
  </footer>
</main>
</body>
</html>
`

// ServeHomePage answers GET / with a static notice page and leaves every other
// path to the dashboard asset chain. It sits in front of the embedded frontend so
// the console keeps its own index.html while "/" stays a plain landing page.
func ServeHomePage(c *gin.Context) {
	if c.Request.URL.Path != "/" {
		return
	}
	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(renderHomePage(common.Version, resolveBuildTime())))
	c.Abort()
}

// renderHomePage fills the page placeholders. Values are HTML-escaped because
// the version can come from the VERSION environment variable.
func renderHomePage(version string, buildTime time.Time) string {
	buildTimeText := "unknown"
	if !buildTime.IsZero() {
		buildTimeText = buildTime.UTC().Format("2006-01-02 15:04") + " UTC"
	}
	return strings.NewReplacer(
		"{{VERSION}}", html.EscapeString(version),
		"{{BUILD_TIME}}", buildTimeText,
	).Replace(homePageHTML)
}

// resolveBuildTime prefers the ldflags-injected timestamp and falls back to the
// VCS commit time recorded in the Go build info, which covers local builds.
func resolveBuildTime() time.Time {
	raw := common.BuildTime
	if raw == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, s := range info.Settings {
				if s.Key == "vcs.time" {
					raw = s.Value
					break
				}
			}
		}
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
