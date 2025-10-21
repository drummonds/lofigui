/**
 * Main application file for Lofigui WASM Example
 * Manages the Web Worker and UI interactions
 */

// DOM elements
const statusDiv = document.getElementById('status');
const statusText = document.getElementById('status-text');
const outputDiv = document.getElementById('output');
const runBasicBtn = document.getElementById('runBasic');
const runAdvancedBtn = document.getElementById('runAdvanced');
const reloadBtn = document.getElementById('reload');
const pythonSourceDiv = document.getElementById('python-source');

// Web Worker
let worker = null;
let isReady = false;

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
    isReady = enabled;
}

/**
 * Load Python source code from file
 */
async function loadPythonSource() {
    try {
        const response = await fetch('hello_model.py');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const code = await response.text();

        // Display the source code
        pythonSourceDiv.textContent = code;

        return code;
    } catch (error) {
        console.error('Failed to load Python source:', error);
        pythonSourceDiv.textContent = `Error loading source: ${error.message}`;
        throw error;
    }
}

/**
 * Initialize the Web Worker and Pyodide
 */
async function initialize() {
    try {
        // Create Web Worker
        updateStatus('Creating Web Worker...');
        worker = new Worker('worker.js');

        // Set up message handler
        worker.onmessage = function(event) {
            const { type, message, html } = event.data;

            switch (type) {
                case 'worker-ready':
                    console.log('Worker is ready');
                    break;

                case 'status':
                    updateStatus(message, 'info');
                    break;

                case 'ready':
                    updateStatus('Ready! Click a button to run Python code.', 'success');
                    setButtonsEnabled(true);
                    break;

                case 'result':
                    displayOutput(html);
                    break;

                case 'error':
                    updateStatus(message, 'danger');
                    displayError(message);
                    break;

                default:
                    console.log('Unknown message type:', type);
            }
        };

        worker.onerror = function(error) {
            console.error('Worker error:', error);
            updateStatus(`Worker error: ${error.message}`, 'danger');
            displayError(error.message);
        };

        // Initialize Pyodide in worker
        worker.postMessage({ type: 'init' });

        // Load Python source code
        const pythonCode = await loadPythonSource();

        // Send Python code to worker
        worker.postMessage({
            type: 'loadCode',
            code: pythonCode
        });

    } catch (error) {
        console.error('Initialization error:', error);
        updateStatus(`Initialization failed: ${error.message}`, 'danger');
        displayError(error.message);
    }
}

/**
 * Execute a Python function in the worker
 */
function executePython(functionName) {
    if (!isReady) {
        updateStatus('Python runtime not ready yet', 'warning');
        return;
    }

    displayOutput('<div class="loading"><progress class="progress is-primary" max="100">Executing...</progress></div>');

    worker.postMessage({
        type: 'execute',
        functionName: functionName
    });
}

/**
 * Event Listeners
 */
runBasicBtn.addEventListener('click', () => {
    executePython('model');
});

runAdvancedBtn.addEventListener('click', () => {
    executePython('advanced_model');
});

reloadBtn.addEventListener('click', () => {
    location.reload();
});

/**
 * Start the application
 */
initialize();
