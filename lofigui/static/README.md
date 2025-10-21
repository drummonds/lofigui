# Lofigui Favicon

This directory contains the default lofigui favicon.

## Files

- **favicon.svg** - SVG version of the favicon (32x32 logical pixels)
  - Blue background (#3273dc - Bulma primary color)
  - White "L" letter logo
  - Scalable vector format

## Usage

### Python

The favicon is embedded in the `lofigui.favicon` module and can be used in several ways:

#### 1. Serve with FastAPI (Recommended)

```python
import lofigui as lg
from fastapi import FastAPI

app = FastAPI()

@app.get("/favicon.ico")
async def favicon():
    return lg.get_favicon_response()
```

#### 2. Data URI in HTML Template

```python
# In your view/template
html_tag = lg.get_favicon_html_tag()
# Returns: <link rel="icon" type="image/x-icon" href="data:image/x-icon;base64,...">
```

#### 3. Save to File

```bash
# From command line
python -m lofigui.favicon path/to/favicon.ico

# Or in code
import lofigui as lg
lg.save_favicon_ico("favicon.ico")
```

### Go

The favicon is embedded in the `lofigui` package:

#### 1. Serve with net/http (Recommended)

```go
import "github.com/drummonds/lofigui/go/lofigui"

http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)
```

#### 2. Get as Bytes

```go
faviconBytes, err := lofigui.GetFaviconICO()
```

#### 3. Get Data URI

```go
dataURI := lofigui.GetFaviconDataURI()
htmlTag := lofigui.GetFaviconHTMLTag()
```

## Customization

To use your own favicon:

1. Replace `favicon.svg` with your own SVG (recommended: 32x32 viewBox)
2. Or create a custom `/favicon.ico` route in your application
3. Or use a data URI in your HTML templates

## Design

The logo is a simple white "L" on a blue background:
- **Blue**: #3273dc (Bulma primary blue)
- **White**: #ffffff
- **Shape**: Stylized "L" letter
- **Size**: 32x32 pixels (logical)

The design is intentionally minimal to:
- Keep file size small
- Work at any scale
- Match the "lofi" (low-fidelity) aesthetic
- Be recognizable in browser tabs
