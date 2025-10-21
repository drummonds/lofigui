"""Tests for lofigui.print module."""

import pytest
from lofigui import print as lg_print
from lofigui.context import PrintContext, buffer, reset


class TestPrint:
    """Test the print function."""

    def test_print_basic(self):
        """Test basic print functionality."""
        ctx = PrintContext()
        lg_print("Hello world", ctx=ctx)
        result = buffer(ctx)
        assert result == "<p>Hello world</p>\n"

    def test_print_inline(self):
        """Test print with inline mode."""
        ctx = PrintContext()
        lg_print("Inline", ctx=ctx, end="")
        result = buffer(ctx)
        assert result == "&nbsp;Inline&nbsp;"

    def test_print_escape_html(self):
        """Test that HTML is escaped by default."""
        ctx = PrintContext()
        lg_print("<script>alert('xss')</script>", ctx=ctx)
        result = buffer(ctx)
        assert "<script>" not in result
        assert "&lt;script&gt;" in result

    def test_print_no_escape(self):
        """Test print with escape=False."""
        ctx = PrintContext()
        lg_print("<b>Bold</b>", ctx=ctx, escape=False)
        result = buffer(ctx)
        assert result == "<p><b>Bold</b></p>\n"

    def test_print_default_context(self):
        """Test print with default global context."""
        reset()
        lg_print("Test")
        result = buffer()
        assert "Test" in result

    def test_print_multiple_messages(self):
        """Test multiple print calls accumulate."""
        ctx = PrintContext()
        lg_print("First", ctx=ctx)
        lg_print("Second", ctx=ctx)
        result = buffer(ctx)
        assert "<p>First</p>" in result
        assert "<p>Second</p>" in result

    def test_print_converts_to_string(self):
        """Test that non-string values are converted to string."""
        ctx = PrintContext()
        lg_print(123, ctx=ctx)
        result = buffer(ctx)
        assert "123" in result

    def test_print_empty_message(self):
        """Test print with empty message."""
        ctx = PrintContext()
        lg_print("", ctx=ctx)
        result = buffer(ctx)
        assert result == "<p></p>\n"
