const outputDiv = document.getElementById('output');
const statusTag = document.getElementById('status-tag');
const startBtn = document.getElementById('startBtn');
const startGzBtn = document.getElementById('startGzBtn');
const cancelBtn = document.getElementById('cancel-btn');
const sizeInfo = document.getElementById('size-info');

let renderInterval = null;
let uncompressedSize = 0;

function render() {
    if (typeof goRender === 'function') {
        outputDiv.innerHTML = goRender();
        updateStatus();
    }
}

function updateStatus() {
    const running = typeof goIsRunning === 'function' && goIsRunning();
    statusTag.textContent = running ? 'Running' : 'Ready';
    statusTag.className = running ? 'tag is-warning' : 'tag is-success';
    startBtn.disabled = running;
    startGzBtn.disabled = running;
    cancelBtn.style.display = running ? 'inline-flex' : 'none';
    if (!running && renderInterval) {
        clearInterval(renderInterval);
        renderInterval = null;
    }
}

function formatBytes(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / 1048576).toFixed(1) + ' MB';
}

function startModel(label) {
    goStart();
    renderInterval = setInterval(render, 500);
    render();
    sizeInfo.style.display = 'block';
    sizeInfo.textContent = label;
}

window.wasmReady = function() {
    startBtn.disabled = false;
    startGzBtn.disabled = false;
    outputDiv.innerHTML = '<p class="has-text-grey">Click Start to run, or Start (gzipped) to reload from compressed binary.</p>';
};

async function loadWASM() {
    try {
        const go = new Go();
        const resp = await fetch(window.WASM_BINARY || 'main.wasm');
        const blob = await resp.clone().blob();
        uncompressedSize = blob.size;
        const result = await WebAssembly.instantiateStreaming(resp, go.importObject);
        go.run(result.instance);
    } catch (err) {
        outputDiv.innerHTML = '<div class="notification is-danger">Failed to load WASM: ' + err.message + '</div>';
    }
}

async function loadWASMGzipped() {
    startBtn.disabled = true;
    startGzBtn.disabled = true;
    if (renderInterval) { clearInterval(renderInterval); renderInterval = null; }
    outputDiv.innerHTML = '<progress class="progress is-primary" max="100">Loading gzipped WASM...</progress>';

    try {
        const go = new Go();
        const resp = await fetch('main.wasm.gz');
        const blob = await resp.clone().blob();
        const gzSize = blob.size;
        const ds = new DecompressionStream('gzip');
        const decompressed = new Response(resp.body.pipeThrough(ds), {
            headers: {'Content-Type': 'application/wasm'}
        });
        const result = await WebAssembly.instantiateStreaming(decompressed, go.importObject);

        window.wasmReady = function() {
            startModel('Gzipped: ' + formatBytes(gzSize) +
                ' (uncompressed: ' + formatBytes(uncompressedSize) + ')');
        };

        go.run(result.instance);
    } catch (err) {
        outputDiv.innerHTML = '<div class="notification is-danger">Failed to load gzipped WASM: ' + err.message + '</div>';
        startBtn.disabled = false;
        startGzBtn.disabled = false;
    }
}

startBtn.addEventListener('click', function() {
    startModel('Uncompressed: ' + formatBytes(uncompressedSize));
});
startGzBtn.addEventListener('click', loadWASMGzipped);
cancelBtn.addEventListener('click', function() {
    if (typeof goCancel === 'function') {
        goCancel();
        render();
    }
});

loadWASM();
