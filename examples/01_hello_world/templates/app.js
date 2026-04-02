const outputDiv = document.getElementById('output');
const statusTag = document.getElementById('status-tag');
const startBtn = document.getElementById('startBtn');
const cancelBtn = document.getElementById('cancel-btn');

let renderInterval = null;

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
    cancelBtn.style.display = running ? 'inline-flex' : 'none';
    if (!running && renderInterval) {
        clearInterval(renderInterval);
        renderInterval = null;
    }
}

function start() {
    goStart();
    renderInterval = setInterval(render, 500);
    render();
}

window.wasmReady = function() {
    startBtn.disabled = false;
    outputDiv.innerHTML = '<p class="has-text-grey">Click Start to run.</p>';
};

async function loadWASM() {
    try {
        const go = new Go();
        const result = await WebAssembly.instantiateStreaming(
            fetch('main.wasm'), go.importObject
        );
        go.run(result.instance);
    } catch (err) {
        outputDiv.innerHTML = '<div class="notification is-danger">Failed to load WASM: ' + err.message + '</div>';
    }
}

startBtn.addEventListener('click', start);
cancelBtn.addEventListener('click', function() {
    if (typeof goCancel === 'function') {
        goCancel();
        render();
    }
});

loadWASM();
