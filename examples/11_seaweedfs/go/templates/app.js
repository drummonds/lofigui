const statusDiv = document.getElementById('status');
const resultsDiv = document.getElementById('results');
const connStatus = document.getElementById('conn-status');

let ready = false;

window.wasmReady = function() {
    ready = true;
    statusDiv.style.display = 'none';
    updateDisplay();
    setInterval(updateDisplay, 1000);
};

function updateDisplay() {
    if (!ready) return;
    try {
        resultsDiv.innerHTML = goRender();
        const connected = goIsConnected();
        connStatus.textContent = connected ? 'Connected' : 'Disconnected';
        connStatus.className = 'tag ' + (connected ? 'is-success' : 'is-danger');
    } catch (e) {
        console.error('Render error:', e);
    }
}

async function loadWASM() {
    try {
        const go = new Go();
        const result = await WebAssembly.instantiateStreaming(
            fetch('main.wasm'),
            go.importObject
        );
        go.run(result.instance);
    } catch (error) {
        statusDiv.innerHTML = '<div class="notification is-danger">' +
            '<strong>Failed to load WASM:</strong> ' + error.message +
            '</div>';
    }
}

loadWASM();
