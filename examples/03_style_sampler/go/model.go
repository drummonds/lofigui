package main

import (
	"sync"

	"codeberg.org/hum3/lofigui"
)

var renderMu sync.Mutex

func renderContent(fn func()) string {
	renderMu.Lock()
	defer renderMu.Unlock()
	lofigui.Reset()
	fn()
	return lofigui.Buffer()
}

// sampleOutput generates teletype content shown in every style demo.
func sampleOutput() string {
	return renderContent(func() {
		lofigui.Markdown("## Teletype Output")
		lofigui.Print("This is sample teletype output.")
		lofigui.Print("Each line appears as the model runs.")
		lofigui.Print("Like a CLI printing to continuous paper.")
		lofigui.Markdown("---")
		lofigui.Table(
			[][]string{
				{"Level 1", "Teletype", "Pure output, no interaction"},
				{"Level 2", "Teletype+ web", "Templates, navbars, forms"},
				{"Level 3", "Polling", "Whole page refresh"},
				{"Level 4", "HTMX", "Partial updates"},
			},
			lofigui.WithHeader([]string{"Level", "Name", "Description"}),
		)
		lofigui.Print("Done.")
	})
}
