//go:build !(js && wasm)

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"codeberg.org/hum3/lofigui"
)

const batchSize = 40

func modelWithYield(app *lofigui.App) {
	lofigui.Markdown("## With Yield()")
	lofigui.Printf("Generating %d customers...", batchSize)

	for i := 1; i <= batchSize; i++ {
		name := fmt.Sprintf("Customer-%04d", i)
		balance := rand.Float64() * 10000
		lofigui.Printf("%s — balance: £%.2f", name, balance)
		app.Sleep(500 * time.Millisecond)
		lofigui.Yield()
	}

	lofigui.Printf("Done — %d customers created.", batchSize)
	app.EndAction()
}

func modelWithoutYield(app *lofigui.App) {
	lofigui.Markdown("## Without Yield()")
	lofigui.Printf("Generating %d customers...", batchSize)

	for i := 1; i <= batchSize; i++ {
		name := fmt.Sprintf("Customer-%04d", i)
		balance := rand.Float64() * 10000
		lofigui.Printf("%s — balance: £%.2f", name, balance)
		time.Sleep(500 * time.Millisecond)
		runtime.Gosched()
	}

	lofigui.Printf("Done — %d customers created.", batchSize)
	lofigui.Markdown("[Run again](/)")
	app.EndAction()
}

func main() {
	app := lofigui.NewApp()
	app.Version = "Batch Yield v1.0"

	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		Name:         "Batch Yield Demo",
		TemplatePath: "../templates/batch.html",
	})
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	app.SetController(ctrl)
	app.SetRefreshTime(1)
	app.SetDisplayURL("/display")

	http.HandleFunc("/yield", func(w http.ResponseWriter, r *http.Request) {
		app.HandleRoot(w, r, modelWithYield, true)
	})
	http.HandleFunc("/noyield", func(w http.ResponseWriter, r *http.Request) {
		app.HandleRoot(w, r, modelWithoutYield, true)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		app.HandleRoot(w, r, func(a *lofigui.App) {
			modelWithYield(a)
			modelWithoutYield(a)
		}, true)
	})
	http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
		app.HandleDisplay(w, r)
	})
	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1352"
	log.Printf("Starting server on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
