//go:build !(js && wasm)

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/drummonds/lofigui"
	"github.com/flosch/pongo2/v6"
)

const htmxLayout = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ controller_name }}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
  <script src="https://unpkg.com/htmx.org@2.0.4"></script>
</head>
<body>
  <nav class="navbar is-primary" role="navigation">
    <div class="navbar-brand">
      <span class="navbar-item has-text-weight-bold">{{ controller_name }}</span>
    </div>
    <div class="navbar-end">
      <div class="navbar-item">
        <span class="tag {% if connected %}is-success{% else %}is-danger{% endif %}">
          {% if connected %}Connected{% else %}Disconnected{% endif %}
        </span>
      </div>
    </div>
  </nav>
  <section class="section">
    <div class="container">
      <div id="results" hx-get="{{ fragment_url }}" hx-trigger="every 1s" hx-swap="innerHTML">
        {{ results | safe }}
      </div>
    </div>
  </section>
  <footer class="footer">
    <div class="content has-text-centered">
      <p>{{ version }}</p>
    </div>
  </footer>
</body>
</html>`

func isConnected() bool {
	return client.Ping() == nil
}

func renderMain() {
	connected := isConnected()
	errMsg := getLastError()

	if errMsg != "" {
		lofigui.HTML(fmt.Sprintf(`<div class="notification is-danger is-light">
  <button class="delete" onclick="this.parentElement.remove()"></button>
  <strong>Error:</strong> %s
</div>`, errMsg))
		setLastError(nil)
	}

	if !connected {
		lofigui.HTML(`<div class="notification is-warning">
  <p class="has-text-weight-bold">SeaweedFS not reachable</p>
  <p>Start SeaweedFS with: <code>weed server -dir=/data</code></p>
  <p>Master expected at: <code>` + client.MasterURL + `</code></p>
</div>`)
		return
	}

	// Controls
	auto := isAutoActive()
	var autoBtn string
	if auto {
		autoBtn = `<form action="/auto/stop" method="post" style="display:inline"><button class="button is-warning" type="submit">Stop Auto-Create</button></form>`
	} else {
		autoBtn = `<form action="/auto/start" method="post" style="display:inline"><button class="button is-info is-outlined" type="submit">Auto-Create (every 3s)</button></form>`
	}

	lofigui.HTML(fmt.Sprintf(`<div class="buttons">
  <form action="/create" method="post" style="display:inline"><button class="button is-success" type="submit">Create Test File</button></form>
  %s
</div>`, autoBtn))

	// File list
	filesMu.Lock()
	filesCopy := make([]StoredFile, len(files))
	copy(filesCopy, files)
	filesMu.Unlock()

	if len(filesCopy) == 0 {
		lofigui.HTML(`<p class="has-text-grey">No files stored yet. Click "Create Test File" to begin.</p>`)
		return
	}

	lofigui.HTML(fmt.Sprintf(`<p class="mb-2"><strong>%d file(s)</strong> stored in SeaweedFS</p>`, len(filesCopy)))
	lofigui.HTML(`<table class="table is-bordered is-striped is-narrow is-fullwidth">
<thead><tr>
  <th>#</th><th>Name</th><th>Fid</th><th>Size</th><th>Created</th><th>Verified</th><th>Actions</th>
</tr></thead><tbody>`)

	for i, f := range filesCopy {
		verified := `<span class="tag is-light">-</span>`
		if f.Verified {
			verified = `<span class="tag is-success">OK</span>`
		}
		lofigui.HTML(fmt.Sprintf(`<tr>
  <td>%d</td>
  <td>%s</td>
  <td><code>%s</code></td>
  <td>%d B</td>
  <td>%s</td>
  <td>%s</td>
  <td>
    <form action="/verify?i=%d" method="post" style="display:inline"><button class="button is-small is-info is-outlined" type="submit">Verify</button></form>
    <a href="/download?i=%d" download="%s" class="button is-small is-link is-outlined">Download</a>
    <form action="/delete?i=%d" method="post" style="display:inline"><button class="button is-small is-danger is-outlined" type="submit">Delete</button></form>
  </td>
</tr>`, i+1, f.Name, f.Fid, f.Size, f.CreatedAt.Format("15:04:05"), verified, i, i, f.Name, i))
	}

	lofigui.HTML(`</tbody></table>`)
}

func main() {
	client = NewClient("http://localhost:9333")

	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		TemplateString: htmxLayout,
		Name:           "SeaweedFS Demo",
	})
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	version := "SeaweedFS Demo v0.1"

	// GET / — full page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		content := renderAndCapture(renderMain)
		ctrl.RenderTemplate(w, pongo2.Context{
			"controller_name": ctrl.Name,
			"version":         version,
			"results":         content,
			"fragment_url":    "/fragment",
			"connected":       isConnected(),
		})
	})

	// GET /fragment — HTMX fragment
	http.HandleFunc("/fragment", func(w http.ResponseWriter, r *http.Request) {
		content := renderAndCapture(renderMain)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, content)
	})

	// POST /create
	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if err := createTestFile(); err != nil {
			setLastError(err)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /verify?i=N
	http.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		var idx int
		fmt.Sscanf(r.URL.Query().Get("i"), "%d", &idx)
		if err := verifyFile(idx); err != nil {
			setLastError(err)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /delete?i=N
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		var idx int
		fmt.Sscanf(r.URL.Query().Get("i"), "%d", &idx)
		if err := deleteFile(idx); err != nil {
			setLastError(err)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// GET /download?i=N — download file content from SeaweedFS
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		var idx int
		fmt.Sscanf(r.URL.Query().Get("i"), "%d", &idx)

		filesMu.Lock()
		if idx < 0 || idx >= len(files) {
			filesMu.Unlock()
			http.NotFound(w, r)
			return
		}
		f := files[idx]
		filesMu.Unlock()

		data, err := client.Download(f.VolumeURL, f.Fid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, f.Name))
		w.Write(data)
	})

	// POST /auto/start
	http.HandleFunc("/auto/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		startAutoCreate()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// POST /auto/stop
	http.HandleFunc("/auto/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		stopAutoCreate()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1350"
	log.Printf("SeaweedFS Demo on http://localhost%s", addr)
	log.Printf("Expects SeaweedFS master at %s", client.MasterURL)
	http.ListenAndServe(addr, nil)
}
