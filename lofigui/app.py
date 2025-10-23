"""
FastAPI app factory for lofigui
Provides a pre-configured FastAPI application with favicon support
"""

from fastapi import FastAPI
from fastapi.templating import Jinja2Templates
from .favicon import get_favicon_response


def create_app(template_dir: str = "templates", **fastapi_kwargs) -> FastAPI:
    """
    Create a FastAPI application with lofigui defaults

    This includes:
    - Automatic favicon.ico endpoint
    - Ready-to-use Jinja2Templates configured

    Args:
        template_dir: Directory containing Jinja2 templates (default: "templates")
        **fastapi_kwargs: Additional keyword arguments to pass to FastAPI()

    Returns:
        Configured FastAPI application instance

    Example:
        from lofigui import create_app

        app = create_app()

        @app.get("/")
        async def home():
            return {"message": "Hello World"}
    """
    app = FastAPI(**fastapi_kwargs)

    # Add favicon route automatically
    @app.get("/favicon.ico")
    async def favicon():
        """Serve the lofigui favicon"""
        return get_favicon_response()

    # Attach templates helper if needed
    app.templates = Jinja2Templates(directory=template_dir)

    return app
