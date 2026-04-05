const outputYield = document.getElementById('output-yield');
const outputNoYield = document.getElementById('output-noyield');
const statusTag = document.getElementById('status-tag');
const yieldBtn = document.getElementById('yieldBtn');
const noYieldBtn = document.getElementById('noYieldBtn');
const bothBtn = document.getElementById('bothBtn');
const resetBtn = document.getElementById('resetBtn');
const backToTop = document.getElementById('back-to-top');

let renderInterval = null;

function render() {
    if (typeof goRenderYield === 'function') {
        outputYield.innerHTML = goRenderYield();
    }
    if (typeof goRenderNoYield === 'function') {
        outputNoYield.innerHTML = goRenderNoYield();
    }
    updateStatus();
}

function updateStatus() {
    const running = typeof goIsRunning === 'function' && goIsRunning();
    statusTag.textContent = running ? 'Running' : 'Ready';
    statusTag.className = running ? 'tag is-warning' : 'tag is-success';
    yieldBtn.disabled = running;
    noYieldBtn.disabled = running;
    bothBtn.disabled = running;
    // Reset stays enabled — it can cancel a running "with yield" batch.
    if (!running && renderInterval) {
        clearInterval(renderInterval);
        renderInterval = null;
        render(); // final render to capture last output
    }
}

function startRendering() {
    if (!renderInterval) {
        renderInterval = setInterval(render, 100);
    }
    render();
}

yieldBtn.addEventListener('click', function() {
    goStartWithYield();
    startRendering();
});

noYieldBtn.addEventListener('click', function() {
    goStartWithoutYield();
    startRendering();
});

bothBtn.addEventListener('click', function() {
    goStartBoth();
    startRendering();
});

resetBtn.addEventListener('click', function() {
    goReset();
    outputYield.innerHTML = readyMsg;
    outputNoYield.innerHTML = readyMsg;
    updateStatus();
});

// Back to top button
window.addEventListener('scroll', function() {
    backToTop.style.display = window.scrollY > 300 ? 'inline-flex' : 'none';
});

const readyMsg = '<p class="has-text-grey">Click a button above to start.</p>';

window.wasmReady = function() {
    yieldBtn.disabled = false;
    noYieldBtn.disabled = false;
    bothBtn.disabled = false;
    resetBtn.disabled = false;
    outputYield.innerHTML = readyMsg;
    outputNoYield.innerHTML = readyMsg;
};

async function loadWASM() {
    try {
        const go = new Go();
        const result = await WebAssembly.instantiateStreaming(
            fetch(window.WASM_BINARY || 'main.wasm'), go.importObject
        );
        go.run(result.instance);
    } catch (err) {
        const msg = '<div class="notification is-danger">Failed to load WASM: ' + err.message + '</div>';
        outputYield.innerHTML = msg;
        outputNoYield.innerHTML = msg;
    }
}

loadWASM();
