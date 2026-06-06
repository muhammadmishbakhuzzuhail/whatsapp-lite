package app

// app_settings.go — setelan persisten ringan (app_meta). Saat ini: retensi pesan.

import (
	"strconv"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func atoiDef(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}

// GetRetention mengembalikan jumlah hari retensi pesan (0 = selamanya).
func (a *App) GetRetention() int { return a.retentionDays }

// SetRetention menyetel retensi (hari; 0 = selamanya), simpan, lalu prune+VACUUM.
func (a *App) SetRetention(days int) {
	if days < 0 {
		days = 0
	}
	a.retentionDays = days
	if a.store == nil {
		return
	}
	_ = a.store.SetMeta(a.ctx, "retention_days", strconv.Itoa(days))
	a.bg(func() {
		if cut := a.retentionCutoff(); cut > 0 {
			if n, _ := a.store.PruneMessages(a.ctx, cut); n > 0 {
				_ = a.store.Vacuum(a.ctx)
			}
		}
		runtime.EventsEmit(a.ctx, "wa:sync", "")
	})
}
