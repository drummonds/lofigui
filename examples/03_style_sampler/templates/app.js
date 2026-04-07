const outputDiv = document.getElementById('output');
const statusTag = document.getElementById('status-tag');
const buttons = document.querySelectorAll('#style-buttons button');

function renderStyle(templateName, path) {
    if (typeof goRenderPage !== 'function') return;

    const html = goRenderPage(templateName, path);
    outputDiv.innerHTML = html;

    // Highlight active button
    buttons.forEach(btn => {
        if (btn.dataset.template === templateName) {
            btn.classList.remove('is-outlined');
        } else {
            btn.classList.add('is-outlined');
        }
    });
}

window.wasmReady = function() {
    statusTag.textContent = 'Ready';
    statusTag.className = 'tag is-success';

    // Wire up style buttons
    buttons.forEach(btn => {
        btn.addEventListener('click', function() {
            renderStyle(this.dataset.template, this.dataset.path);
        });
    });

    // Render home page by default
    renderStyle('home.html', '/');
};

async function loadWASM() {
    try {
        const go = new Go();
        const result = await WebAssembly.instantiateStreaming(
            fetch(window.WASM_BINARY || 'main.wasm'), go.importObject
        );
        go.run(result.instance);
    } catch (err) {
        outputDiv.innerHTML = '<div class="notification is-danger">Failed to load WASM: ' + err.message + '</div>';
        statusTag.textContent = 'Error';
        statusTag.className = 'tag is-danger';
    }
}

loadWASM();
