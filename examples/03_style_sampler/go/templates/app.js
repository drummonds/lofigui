const outputDiv = document.getElementById('output');
const statusTag = document.getElementById('status-tag');
const buttons = document.querySelectorAll('#style-buttons button');

// Map server-style paths (used by the cards inside home.html) to WASM templates
// so the demo works without a real HTTP server behind it.
const pathToTemplate = {
    '/': 'home.html',
    '/style/scrolling': 'style_scrolling.html',
    '/style/fixed': 'style_fixed.html',
    '/style/three-panel-nav': 'style_three_panel_nav.html',
    '/style/three-panel-controls': 'style_three_panel_controls.html',
    '/style/fullwidth': 'style_fullwidth.html',
};

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

// Intercept clicks on rendered internal links (e.g. the "View" buttons on
// home.html) and route them through goRenderPage instead of navigating.
outputDiv.addEventListener('click', function(e) {
    const link = e.target.closest('a[href]');
    if (!link) return;
    const href = link.getAttribute('href');
    const tpl = pathToTemplate[href];
    if (tpl) {
        e.preventDefault();
        renderStyle(tpl, href);
    }
});

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
