"""
Hello World example using lofigui with extensible Controller.

This example demonstrates:
1. Using the Controller with custom template directory
2. Manual route configuration for full control
3. Background task execution with action state management

"""

import lofigui as lg
from time import sleep

from fastapi import FastAPI, Request, BackgroundTasks
from fastapi.responses import HTMLResponse
import uvicorn


# Create controller with custom template directory
controller = lg.Controller()

# Create lofigui state manager and FastAPI app separately
lg_app = lg.App(template_dir="../templates", controller=controller)
app = FastAPI()


@app.get("/favicon.ico")
async def favicon():
    return lg.get_favicon_response()


# This is the model/business logic
def model():
    """Main business logic that runs in the background"""
    lg.print("Hello world.")
    for i in range(5):
        sleep(1)
        lg.print(f"Count {i}")
    lg.markdown('<a href="/">Restart</a>')
    lg.print("Done.")
    lg_app.end_action()  # Signal that the action is complete


@app.get("/", response_class=HTMLResponse)
async def root(background_tasks: BackgroundTasks):
    """Root endpoint - starts the action and redirects to display"""
    lg.reset()
    background_tasks.add_task(model)
    lg_app.start_action()
    return '<head><meta http-equiv="Refresh" content="0; URL=/display"/></head>'


@app.get("/display", response_class=HTMLResponse)
async def display(request: Request):
    """Display endpoint - shows the current state with auto-refresh while running"""
    return lg_app.template_response(request, "hello.html")


if __name__ == "__main__":
    uvicorn.run("hello:app", host="127.0.0.1", port=1342, reload=True)
