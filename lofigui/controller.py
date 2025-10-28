"""
Controller for managing application state and routing.

The Controller class provides extensible logic for managing async actions,
template rendering, and state management. It can be customized for different
use cases by configuring template directories, refresh times, and custom
template context.
"""

from typing import Any, Optional, Dict
from .context import buffer, reset


class Controller:
    """
    Extensible controller class for managing state and routing.
    This is actually a sub controller with the default being set by
    the app.
    """

    def __init__(
        self,
    ) -> None:
        """
        Initialize the controller.

        Args:
            template_dir: Directory containing templates (default: "templates")
            template_name: Optional template filename to use (e.g., "hello.html")
            refresh_time: Seconds between auto-refresh when action is running (default: 1)
        """
        self.name = "Demo Controller"

    # actions will be called by app and are to do any specific commands

    def start_subaction(self, refresh_time: Optional[int] = None) -> None:
        """
        Start.

        Args:
            refresh_time: Optional override for refresh time in seconds
        """
        pass

    def end_subaction(self) -> None:
        """Stop the action and disable polling/auto-refresh."""
        pass

    def state_dict(self, extra: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
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

        return d
