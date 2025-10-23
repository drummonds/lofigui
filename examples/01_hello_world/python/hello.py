"""
Hello World example using lofigui with extensible Controller.

This example demonstrates:
1. Using the Controller with custom template directory
2. Manual route configuration for full control
3. Background task execution with action state management

"""

import lofigui as lg
from time import sleep

from fastapi import Request, BackgroundTasks
from fastapi.responses import HTMLResponse
import uvicorn


# Create controller with custom template directory
# The template directory can be anywhere, not just the default "templates"
controller = lg.Controller(
    template_dir="../templates",  # Custom location
    template_name="hello.html",  # Template to use
    refresh_time=1,  # Refresh every 1 second while action runs
)

# Use create_app which automatically includes favicon route
# Pass the same template_dir to create_app for consistency
app = lg.create_app(template_dir="../templates", controller=controller)


# This is the model/business logic
def model():
    """Main business logic that runs in the background"""
    lg.print("Hello world.")
    for i in range(5):
        sleep(1)
        lg.print(f"Count {i}")
    lg.markdown('<a href="/">Restart</a>')
    lg.print("Done.")
    app.end_action()  # Signal that the action is complete


@app.get("/", response_class=HTMLResponse)
async def root(background_tasks: BackgroundTasks):
    """Root endpoint - starts the action and redirects to display"""
    # Reset the buffer so runs don't concatenate
    lg.reset()

    # Start the action in the background
    background_tasks.add_task(model)
    app.start_action()

    # Redirect to display page
    return '<head><meta http-equiv="Refresh" content="0; URL=/display"/></head>'


@app.get("/display", response_class=HTMLResponse)
async def display(request: Request):
    """Display endpoint - shows the current state with auto-refresh while running"""
    return app.template_response(request, "hello.html")


if __name__ == "__main__":
    uvicorn.run("hello:app", host="127.0.0.1", port=1340, reload=True)
