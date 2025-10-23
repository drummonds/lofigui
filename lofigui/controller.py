"""
Controller for managing application state and routing.

The Controller class provides extensible logic for managing async actions,
template rendering, and state management. It can be customized for different
use cases by configuring template directories, refresh times, and custom
template context.
"""

import os
from pathlib import Path
from typing import Any, Callable, Optional, Dict
from .context import buffer, reset


class Controller:
    """
    Extensible controller class for managing state and routing.

    This controller manages:
    - Action state (running/stopped)
    - Auto-refresh polling during actions
    - Template rendering with state
    - Customizable template directories

    Example usage:
        # Basic usage with defaults
        controller = Controller()

        # Custom template directory
        controller = Controller(template_dir="my_templates")

        # Custom template and refresh time
        controller = Controller(
            template_dir="../templates",
            template_name="custom.html",
            refresh_time=2
        )
    """

    def __init__(
        self,
        template_dir: str = "templates",
        template_name: Optional[str] = None,
        refresh_time: int = 1,
    ):
        """
        Initialize the controller.

        Args:
            template_dir: Directory containing templates (default: "templates")
            template_name: Optional template filename to use (e.g., "hello.html")
            refresh_time: Seconds between auto-refresh when action is running (default: 1)
        """
        self.template_dir = template_dir
        self.template_name = template_name
        self.refresh_time = refresh_time
        self.poll = False
        self.poll_count = 0
        self._action_running = False

    def start_action(self, refresh_time: Optional[int] = None):
        """
        Start an action and enable polling/auto-refresh.

        Args:
            refresh_time: Optional override for refresh time in seconds
        """
        self._action_running = True
        self.poll = True
        if refresh_time is not None:
            self.refresh_time = refresh_time

    def end_action(self):
        """Stop the action and disable polling/auto-refresh."""
        self._action_running = False
        self.poll = False

    def is_action_running(self) -> bool:
        """Check if an action is currently running."""
        return self._action_running

    def state_dict(
        self, request: Any = None, extra: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Generate template context dictionary with current state.

        Args:
            request: The HTTP request object (framework-specific)
            extra: Additional custom context to merge into the state dict

        Returns:
            Dictionary containing:
                - request: The request object
                - results: Buffer content from print/markdown calls
                - refresh: Meta tag for auto-refresh (if action is running)
                - poll_count: Number of polls (if action is running)
                - Any additional keys from extra dict
        """
        d = extra.copy() if extra else {}

        d["request"] = request
        d["results"] = buffer()

        if self.poll:
            d["poll_count"] = self.poll_count
            self.poll_count += 1
            d["refresh"] = f'<meta http-equiv="Refresh" content="{self.refresh_time}">'
        else:
            self.poll_count = 0
            d["refresh"] = ""

        return d

    def get_template_path(self, template_name: Optional[str] = None) -> str:
        """
        Get the full path to a template file.

        Args:
            template_name: Optional template name to use instead of default

        Returns:
            Full path to template file

        Raises:
            ValueError: If no template name is provided and none was set during init
        """
        name = template_name or self.template_name
        if not name:
            raise ValueError("No template name specified")

        return os.path.join(self.template_dir, name)

    def handle_root(
        self,
        model_func: Callable,
        redirect_url: str = "/display",
        reset_buffer: bool = True,
    ) -> str:
        """
        Handle the root endpoint that starts an action.

        This is a helper method that:
        1. Optionally resets the buffer
        2. Starts the action
        3. Triggers the model function (should be run in background by caller)
        4. Returns HTML to redirect to display page

        Args:
            model_func: Function to run in background (caller must handle async execution)
            redirect_url: URL to redirect to (default: "/display")
            reset_buffer: Whether to reset the buffer (default: True)

        Returns:
            HTML string with meta refresh redirect

        Example (FastAPI):
            @app.get("/")
            async def root(background_tasks: BackgroundTasks):
                background_tasks.add_task(model)
                return controller.handle_root(model)
        """
        if reset_buffer:
            reset()

        self.start_action()

        return (
            f'<head><meta http-equiv="Refresh" content="0; URL={redirect_url}"/></head>'
        )
