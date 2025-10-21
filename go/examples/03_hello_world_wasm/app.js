/**
 * Main application file for Lofigui Go WASM Example
 * Loads and manages the Go WASM module
 */

// DOM elements
const statusDiv = document.getElementById('status');
const statusText = document.getElementById('status-text');
const outputDiv = document.getElementById('output');
const runBasicBtn = document.getElementById('runBasic');
const runAdvancedBtn = document.getElementById('runAdvanced');
const reloadBtn = document.getElementById('reload');
const goSourceDiv = document.getElementById('go-source');

let wasmReady = false;

/**
 * Update status display
 */
function updateStatus(message, type = 'info') {
    statusText.textContent = message;
    statusDiv.className = `notification is-${type}`;
}

/**
 * Display output HTML
 */
function displayOutput(html) {
    outputDiv.innerHTML = html;
}

/**
 * Display error message
 */
function displayError(message) {
    outputDiv.innerHTML = `
        <div class="notification is-danger">
            <strong>Error:</strong> ${message}
        </div>
    `;
}

/**
 * Enable/disable buttons
 */
function setButtonsEnabled(enabled) {
    runBasicBtn.disabled = !enabled;
    runAdvancedBtn.disabled = !enabled;
    wasmReady = enabled;
}

/**
 * Called by Go WASM when it's ready
 */
window.wasmReady = function() {
    console.log('Go WASM is ready!');
    updateStatus('Ready! Click a button to run Go code.', 'success');
    setButtonsEnabled(true);

    // Load source code
    if (typeof goGetSourceCode === 'function') {
        goSourceDiv.textContent = goGetSourceCode();
    }
};

/**
 * Initialize and load the WASM module
 */
async function loadWASM() {
    try {
        updateStatus('Loading Go WASM module...', 'info');

        // Create the Go instance
        const go = new Go();

        // Fetch and instantiate the WASM module
        const result = await WebAssembly.instantiateStreaming(
            fetch('main.wasm'),
            go.importObject
        );

        // Run the Go program
        go.run(result.instance);

        // Note: wasmReady() will be called from Go code when ready

    } catch (error) {
        console.error('Failed to load WASM:', error);
        updateStatus(`Failed to load WASM: ${error.message}`, 'danger');
        displayError(error.message);
    }
}

/**
 * Run the basic model
 */
function runBasic() {
    if (!wasmReady) {
        updateStatus('WASM not ready yet', 'warning');
        return;
    }

    try {
        updateStatus('Executing Go code...', 'info');
        const result = goRunModel();
        displayOutput(result);
        updateStatus('Execution complete!', 'success');
    } catch (error) {
        console.error('Execution error:', error);
        updateStatus(`Execution error: ${error.message}`, 'danger');
        displayError(error.message);
    }
}

/**
 * Run the advanced model
 */
function runAdvanced() {
    if (!wasmReady) {
        updateStatus('WASM not ready yet', 'warning');
        return;
    }

    try {
        updateStatus('Executing Go code...', 'info');
        const result = goRunAdvancedModel();
        displayOutput(result);
        updateStatus('Execution complete!', 'success');
    } catch (error) {
        console.error('Execution error:', error);
        updateStatus(`Execution error: ${error.message}`, 'danger');
        displayError(error.message);
    }
}

/**
 * Event Listeners
 */
runBasicBtn.addEventListener('click', runBasic);
runAdvancedBtn.addEventListener('click', runAdvanced);
reloadBtn.addEventListener('click', () => location.reload());

/**
 * Start the application
 */
loadWASM();
