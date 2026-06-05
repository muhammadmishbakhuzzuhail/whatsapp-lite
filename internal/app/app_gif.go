package app

// app_gif.go — pencarian GIF & stiker lewat Tenor DARI SISI GO. WebKitGTK sering
// memblok fetch() lintas-asal (CORS) sehingga picker kosong; menarik dari backend
// menghindari itu. FE menampilkan <img src=preview> lalu mengunduh media penuh
// saat dipilih via FetchRemoteMedia.
//
// Pagination: Tenor mengembalikan kursor `next` (string). FE kirim balik sbg
// `pos` utk halaman berikutnya → infinite scroll (bukan lagi cap 24 sekali muat).

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

// GifDTO = satu hasil (URL preview + URL media penuh utk dikirim).
type GifDTO struct {
	ID      string `json:"id"`
	Preview string `json:"preview"`
	Mp4     string `json:"mp4"`
}

// GifPage = satu halaman hasil + kursor `next` (kosong = habis).
type GifPage struct {
	Items []GifDTO `json:"items"`
	Next  string   `json:"next"`
}

// tenorKey = key demo publik anonim Tenor v1 (tak perlu pengguna daftar).
const tenorKey = "LIVDSRZULELA"
const tenorLimit = 50 // maks per halaman (Tenor)

var tenorHTTP = &http.Client{Timeout: 15 * time.Second}

// tenorResp = bentuk respons Tenor v1 yang kita pakai (results + next).
type tenorResp struct {
	Next    string `json:"next"`
	Results []struct {
		ID    string `json:"id"`
		Media []map[string]struct {
			URL string `json:"url"`
		} `json:"media"`
	} `json:"results"`
}

// tenorFetch menjalankan satu query Tenor (trending bila query kosong) dgn
// extra param (mis. searchfilter=sticker) + kursor pos. Mengembalikan respons mentah.
func (a *App) tenorFetch(query, pos, extra string) (*tenorResp, bool) {
	endpoint := "https://g.tenor.com/v1/trending"
	if query != "" {
		endpoint = "https://g.tenor.com/v1/search"
	}
	u := endpoint + "?key=" + tenorKey + "&limit=" + itoa(tenorLimit) + "&contentfilter=high" + extra
	if query != "" {
		u += "&q=" + url.QueryEscape(query)
	}
	if pos != "" {
		u += "&pos=" + url.QueryEscape(pos)
	}
	ctx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; WhatsAppLite/1.0)")
	resp, err := tenorHTTP.Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, false
	}
	var body tenorResp
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, false
	}
	return &body, true
}

// SearchGifs mengembalikan satu halaman GIF (trending / hasil cari) + kursor next.
func (a *App) SearchGifs(query, pos string) GifPage {
	page := GifPage{Items: []GifDTO{}}
	body, ok := a.tenorFetch(query, pos, "&media_filter=minimal")
	if !ok {
		return page
	}
	page.Next = body.Next
	for _, r := range body.Results {
		if len(r.Media) == 0 {
			continue
		}
		m := r.Media[0]
		preview := first(m["tinygif"].URL, m["nanogif"].URL)
		mp4 := first(m["mp4"].URL, m["tinymp4"].URL)
		if preview == "" || mp4 == "" {
			continue
		}
		page.Items = append(page.Items, GifDTO{ID: r.ID, Preview: preview, Mp4: mp4})
	}
	return page
}

// SearchStickers mengembalikan satu halaman stiker TRANSPARAN (searchfilter=sticker)
// + kursor next. Preview = format kecil transparan; Mp4 (URL unduh) = webp/gif
// transparan penuh utk dikirim sbg stiker.
func (a *App) SearchStickers(query, pos string) GifPage {
	page := GifPage{Items: []GifDTO{}}
	body, ok := a.tenorFetch(query, pos, "&searchfilter=sticker")
	if !ok {
		return page
	}
	page.Next = body.Next
	for _, r := range body.Results {
		if len(r.Media) == 0 {
			continue
		}
		m := r.Media[0]
		// HANYA format transparan (stiker tanpa background).
		preview := first(m["tinygif_transparent"].URL, m["nanogif_transparent"].URL)
		full := first(m["webp_transparent"].URL, m["gif_transparent"].URL, m["png_transparent"].URL)
		if preview == "" || full == "" {
			continue
		}
		page.Items = append(page.Items, GifDTO{ID: r.ID, Preview: preview, Mp4: full})
	}
	return page
}

// first mengembalikan argumen non-kosong pertama.
func first(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// itoa kecil tanpa import strconv di banyak tempat.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
