"""Tests for lofigui.markdown module."""

import pytest
from lofigui import markdown, html, table
from lofigui.context import PrintContext, buffer, reset


class TestMarkdown:
    """Test the markdown function."""

    def test_markdown_basic(self):
        """Test basic markdown conversion."""
        ctx = PrintContext()
        markdown("# Hello", ctx=ctx)
        result = buffer(ctx)
        assert "<h1>Hello</h1>" in result

    def test_markdown_bold(self):
        """Test markdown bold syntax."""
        ctx = PrintContext()
        markdown("**bold text**", ctx=ctx)
        result = buffer(ctx)
        assert "<strong>bold text</strong>" in result

    def test_markdown_empty(self):
        """Test markdown with empty string."""
        ctx = PrintContext()
        markdown("", ctx=ctx)
        result = buffer(ctx)
        assert result == ""

    def test_markdown_default_context(self):
        """Test markdown with default global context."""
        reset()
        markdown("## Test")
        result = buffer()
        assert "<h2>Test</h2>" in result


class TestHtml:
    """Test the html function."""

    def test_html_basic(self):
        """Test basic HTML passthrough."""
        ctx = PrintContext()
        html("<div>Test</div>", ctx=ctx)
        result = buffer(ctx)
        assert result == "<div>Test</div>"

    def test_html_no_escape(self):
        """Test that HTML is not escaped."""
        ctx = PrintContext()
        html("<script>alert('test')</script>", ctx=ctx)
        result = buffer(ctx)
        assert "<script>" in result

    def test_html_default_context(self):
        """Test html with default global context."""
        reset()
        html("<span>Test</span>")
        result = buffer()
        assert result == "<span>Test</span>"


class TestTable:
    """Test the table function."""

    def test_table_basic(self):
        """Test basic table generation."""
        ctx = PrintContext()
        data = [["Alice", 30], ["Bob", 25]]
        table(data, ctx=ctx)
        result = buffer(ctx)
        assert "<table" in result
        assert "Alice" in result
        assert "Bob" in result
        assert "30" in result
        assert "25" in result

    def test_table_with_header(self):
        """Test table with header."""
        ctx = PrintContext()
        data = [["Alice", 30], ["Bob", 25]]
        header = ["Name", "Age"]
        table(data, header=header, ctx=ctx)
        result = buffer(ctx)
        assert "<thead>" in result
        assert "<th>Name</th>" in result
        assert "<th>Age</th>" in result

    def test_table_escape_html(self):
        """Test that table content is escaped by default."""
        ctx = PrintContext()
        data = [["<script>alert('xss')</script>", "test"]]
        table(data, ctx=ctx)
        result = buffer(ctx)
        assert "<script>" not in result or "&lt;script&gt;" in result

    def test_table_no_escape(self):
        """Test table with escape=False."""
        ctx = PrintContext()
        data = [["<b>Bold</b>", "test"]]
        table(data, ctx=ctx, escape=False)
        result = buffer(ctx)
        assert "<b>Bold</b>" in result

    def test_table_colspan(self):
        """Test table with colspan for last field."""
        ctx = PrintContext()
        data = [["Alice", 30], ["Long message"]]
        header = ["Name", "Age"]
        table(data, header=header, ctx=ctx)
        result = buffer(ctx)
        assert 'colspan="2"' in result

    def test_table_empty(self):
        """Test table with empty data."""
        ctx = PrintContext()
        table([], ctx=ctx)
        result = buffer(ctx)
        assert "<table" in result
        assert "</table>" in result

    def test_table_bulma_classes(self):
        """Test that table has Bulma CSS classes."""
        ctx = PrintContext()
        data = [["test"]]
        table(data, ctx=ctx)
        result = buffer(ctx)
        assert 'class="table is-bordered is-striped"' in result

    def test_table_default_context(self):
        """Test table with default global context."""
        reset()
        data = [["test"]]
        table(data)
        result = buffer()
        assert "test" in result

    def test_table_invalid_structure(self):
        """Test table with invalid structure raises error."""
        ctx = PrintContext()
        with pytest.raises(ValueError):
            table([1, 2, 3], ctx=ctx)  # Not iterable rows
