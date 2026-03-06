//go:build js && wasm

package main

import (
	"fmt"
	"net/http"
	"sync"
	"syscall/js"
	"time"

	"github.com/drummonds/lofigui"
)

var (
	connMu    sync.Mutex
	connState bool
)

func wasmIsConnected() bool {
	connMu.Lock()
	defer connMu.Unlock()
	return connState
}

func connectionChecker() {
	for {
		ok := client.Ping() == nil
		connMu.Lock()
		connState = ok
		connMu.Unlock()
		time.Sleep(2 * time.Second)
	}
}

func wasmRender() {
	connected := wasmIsConnected()
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
  <p>Waiting for connection via proxy...</p>
  <p>Master expected at: <code>` + client.MasterURL + `</code></p>
</div>`)
		return
	}

	// Controls
	auto := isAutoActive()
	var autoBtn string
	if auto {
		autoBtn = `<button class="button is-warning" onclick="goStopAuto()">Stop Auto-Create</button>`
	} else {
		autoBtn = `<button class="button is-info is-outlined" onclick="goStartAuto()">Auto-Create (every 3s)</button>`
	}

	lofigui.HTML(fmt.Sprintf(`<div class="buttons">
  <button class="button is-success" onclick="goCreateFile()">Create Test File</button>
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
		downloadURL := client.VolumeFileURL(f.VolumeURL, f.Fid)
		lofigui.HTML(fmt.Sprintf(`<tr>
  <td>%d</td>
  <td>%s</td>
  <td><code>%s</code></td>
  <td>%d B</td>
  <td>%s</td>
  <td>%s</td>
  <td>
    <button class="button is-small is-info is-outlined" onclick="goVerifyFile(%d)">Verify</button>
    <a href="%s" download="%s" class="button is-small is-link is-outlined">Download</a>
    <button class="button is-small is-danger is-outlined" onclick="goDeleteFile(%d)">Delete</button>
  </td>
</tr>`, i+1, f.Name, f.Fid, f.Size, f.CreatedAt.Format("15:04:05"), verified, i, downloadURL, f.Name, i))
	}

	lofigui.HTML(`</tbody></table>`)
}

func goRenderFn(this js.Value, args []js.Value) any {
	content := renderAndCapture(wasmRender)
	return js.ValueOf(content)
}

func goCreateFileFn(this js.Value, args []js.Value) any {
	go func() {
		if err := createTestFile(); err != nil {
			setLastError(err)
		}
	}()
	return nil
}

func goVerifyFileFn(this js.Value, args []js.Value) any {
	idx := args[0].Int()
	go func() {
		if err := verifyFile(idx); err != nil {
			setLastError(err)
		}
	}()
	return nil
}

func goDeleteFileFn(this js.Value, args []js.Value) any {
	idx := args[0].Int()
	go func() {
		if err := deleteFile(idx); err != nil {
			setLastError(err)
		}
	}()
	return nil
}

func goStartAutoFn(this js.Value, args []js.Value) any {
	startAutoCreate()
	return nil
}

func goStopAutoFn(this js.Value, args []js.Value) any {
	stopAutoCreate()
	return nil
}

func goIsConnectedFn(this js.Value, args []js.Value) any {
	return js.ValueOf(wasmIsConnected())
}

func main() {
	client = &Client{
		MasterURL:  "/api/master",
		ProxyBase:  "/api/vol",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	go connectionChecker()

	js.Global().Set("goRender", js.FuncOf(goRenderFn))
	js.Global().Set("goCreateFile", js.FuncOf(goCreateFileFn))
	js.Global().Set("goVerifyFile", js.FuncOf(goVerifyFileFn))
	js.Global().Set("goDeleteFile", js.FuncOf(goDeleteFileFn))
	js.Global().Set("goStartAuto", js.FuncOf(goStartAutoFn))
	js.Global().Set("goStopAuto", js.FuncOf(goStopAutoFn))
	js.Global().Set("goIsConnected", js.FuncOf(goIsConnectedFn))

	js.Global().Call("wasmReady")

	<-make(chan struct{})
}
