from types import CodeType
import lofigui as lg
from time import sleep

from fastapi import Request, BackgroundTasks
from fastapi.responses import HTMLResponse
import uvicorn

# Use create_app which automatically includes favicon route
app = lg.create_app(template_dir="../templates")


# This is the guts of code.
def model():
    lg.print("Hello world.")
    for i in range(5):
        sleep(1)
        lg.print(f"Count {i}")
    lg.markdown("<a href='/'>Restart</a>")
    lg.print("Done.")
    controller.end_action()


controller = lg.Controller()


@app.get("/", response_class=HTMLResponse)
async def root(background_tasks: BackgroundTasks):
    """This is thee root and also the action state so it is transititive"""
    lg.reset()  # If you don't have this the runs keep on concatenating.
    background_tasks.add_task(model)
    controller.start_action()
    return f'<head> <meta http-equiv="Refresh" content="0; URL=/display"/></head>'


@app.get("/display", response_class=HTMLResponse)
async def action(request: Request):
    d = controller.state_dict(request)
    return app.templates.TemplateResponse("hello.html", d)


if __name__ == "__main__":
    uvicorn.run("hello:app", host="127.0.0.1", port=1340, reload=True)
