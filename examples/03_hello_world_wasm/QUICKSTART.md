# Quick Start Guide - WASM Example

## 5-Minute Setup

### 1. Clone or Download

```bash
git clone https://github.com/drummonds/lofigui.git
cd lofigui/examples/03_hello_world_wasm
```

### 2. Start Local Server

```bash
python3 serve.py
```

Or use Python's built-in server:

```bash
python3 -m http.server 8000
```

### 3. Open Browser

Navigate to: http://localhost:8000

### 4. Try It Out

1. Wait for "Ready!" message (5-10 seconds first time)
2. Click "Run Basic Example"
3. See Python code execute in your browser!

## Deploy to GitHub Pages in 3 Steps

### Step 1: Copy Files

```bash
# From your repo root
mkdir -p docs
cp examples/03_hello_world_wasm/* docs/
git add docs
git commit -m "Add WASM example"
git push
```

### Step 2: Enable GitHub Pages

1. Go to repository Settings → Pages
2. Source: Deploy from branch `main`
3. Folder: `/docs`
4. Click Save

### Step 3: Visit Your Page

After 1-2 minutes, visit:
```
https://YOUR-USERNAME.github.io/YOUR-REPO-NAME/
```

## Customize

Edit [`hello_model.py`](hello_model.py):

```python
def model():
    output = []
    output.append("<h1>My Custom App</h1>")
    output.append("<p>Running Python in the browser!</p>")

    # Add your logic here
    data = [1, 2, 3, 4, 5]
    total = sum(data)
    output.append(f"<p>Sum: {total}</p>")

    return "\n".join(output)
```

Refresh the page and click "Run Basic Example" to see your changes!

## That's It!

You now have Python running as WASM in the browser, deployable to GitHub Pages for free.

See [readme.md](readme.md) for full documentation.
