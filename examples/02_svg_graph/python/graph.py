from lofigui import buffer
import lofigui as lg
from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
import pygal
import uvicorn

app = FastAPI()


# This is the guts of code.  This is a static app so doesn't really need the overhead
# but the structure is the same once it becomes dynamic.
def model():
    lg.print("Hello to graph.")
    # add graph
    bar_chart = pygal.Bar()  # Then create a bar graph object
    bar_chart.add("Fibonacci", [0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55])  # Add some values
    lg.html(bar_chart.render().decode("utf-8"))


class Controller:
    """Async controller class for datacolour, but also combined with model"""

    def __init__(self):
        pass

    def state_dict(self, d):
        d["results"] = buffer()
        return d


controller = Controller()

templates = Jinja2Templates(directory="../templates")


@app.get("/favicon.ico")
async def favicon():
    """Serve the lofigui favicon"""
    return lg.get_favicon_response()


@app.get("/", response_class=HTMLResponse)
async def root(request: Request):
    lg.reset()  # If you don't have this the runs keep concatenating.
    model()
    return templates.TemplateResponse(
        "hello.html", controller.state_dict({"request": request})
    )


if __name__ == "__main__":
    uvicorn.run("graph:app", host="127.0.0.1", port=1340, reload=True)
