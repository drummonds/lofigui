/**
 * Web Worker for running Python code via Pyodide
 * This keeps the Python execution off the main thread
 */

// Import Pyodide from CDN
importScripts('https://cdn.jsdelivr.net/pyodide/v0.25.0/full/pyodide.js');

let pyodide = null;
let pythonCode = null;

/**
 * Initialize Pyodide runtime
 */
async function initPyodide() {
    try {
        self.postMessage({ type: 'status', message: 'Loading Pyodide runtime...' });
        pyodide = await loadPyodide({
            indexURL: 'https://cdn.jsdelivr.net/pyodide/v0.25.0/full/'
        });
        self.postMessage({ type: 'status', message: 'Pyodide loaded successfully!' });
        return true;
    } catch (error) {
        self.postMessage({
            type: 'error',
            message: `Failed to load Pyodide: ${error.message}`
        });
        return false;
    }
}

/**
 * Load Python source code
 */
async function loadPythonCode(code) {
    try {
        pythonCode = code;
        self.postMessage({ type: 'status', message: 'Loading Python code...' });

        // Execute the Python module code to define functions
        await pyodide.runPythonAsync(code);

        self.postMessage({
            type: 'status',
            message: 'Python code loaded and ready!'
        });
        self.postMessage({ type: 'ready' });
        return true;
    } catch (error) {
        self.postMessage({
            type: 'error',
            message: `Failed to load Python code: ${error.message}`
        });
        return false;
    }
}

/**
 * Execute a Python function
 */
async function executePythonFunction(functionName) {
    try {
        self.postMessage({
            type: 'status',
            message: `Executing ${functionName}()...`
        });

        // Call the Python function and get the result
        const result = await pyodide.runPythonAsync(`${functionName}()`);

        self.postMessage({
            type: 'result',
            html: result
        });

        self.postMessage({
            type: 'status',
            message: 'Execution complete!'
        });
    } catch (error) {
        self.postMessage({
            type: 'error',
            message: `Execution error: ${error.message}`
        });
    }
}

/**
 * Message handler from main thread
 */
self.onmessage = async function(event) {
    const { type, code, functionName } = event.data;

    switch (type) {
        case 'init':
            await initPyodide();
            break;

        case 'loadCode':
            if (!pyodide) {
                self.postMessage({
                    type: 'error',
                    message: 'Pyodide not initialized'
                });
                return;
            }
            await loadPythonCode(code);
            break;

        case 'execute':
            if (!pyodide || !pythonCode) {
                self.postMessage({
                    type: 'error',
                    message: 'Python runtime not ready'
                });
                return;
            }
            await executePythonFunction(functionName);
            break;

        default:
            self.postMessage({
                type: 'error',
                message: `Unknown message type: ${type}`
            });
    }
};

// Signal that worker is ready to receive messages
self.postMessage({ type: 'worker-ready' });
