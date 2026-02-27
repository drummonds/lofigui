const outputDiv = document.getElementById('output');
const statusTag = document.getElementById('status-tag');
const startStopBtn = document.getElementById('startStopBtn');
const pumpBtn = document.getElementById('pumpBtn');
const valveBtn = document.getElementById('valveBtn');

let renderInterval = null;

function render() {
    if (typeof goRender === 'function') {
        outputDiv.innerHTML = goRender();
        updateStatus();
    }
}

function updateStatus() {
    const running = typeof goIsRunning === 'function' && goIsRunning();
    statusTag.textContent = running ? 'Running' : 'Stopped';
    statusTag.className = running ? 'tag is-warning' : 'tag is-success';
    startStopBtn.textContent = running ? 'Stop Simulation' : 'Start Simulation';
    startStopBtn.className = running ? 'button is-danger' : 'button is-success';
}

function toggleStartStop() {
    if (goIsRunning()) {
        goStop();
        if (renderInterval) {
            clearInterval(renderInterval);
            renderInterval = null;
        }
    } else {
        goStart();
        renderInterval = setInterval(render, 500);
    }
    render();
}

window.wasmReady = function() {
    startStopBtn.disabled = false;
    pumpBtn.disabled = false;
    valveBtn.disabled = false;
    render();
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

startStopBtn.addEventListener('click', toggleStartStop);
pumpBtn.addEventListener('click', function() { goTogglePump(); render(); });
valveBtn.addEventListener('click', function() { goToggleValve(); render(); });

loadWASM();
