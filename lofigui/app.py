"""
App for lofigui.

Provides a state manager for lofigui applications with template rendering
via Jinja2. Framework-agnostic — works with FastAPI, Flask, or plain HTTP.
"""

from typing import Optional, Any, Dict
from jinja2 import Environment, FileSystemLoader, BaseLoader
from .controller import Controller
from .context import buffer


class App:
    """State manager for lofigui applications.

    Manages action state (running/stopped), auto-refresh polling,
    template rendering, and controller lifecycle. Single-session:
    multiple browsers share one worker state.
    """

    def __init__(
        self, template_dir: str = "templates", controller: Optional[Controller] = None
    ) -> None:
        self.env = Environment(
            loader=FileSystemLoader(template_dir),
            autoescape=False,
        )
        self._controller: Optional[Controller] = None
        self.controller = controller
        self.startup = True
        self.startup_bounce_count = 0
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
        """Set a new controller with safe cleanup of existing controller."""
        if self._controller is new_controller:
            return
        if self._controller is not None:
            try:
                if self.is_action_running():
                    self.end_action()
            except Exception:
                pass
        self._controller = new_controller

    def state_dict(self, extra: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Build template context dictionary with current state."""
        if extra is None:
            extra = {}
        d: Dict[str, Any] = {}
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

    def render_template(self, template_name: str, extra: Optional[Dict[str, Any]] = None) -> str:
        """Render a template with current state, return HTML string."""
        d = self.state_dict(extra)
        template = self.env.get_template(template_name)
        return template.render(**d)

    def template_response(
        self, request: Any, template_name: str, extra: Optional[Dict[str, Any]] = None
    ) -> Any:
        """Render and return a FastAPI/Starlette HTMLResponse.

        This is a convenience wrapper for FastAPI apps. For framework-agnostic
        usage, use render_template() instead.
        """
        from starlette.responses import HTMLResponse

        if self.startup:
            self.startup = False
            self.startup_bounce_count += 1
            if (
                self.startup_bounce_count <= 3
                and hasattr(request, "url")
                and request.url.path != "/"
            ):
                return HTMLResponse('<head><meta http-equiv="Refresh" content="0; URL=/"/></head>')

        d = self.state_dict(extra)
        d["request"] = request
        html = self.env.get_template(template_name).render(**d)
        return HTMLResponse(html)

    def start_action(self, refresh_time: Optional[int] = 1) -> None:
        self.startup = False
        self._action_running = True
        self.poll = True
        if refresh_time is not None:
            self.refresh_time = refresh_time
        if self.controller:
            self.controller.start_subaction(refresh_time)

    def is_action_running(self) -> bool:
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


def create_app(
    template_dir: str = "templates", add_favicon: bool = True, **fastapi_kwargs: Any
) -> Any:
    """Create a FastAPI application wrapping a lofigui App.

    Requires FastAPI to be installed (pip install lofigui[examples]).
    Returns a FastAPI instance with app.lofigui set to the lofigui App.
    """
    from fastapi import FastAPI

    from .favicon import get_favicon_response

    lofigui_app = App(template_dir)
    fastapi_app = FastAPI(**fastapi_kwargs)
    fastapi_app.lofigui = lofigui_app  # type: ignore[attr-defined]

    if add_favicon:

        @fastapi_app.get("/favicon.ico")
        async def favicon() -> Any:
            return get_favicon_response()

    return fastapi_app
