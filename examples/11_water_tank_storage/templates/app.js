const outputDiv = document.getElementById('output');
const statusTag = document.getElementById('status-tag');
const saveTag = document.getElementById('save-tag');
const startStopBtn = document.getElementById('startStopBtn');
const pumpBtn = document.getElementById('pumpBtn');
const valveBtn = document.getElementById('valveBtn');
const tabSchematic = document.getElementById('tab-schematic');
const tabDiagnostics = document.getElementById('tab-diagnostics');

let renderInterval = null;
let saveInterval = null;
let currentTab = 'schematic';

function render() {
    if (currentTab === 'schematic' && typeof goRenderSchematic === 'function') {
        outputDiv.innerHTML = goRenderSchematic();
        // Intercept SVG link clicks
        outputDiv.querySelectorAll('a[href="/pump"]').forEach(function(a) {
            a.addEventListener('click', function(e) { e.preventDefault(); goTogglePump(); saveState(); render(); });
        });
        outputDiv.querySelectorAll('a[href="/valve"]').forEach(function(a) {
            a.addEventListener('click', function(e) { e.preventDefault(); goToggleValve(); saveState(); render(); });
        });
    } else if (currentTab === 'diagnostics' && typeof goRenderDiagnostics === 'function') {
        outputDiv.innerHTML = goRenderDiagnostics();
    }
    updateStatus();
}

function updateStatus() {
    const running = typeof goIsRunning === 'function' && goIsRunning();
    statusTag.textContent = running ? 'Running' : 'Stopped';
    statusTag.className = running ? 'tag is-warning' : 'tag is-success';
    startStopBtn.textContent = running ? 'Stop Simulation' : 'Start Simulation';
    startStopBtn.className = running ? 'button is-danger' : 'button is-success';
}

function switchTab(tab) {
    currentTab = tab;
    tabSchematic.className = tab === 'schematic' ? 'is-active' : '';
    tabDiagnostics.className = tab === 'diagnostics' ? 'is-active' : '';
    render();
}

function toggleStartStop() {
    if (goIsRunning()) {
        goStop();
        saveState();
        if (renderInterval) {
            clearInterval(renderInterval);
            renderInterval = null;
        }
        if (saveInterval) {
            clearInterval(saveInterval);
            saveInterval = null;
        }
    } else {
        goStart();
        saveState();
        renderInterval = setInterval(render, 500);
        saveInterval = setInterval(saveState, 5000);
    }
    render();
}

function ensureRenderInterval() {
    // Start polling when maintenance is running so progress updates show
    if (!renderInterval) {
        renderInterval = setInterval(render, 500);
    }
}

// --- Persistence layer ---

function showSaveIndicator() {
    saveTag.style.display = '';
    saveTag.textContent = 'Saved';
    saveTag.className = 'tag is-success is-light ml-2';
    setTimeout(function() { saveTag.style.display = 'none'; }, 1500);
}

function showSaveError() {
    saveTag.style.display = '';
    saveTag.textContent = 'Save failed';
    saveTag.className = 'tag is-danger is-light ml-2';
    setTimeout(function() { saveTag.style.display = 'none'; }, 3000);
}

function saveState() {
    if (typeof goExportState !== 'function') return;
    var json = goExportState();
    if (!json) return;
    fetch('/api/state', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: json
    }).then(function(resp) {
        if (resp.ok) showSaveIndicator();
        else showSaveError();
    }).catch(function() {
        showSaveError();
    });
}

function loadState() {
    return fetch('/api/state')
        .then(function(resp) { return resp.json(); })
        .then(function(data) {
            if (data && typeof goImportState === 'function') {
                goImportState(JSON.stringify(data));
            }
        })
        .catch(function(err) {
            console.log('No saved state found:', err);
        });
}

function saveMaintenanceLog(type, message) {
    var entry = {
        timestamp: new Date().toISOString(),
        type: type,
        message: message
    };
    fetch('/api/logs', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(entry)
    }).catch(function() {});
}

function saveDiagnostic() {
    if (typeof goExportDiagnostics !== 'function') return;
    var json = goExportDiagnostics();
    if (!json) return;
    fetch('/api/diagnostics', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: json
    }).catch(function() {});
}

// Save state on page unload
window.addEventListener('beforeunload', function() {
    if (typeof goExportState !== 'function') return;
    var json = goExportState();
    if (json) {
        navigator.sendBeacon('/api/state', json);
    }
});

// --- WASM loading ---

window.wasmReady = function() {
    loadState().then(function() {
        startStopBtn.disabled = false;
        pumpBtn.disabled = false;
        valveBtn.disabled = false;
        render();
    });
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
pumpBtn.addEventListener('click', function() { goTogglePump(); saveState(); render(); });
valveBtn.addEventListener('click', function() { goToggleValve(); saveState(); render(); });
tabSchematic.addEventListener('click', function() { switchTab('schematic'); });
tabDiagnostics.addEventListener('click', function() { switchTab('diagnostics'); });

loadWASM();
