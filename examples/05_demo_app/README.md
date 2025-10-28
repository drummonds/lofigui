# Lofigui Demo Application

A comprehensive demonstration of lofigui features including template inheritance, controller pattern, data tables, and process management.

## Features Demonstrated

This demo application showcases:

- **Template Inheritance**: All pages extend a common `base.html` template with Jinja2 template blocks
- **Navigation**: Responsive navigation bar with Bulma CSS and Font Awesome icons
- **Data Tables**: Display tabular data using lofigui's table rendering
- **Charts & Visualizations**: Integration points for charting libraries
- **Process Management**: Long-running processes with automatic page refresh
- **Controller Pattern**: Managing application state and process lifecycle
- **Status Panel**: Real-time status updates during process execution

## Quick Start

### Prerequisites

- Python 3.9 or higher
- lofigui installed (this is an example within the lofigui package)

### Installation

If running from the lofigui repository:

```bash
# Navigate to the demo app directory
cd examples/05_demo_app

# Install dependencies (if not already installed)
uv pip install -e ../..

# Or using regular pip
pip install -e ../..
```

### Running the Application

```bash
# From the examples/05_demo_app directory
python demo_app.py
```

Or using the Taskfile from the root directory:

```bash
# From the lofigui root directory
task demo
```

The application will start on http://localhost:8050

### Using the Application

1. **Home Page** (`/`) - Overview of features and navigation
2. **Data Tables** (`/data`) - View sample data tables and comparisons
3. **Charts** (`/charts`) - See chart integration examples and metrics
4. **Run Process** (`/process`) - Start a demo process with auto-refresh
5. **View Results** (`/display`) - See completed process output
6. **About** (`/about`) - Learn more about lofigui

## Project Structure

```
05_demo_app/
├── demo_app.py              # Main FastAPI application
├── pyproject.toml           # Project configuration
├── README.md                # This file
└── templates/
    ├── base.html            # Base template with navigation and layout
    ├── home.html            # Home page (extends base.html)
    ├── data.html            # Data tables page
    ├── charts.html          # Charts and visualizations page
    ├── process.html         # Process management page
    ├── display.html         # Results display page
    └── about.html           # About page
```

## Template Inheritance

All page templates extend `base.html`:

```html
{% extends "base.html" %}

{% block title %}
<title>My Page - Lofigui Demo App</title>
{% endblock %}

{% block mainpanel %}
<h1 class="title">My Page Content</h1>
<!-- Your content here -->
{% endblock %}
```

### Available Template Blocks

- `title` - Page title in `<head>`
- `content` - Full page content area (includes default two-column layout)
- `mainpanel` - Main content area (left column)
- `statuspanel` - Status panel (right column)
- `extra_css` - Additional CSS styles
- `extra_js` - Additional JavaScript

## Controller Pattern

The demo shows how to use lofigui's Controller for long-running processes:

```python
from lofigui import Controller
import lofigui as lf

# Create controller
controller = Controller(template_path="templates/process.html")
app.controller = controller

# Start process with auto-refresh
controller.start_action(refresh_time=2)

# Do work and update output
lf.reset()
lf.print("Starting process...")
# ... do work ...
lf.print("Process complete!")

# End process (stops auto-refresh)
controller.end_action()
```

## Key Code Patterns

### 1. Template Response

```python
from lofigui import App

app = App(template_dir="templates")

@fastapi_app.get("/")
async def home(request: Request):
    return app.template_response(request, "home.html", {
        "status": "Ready",
        "progress": "0%",
    })
```

### 2. Data Tables

```python
import lofigui as lf

lf.reset()
data = [
    ["Alice", "Developer", "5 years"],
    ["Bob", "Designer", "3 years"],
]
lf.table(data, header=["Name", "Role", "Experience"])
table_html = lf.buffer()

# Pass to template
return app.template_response(request, "page.html", {
    "table_html": table_html
})
```

### 3. Process Management

```python
# Start process
controller.start_action(refresh_time=2)  # Refresh every 2 seconds

# During process, page auto-refreshes
lf.print("Processing...")

# End process
controller.end_action()  # Stops auto-refresh
```

## Customization

### Styling

The demo uses Bulma CSS framework. You can customize by:

1. Adding custom CSS in the `extra_css` block of your templates
2. Modifying `base.html` styles
3. Using Bulma's extensive class system

### Navigation

Edit `base.html` to add/remove navigation items:

```html
<a class="navbar-item" href="/your-page">
    <span class="icon"><i class="fas fa-star"></i></span>
    <span>Your Page</span>
</a>
```

### Adding Pages

1. Create a new template extending `base.html`
2. Add a route in `demo_app.py`
3. Add navigation link in `base.html` (optional)

## Integration Examples

### Chart.js Integration

Add to your template's `extra_js` block:

```html
{% block extra_js %}
<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<script>
const ctx = document.getElementById('myChart');
new Chart(ctx, {
    type: 'bar',
    data: { /* your data */ }
});
</script>
{% endblock %}
```

### Custom Styling

Add to your template's `extra_css` block:

```html
{% block extra_css %}
.custom-card {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 2rem;
    border-radius: 8px;
}
{% endblock %}
```

## Learn More

- [Lofigui Documentation](https://github.com/drummonds/lofigui)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Bulma CSS Documentation](https://bulma.io/documentation/)
- [Jinja2 Template Documentation](https://jinja.palletsprojects.com/)

## License

This demo application is part of the lofigui project and shares the same license.
