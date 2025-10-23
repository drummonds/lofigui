"""Tests for lofigui.favicon module."""

import pytest
from lofigui.favicon import (
    get_favicon_ico,
    get_favicon_svg,
    get_favicon_data_uri,
    get_favicon_html_tag,
    get_favicon_response,
)


class TestFaviconBasics:
    """Test basic favicon functions."""

    def test_get_favicon_ico(self):
        """Test that get_favicon_ico returns bytes."""
        favicon = get_favicon_ico()
        assert isinstance(favicon, bytes)
        assert len(favicon) > 0
        # ICO files start with specific header
        assert favicon[0:2] == b"\x00\x00"

    def test_get_favicon_svg(self):
        """Test that get_favicon_svg returns SVG string."""
        svg = get_favicon_svg()
        assert isinstance(svg, str)
        assert "<svg" in svg
        assert "</svg>" in svg
        assert "xmlns" in svg

    def test_get_favicon_data_uri(self):
        """Test that get_favicon_data_uri returns valid data URI."""
        data_uri = get_favicon_data_uri()
        assert isinstance(data_uri, str)
        assert data_uri.startswith("data:image/x-icon;base64,")
        # Should have base64 encoded data after the prefix
        assert len(data_uri) > len("data:image/x-icon;base64,")

    def test_get_favicon_html_tag(self):
        """Test that get_favicon_html_tag returns valid HTML."""
        html_tag = get_favicon_html_tag()
        assert isinstance(html_tag, str)
        assert html_tag.startswith('<link rel="icon"')
        assert 'href="data:image/x-icon;base64,' in html_tag
        assert html_tag.endswith(">")


class TestFaviconResponse:
    """Test favicon response function for web frameworks."""

    def test_get_favicon_response_returns_response(self):
        """Test that get_favicon_response returns a Response object."""
        response = get_favicon_response()
        assert response is not None
        # Check if it's a FastAPI Response or fallback SimpleResponse
        # Starlette/FastAPI Response uses 'body', SimpleResponse uses 'content'
        assert hasattr(response, "body") or hasattr(response, "content")
        assert hasattr(response, "media_type")

    def test_favicon_response_content_type(self):
        """Test that favicon response has correct media type."""
        response = get_favicon_response()
        assert response.media_type == "image/x-icon"

    def test_favicon_response_content_is_bytes(self):
        """Test that favicon response content is bytes."""
        response = get_favicon_response()
        # Starlette/FastAPI Response uses 'body', SimpleResponse uses 'content'
        body = getattr(response, "body", None) or getattr(response, "content", None)
        assert isinstance(body, bytes)
        assert len(body) > 0

    def test_favicon_response_content_is_valid_ico(self):
        """Test that favicon response content is valid ICO format."""
        response = get_favicon_response()
        # Starlette/FastAPI Response uses 'body', SimpleResponse uses 'content'
        body = getattr(response, "body", None) or getattr(response, "content", None)
        # ICO files start with 0x0000
        assert body[0:2] == b"\x00\x00"

    def test_favicon_response_matches_get_favicon_ico(self):
        """Test that response content matches get_favicon_ico output."""
        response = get_favicon_response()
        direct_ico = get_favicon_ico()
        # Starlette/FastAPI Response uses 'body', SimpleResponse uses 'content'
        body = getattr(response, "body", None) or getattr(response, "content", None)
        assert body == direct_ico


class TestFaviconIntegration:
    """Integration tests for favicon in web context."""

    @pytest.mark.asyncio
    async def test_favicon_in_fastapi_context(self):
        """Test that favicon works in FastAPI context."""
        try:
            from fastapi import FastAPI
            from fastapi.testclient import TestClient

            app = FastAPI()

            @app.get("/favicon.ico")
            async def favicon():
                return get_favicon_response()

            client = TestClient(app)
            response = client.get("/favicon.ico")

            assert response.status_code == 200
            assert response.headers["content-type"] == "image/x-icon"
            assert len(response.content) > 0
            # Verify it's a valid ICO file
            assert response.content[0:2] == b"\x00\x00"

        except ImportError:
            pytest.skip("FastAPI not available for integration test")
