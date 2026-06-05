package app

// app_linkpreview.go — pratinjau tautan: ambil HTML (sisi Go → tanpa CORS),
// parse Open Graph / <title> jadi kartu pratinjau untuk UI.

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// LinkPreviewDTO = metadata satu tautan.
type LinkPreviewDTO struct {
	URL   string `json:"url"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Image string `json:"image"`
}

var (
	reMeta  = regexp.MustCompile(`(?is)<meta[^>]+>`)
	reTitle = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	reProp  = regexp.MustCompile(`(?is)(?:property|name)\s*=\s*["']([^"']+)["']`)
	reCont  = regexp.MustCompile(`(?is)content\s*=\s*["']([^"']*)["']`)
)

// GetLinkPreview mengambil & mem-parse pratinjau tautan. Kosong bila gagal.
func (a *App) GetLinkPreview(rawURL string) *LinkPreviewDTO {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return nil
	}
	ctx, cancel := context.WithTimeout(a.ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; WhatsAppLite/1.0; +link-preview)")
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 512<<10)) // maks 512KB
	if err != nil {
		return nil
	}
	html := string(body)
	og := map[string]string{}
	for _, tag := range reMeta.FindAllString(html, -1) {
		p := reProp.FindStringSubmatch(tag)
		c := reCont.FindStringSubmatch(tag)
		if len(p) == 2 && len(c) == 2 {
			og[strings.ToLower(p[1])] = unescapeHTML(c[1])
		}
	}
	out := &LinkPreviewDTO{URL: rawURL}
	out.Title = firstNonEmpty(og["og:title"], og["twitter:title"])
	if out.Title == "" {
		if m := reTitle.FindStringSubmatch(html); len(m) == 2 {
			out.Title = unescapeHTML(strings.TrimSpace(m[1]))
		}
	}
	out.Desc = firstNonEmpty(og["og:description"], og["twitter:description"], og["description"])
	out.Image = firstNonEmpty(og["og:image"], og["og:image:url"], og["twitter:image"])
	if out.Title == "" && out.Image == "" {
		return nil
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func unescapeHTML(s string) string {
	r := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'", "&#x27;", "'")
	return r.Replace(s)
}
