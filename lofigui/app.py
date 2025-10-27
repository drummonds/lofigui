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


class App(FastAPI):
    """At them moment this is single session, ie multiple browsers will interact
    on a single worker state.
    Otherwise the worker will need to be derived from the request"""

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
                if hasattr(self._controller, "is_action_running") and callable(
                    self._controller.is_action_running
                ):
                    if self._controller.is_action_running():
                        # Action is running, try to end it
                        if hasattr(self._controller, "end_action") and callable(
                            self._controller.end_action
                        ):
                            try:
                                self._controller.end_action()
                            except Exception:
                                # Silently ignore errors during cleanup
                                # We're replacing the controller anyway
                                pass
            except Exception:
                # Silently ignore any errors during cleanup check
                pass

        # Set the new controller
        self._controller = new_controller

    def template_response(
        self, request: Request, templateName: str, extra: dict = {}
    ) -> HTMLResponse:
        if self.startup:
            self.startup = False
            self.startup_bounce_count += 1
            if self.startup_bounce_count <= 3 and request.url.path != "/":
                # Redirect to home page
                return HTMLResponse('<head><meta http-equiv="Refresh" content="0; URL=/"/></head>')
        if self.controller:
            d = self.controller.state_dict(extra=extra)
        else:
            d = extra.copy()
        d["request"] = request
        return self.templates.TemplateResponse(templateName, d)

    def start_action(self, refresh_time: Optional[int] = 1) -> None:
        self.startup = False
        if self.controller:
            self.controller.start_action(refresh_time)

    def is_action_running(self) -> bool:
        """Check if an action is currently running."""
        if self.controller:
            return self.controller.is_action_running()
        return False

    def end_action(self) -> None:
        if self.controller:
            self.controller.end_action()


def create_app(template_dir: str = "templates", **fastapi_kwargs: Any) -> App:
    """
        Create a lofigui application with defaults.  This is a wrapper for FastAPI.
    **kwar
        This includes:
        - Automatic favicon.ico endpoint
        - Ready-to-use Jinja2Templates configured
        - Optional controller integration
    """
    app = App(template_dir, **fastapi_kwargs)

    # Add favicon route automatically
    @app.get("/favicon.ico")
    async def favicon() -> Any:
        """Serve the lofigui favicon"""
        return get_favicon_response()

    return app
