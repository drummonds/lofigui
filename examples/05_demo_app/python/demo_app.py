"""
Lofigui Demo Application

This demo showcases the key features of lofigui including:
- Template inheritance with base.html
- Controller pattern for long-running processes
- Data table rendering
- Chart/visualization display
- Process management with auto-refresh
"""

import asyncio
from fastapi import FastAPI, Request, Form
from fastapi.responses import HTMLResponse, RedirectResponse
import uvicorn

import lofigui as lf
from lofigui import App, Controller

# Initialize app with controller
app_instance = App(template_dir="templates")
app_instance.controller = Controller()
fastapi_app = FastAPI(title="Lofigui Demo")


@fastapi_app.get("/", response_class=HTMLResponse)
async def home(request: Request):
    """Home page with overview of features."""
    return app_instance.template_response(
        request,
        "home.html",
        {
            "status": "Ready",
            "progress": "0%",
            "polling": "Stopped",
        },
    )


@fastapi_app.get("/data", response_class=HTMLResponse)
async def data_tables(request: Request):
    """Data tables demonstration page."""
    # Generate sample table
    lf.reset()

    # Employee table
    lf.markdown("### Employee Data")
    employee_data = [
        ["Alice Johnson", "Senior Developer", "Python, Go", "5 years"],
        ["Bob Smith", "UX Designer", "Figma, Sketch", "3 years"],
        ["Carol White", "DevOps Engineer", "Docker, K8s", "4 years"],
        ["David Brown", "Product Manager", "Agile, Scrum", "6 years"],
    ]
    lf.table(employee_data, header=["Name", "Role", "Skills", "Experience"])
    table_html = lf.buffer()

    # Technology comparison table
    lf.reset()
    lf.markdown("### Framework Comparison")
    comparison_data = [
        ["Lofigui", "Python/Go", "Lightweight", "Templates + Controller"],
        ["Flask", "Python", "Micro-framework", "Full web framework"],
        ["FastAPI", "Python", "High performance", "API-first, async"],
        ["Gin", "Go", "Fast", "Minimal web framework"],
    ]
    lf.table(comparison_data, header=["Framework", "Language", "Type", "Features"])
    comparison_table = lf.buffer()

    return app_instance.template_response(
        request,
        "data.html",
        {
            "table_html": table_html,
            "comparison_table": comparison_table,
            "status": "Ready",
            "progress": "0%",
            "polling": "Stopped",
        },
    )


@fastapi_app.get("/charts", response_class=HTMLResponse)
async def charts(request: Request):
    """Charts and visualizations page."""
    lf.reset()

    lf.markdown("### Sample Data Visualization")
    lf.print("Chart rendering can be integrated with libraries like Chart.js, Plotly, or D3.js")

    # Add some sample markdown content
    lf.markdown(
        """
**Example Integration:**

```javascript
// Chart.js example
const ctx = document.getElementById('myChart');
new Chart(ctx, {
    type: 'bar',
    data: {
        labels: ['Red', 'Blue', 'Yellow', 'Green', 'Purple', 'Orange'],
        datasets: [{
            label: '# of Votes',
            data: [12, 19, 3, 5, 2, 3]
        }]
    }
});
```
    """
    )

    chart_content = lf.buffer()

    return app_instance.template_response(
        request,
        "charts.html",
        {
            "chart_content": chart_content,
            "status": "Ready",
            "progress": "0%",
            "polling": "Stopped",
        },
    )


@fastapi_app.get("/process", response_class=HTMLResponse)
async def process_page(request: Request):
    """Process management page."""
    lf.reset()

    if app_instance.is_action_running():
        lf.markdown("### Process Running...")
        lf.print("A process is currently running. Check the status panel for progress.")
        lf.print("")
        lf.print("You can cancel the process using the Cancel button in the top navigation.")
        cancel_class = ""
    else:
        lf.markdown("### Ready to Start")
        lf.print("No process is currently running.")
        lf.print("Fill in the form below and click 'Start Process' to begin.")
        cancel_class = "is-static"

    process_output = lf.buffer()

    # Get status info
    if app_instance.is_action_running():
        status = "Running"
        polling = "Active (2s refresh)"
    else:
        status = "Idle"
        polling = "Stopped"

    return app_instance.template_response(
        request,
        "process.html",
        {
            "process_output": process_output,
            "status": status,
            "progress": "0%",
            "polling": polling,
            "cancel_class": cancel_class,
        },
    )


@fastapi_app.post("/start_demo_process")
async def start_demo_process(duration: int = Form(10)):
    """Start a demo process with specified duration."""
    # Start the action with 2-second refresh
    app_instance.start_action(refresh_time=2)

    # Simulate a long-running process
    lf.reset()
    lf.markdown("## Demo Process Started")
    lf.print(f"Running for {duration} seconds...")
    lf.print("")

    # Clamp duration to reasonable range
    duration = max(1, min(duration, 60))

    for i in range(duration):
        progress = int((i + 1) / duration * 100)
        lf.print(f"Step {i + 1}/{duration} - Progress: {progress}%")
        await asyncio.sleep(1)

    lf.print("")
    lf.markdown("### Process Complete!")
    lf.print(f"Successfully completed {duration} steps.")

    # End the action
    app_instance.end_action()

    return RedirectResponse(url="/display", status_code=303)


@fastapi_app.post("/stop")
async def stop_process():
    """Stop the currently running process."""
    if app_instance.is_action_running():
        app_instance.end_action()

    return RedirectResponse(url="/", status_code=303)


@fastapi_app.get("/display", response_class=HTMLResponse)
async def display_results(request: Request):
    """Display results from completed process."""
    lf.reset()

    lf.markdown("### Process Results")
    lf.print("The process has completed successfully.")
    lf.print("")
    lf.markdown("You can view the output above or start a new process.")

    results_content = lf.buffer()

    return app_instance.template_response(
        request,
        "display.html",
        {
            "results_content": results_content,
            "status": "Complete",
            "progress": "100%",
            "polling": "Stopped",
        },
    )


@fastapi_app.get("/about", response_class=HTMLResponse)
async def about(request: Request):
    """About page with information about lofigui."""
    return app_instance.template_response(
        request,
        "about.html",
        {
            "status": "Ready",
            "progress": "0%",
            "polling": "Stopped",
        },
    )


@fastapi_app.get("/favicon.ico")
async def favicon():
    """Serve favicon."""
    return lf.get_favicon_response()


if __name__ == "__main__":
    print("=" * 60)
    print("Lofigui Demo Application")
    print("=" * 60)
    print("")
    print("Starting server at http://localhost:8050")
    print("")
    print("Features demonstrated:")
    print("  - Template inheritance with base.html")
    print("  - Data table rendering")
    print("  - Chart/visualization display")
    print("  - Process management with auto-refresh")
    print("  - Controller pattern for long-running tasks")
    print("")
    print("Press Ctrl+C to stop the server")
    print("=" * 60)

    uvicorn.run(fastapi_app, host="0.0.0.0", port=8050)
