"""
Favicon utilities for lofigui
Provides a simple favicon that can be served with your application
"""

import os
from pathlib import Path
from typing import Any, Union

# Base64 encoded favicon.ico (16x16 pixels, simple "L" logo)
# This is a minimal ICO file that works in all browsers
FAVICON_ICO_BASE64 = (
    "AAABAAEAEBAQAAEABAAoAQAAFgAAACgAAAAQAAAAIAAAAAEABAAAAAAAgAAAAAAAAAAAAAAAEAA"
    + "AAAAAAAAAAAAAMnPcAP///wAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
    + "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
    + "QEQEQEQEQERERERERERERERERERERERERERERERERERERERERERERERERERERERERERERERERE"
    + "REf////8P////D////w////8P////D////w////8P////D////w////8AAAAA=="
)

# Path to SVG favicon
STATIC_DIR = Path(__file__).parent / "static"
FAVICON_SVG_PATH = STATIC_DIR / "favicon.svg"


def get_favicon_ico() -> bytes:
    """
    Get the favicon as ICO format (bytes)
    Returns the base64-decoded ICO file
    """
    import base64

    return base64.b64decode(FAVICON_ICO_BASE64)


def get_favicon_svg() -> str:
    """
    Get the favicon as SVG format (string)
    Returns the SVG content
    """
    if FAVICON_SVG_PATH.exists():
        return FAVICON_SVG_PATH.read_text()
    else:
        # Fallback inline SVG
        return """<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
  <rect width="32" height="32" fill="#3273dc" rx="4"/>
  <path d="M 10 8 L 10 24 L 22 24 L 22 21 L 13 21 L 13 8 Z" fill="#ffffff"/>
</svg>"""


def get_favicon_data_uri() -> str:
    """
    Get the favicon as a data URI that can be used directly in HTML
    Returns: data:image/x-icon;base64,... string
    """
    return f"data:image/x-icon;base64,{FAVICON_ICO_BASE64}"


def get_favicon_html_tag() -> str:
    """
    Get an HTML link tag for the favicon using data URI
    Can be inserted directly into <head> section

    Example:
        <head>
            {get_favicon_html_tag()}
            ...
        </head>
    """
    return f'<link rel="icon" type="image/x-icon" href="{get_favicon_data_uri()}">'


# For FastAPI/Starlette integration
def get_favicon_response() -> Any:
    """
    Get a Response object for serving the favicon
    Works with FastAPI, Starlette, etc.

    Usage in FastAPI:
        from lofigui.favicon import get_favicon_response

        @app.get("/favicon.ico")
        async def favicon():
            return get_favicon_response()
    """
    try:
        from fastapi.responses import Response

        return Response(content=get_favicon_ico(), media_type="image/x-icon")
    except ImportError:
        # Fallback for plain Python
        class SimpleResponse:
            def __init__(self, content: bytes, media_type: str) -> None:
                self.content = content
                self.media_type = media_type

        return SimpleResponse(content=get_favicon_ico(), media_type="image/x-icon")


def save_favicon_ico(path: str) -> None:
    """
    Save the favicon.ico file to disk

    Args:
        path: Path where to save the favicon.ico file
    """
    favicon_path = Path(path)
    favicon_path.write_bytes(get_favicon_ico())
    print(f"Favicon saved to: {favicon_path}")


def add_favicon_route(app: Any) -> None:
    """
    Add the default favicon route to a FastAPI app

    Usage:
        from lofigui import add_favicon_route

        app = FastAPI()
        add_favicon_route(app)

    Args:
        app: FastAPI application instance
    """

    @app.get("/favicon.ico")
    async def favicon() -> Any:
        """Serve the lofigui favicon"""
        return get_favicon_response()


if __name__ == "__main__":
    # CLI tool to save favicon
    import sys

    if len(sys.argv) > 1:
        save_favicon_ico(sys.argv[1])
    else:
        save_favicon_ico("favicon.ico")
        print("\nYou can also use:")
        print("  python -m lofigui.favicon path/to/favicon.ico")
        print("\nOr in your code:")
        print("  from lofigui.favicon import get_favicon_html_tag")
        print("  # Insert into <head>: get_favicon_html_tag()")
