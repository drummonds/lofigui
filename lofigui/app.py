"""
App factory for lofigui.

Provides a pre-configured FastAPI application with favicon support and
optional controller integration for common patterns.
"""

from typing import Optional, Any
from fastapi import FastAPI, Request, BackgroundTasks
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
from .favicon import get_favicon_response
from .controller import Controller
from .context import buffer


class App(FastAPI):
    """At them moment this is single session, ie multiple browsers will interact
    on a single worker state.
    Otherwise the worker will need to be derived from the request.

    The app manages:
    - Action state (running/stopped)
    - Auto-refresh polling during actions
    - Basic Template rendering with state
    - Customizable template directories
    - which contoller is being used
    - version

    """

    def __init__(
        self, template_dir: str = "templates", controller: Optional[Controller] = None
    ) -> None:
        super(App, self).__init__()
        # Attach templates helper
        self.templates = Jinja2Templates(directory=template_dir)
        self._controller: Optional[Controller] = None  # Private storage for controller
        self.controller = controller  # Use property setter for initialization
        self.startup = True  # If you go to an action page before you
        # have triggered it nothing will display.  startup finds this condition.
        self.startup_bounce_count = 0  # prevent endless loop
        self.refresh_time = 1
        self.version = "Lofigui"
        self.poll = False
        self.poll_count = 0
        self._action_running = False

    @property
    def controller(self) -> Optional[Controller]:
        """Get the current controller."""
        return self._controller

    @controller.setter
    def controller(self, new_controller: Optional[Controller]) -> None:
        """
        Set a new controller with safe cleanup of existing controller.

        If there's an existing controller with a running action, this will
        safely attempt to stop it before replacing with the new controller.

        This setter is idempotent - if the same controller is being set again,
        no cleanup is performed and the running action continues.

        Args:
            new_controller: The new controller to set (or None to clear)
        """
        # If setting the same controller, do nothing (idempotent)
        if self._controller is new_controller:
            return

        # If there's an existing controller, try to clean it up
        if self._controller is not None:
            # Safely check if action is running
            try:
                if self.is_action_running():
                    self.end_action()
            except Exception:
                # Silently ignore any errors during cleanup check
                pass

        # Set the new controller
        self._controller = new_controller

    def state_dict(self, request: Request, extra: dict = {}) -> dict:
        """Merge in local state and from controller so that they can be overidden"""
        # Do the dict in this order so defaults can be overriden
        d: dict = {}
        d["request"] = request
        d["version"] = self.version
        d["results"] = buffer()
        d["controller_name"] = self.controller_name()

        if self.poll:
            d["polling"] = "Running"
            self.poll_count += 1
            d["refresh"] = f'<meta http-equiv="Refresh" content="{self.refresh_time}">'
        else:
            d["polling"] = "Stopped"
            self.poll_count = 0
            d["refresh"] = ""
        d["poll_count"] = self.poll_count

        if self.controller:
            d = d | self.controller.state_dict(extra=extra)
        else:
            d = d | extra.copy()
        return d

    def template_response(
        self, request: Request, templateName: str, extra: dict = {}
    ) -> HTMLResponse:
        if self.startup:  # To cover the first call being a refresh of an api endpoint
            self.startup = False
            self.startup_bounce_count += 1
            if self.startup_bounce_count <= 3 and request.url.path != "/":
                # Redirect to home page
                return HTMLResponse('<head><meta http-equiv="Refresh" content="0; URL=/"/></head>')
        # Do the dict in this order so defaults can be overriden
        d = self.state_dict(request, extra)
        return self.templates.TemplateResponse(request, name=templateName, context=d)

    def start_action(self, refresh_time: Optional[int] = 1) -> None:
        self.startup = False
        self._action_running = True
        self.poll = True
        if refresh_time is not None:
            self.refresh_time = refresh_time

        if self.controller:
            self.controller.start_subaction(refresh_time)

    def is_action_running(self) -> bool:
        """Check if an action is currently running."""
        return self._action_running

    def end_action(self) -> None:
        self._action_running = False
        self.poll = False
        if self.controller:
            self.controller.end_subaction()

    def controller_name(self) -> str:
        if self._controller is not None and hasattr(self._controller, "name"):
            return self.controller.name
        return "Lofigui no controller"


def create_app(template_dir: str = "templates", add_favicon=True, **fastapi_kwargs: Any) -> App:
    """
        Create a lofigui application with defaults.  This is a wrapper for FastAPI.
    **kwar
        This includes:
        - Automatic favicon.ico endpoint
        - Ready-to-use Jinja2Templates configured
        - Optional controller integration
    """
    app = App(template_dir, **fastapi_kwargs)

    if add_favicon:
        # Add favicon route automatically
        @app.get("/favicon.ico")
        async def favicon() -> Any:
            """Serve the lofigui favicon"""
            return get_favicon_response()

    return app
