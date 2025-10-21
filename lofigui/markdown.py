import html as html_lib
from typing import Optional, List, Sequence, Any

import markdown as mkdwn

from .context import _ctx, PrintContext


def markdown(msg: str = "", ctx: Optional[PrintContext] = None) -> None:
    """Convert markdown text to HTML and add to buffer.

    Args:
        msg: Markdown-formatted text to convert to HTML.
        ctx: Optional PrintContext to use. If None, uses the default global context.

    Raises:
        RuntimeError: If markdown conversion or buffer addition fails.

    Example:
        >>> import lofigui as lg
        >>> lg.markdown("# Hello\\n\\nThis is **bold** text")
    """
    if ctx is None:
        ctx = _ctx

    try:
        md = mkdwn.markdown(str(msg))
        ctx.queue.put_nowait(md)
    except Exception as e:
        raise RuntimeError(f"Failed to convert markdown or add to buffer: {e}") from e


def html(msg: str = "", ctx: Optional[PrintContext] = None) -> None:
    """Add raw HTML to the buffer.

    WARNING: This function does not escape HTML. Only use with trusted input
    to avoid XSS vulnerabilities.

    Args:
        msg: Raw HTML string to add to buffer.
        ctx: Optional PrintContext to use. If None, uses the default global context.

    Raises:
        RuntimeError: If buffer addition fails.

    Example:
        >>> import lofigui as lg
        >>> lg.html("<div class='custom'>Custom HTML</div>")
    """
    if ctx is None:
        ctx = _ctx

    try:
        ctx.queue.put_nowait(str(msg))
    except Exception as e:
        raise RuntimeError(f"Failed to add HTML to buffer: {e}") from e


def table(
    table: Sequence[Sequence[Any]],
    header: List[str] = None,
    ctx: Optional[PrintContext] = None,
    escape: bool = True,
) -> None:
    """Generate an HTML table and add to buffer.

    Creates a styled table using Bulma CSS classes. Supports colspan for the last
    field if it needs to extend across remaining columns.

    Args:
        table: A sequence of rows, where each row is a sequence of cell values.
        header: Optional list of header column names.
        ctx: Optional PrintContext to use. If None, uses the default global context.
        escape: Whether to escape HTML in cell content (default: True).

    Raises:
        ValueError: If table structure is invalid.
        RuntimeError: If buffer addition fails.

    Example:
        >>> import lofigui as lg
        >>> data = [["Alice", 30], ["Bob", 25]]
        >>> lg.table(data, header=["Name", "Age"])
    """
    if ctx is None:
        ctx = _ctx

    if header is None:
        header = []

    try:
        # Validate table structure
        if table and not all(hasattr(row, "__iter__") for row in table):
            raise ValueError("All table rows must be iterable")

        result = '<table class="table is-bordered is-striped">\n'

        if header:
            result += "  <thead><tr>\n"
            for field in header:
                escaped_field = html_lib.escape(str(field)) if escape else str(field)
                result += f"    <th>{escaped_field}</th>\n"
            result += "  </tr></thead>\n"

        if table:
            result += "  <tbody>\n"
            for row in table:
                # Make last field expand eg use one field to go alway across
                extend_last_field = header and len(header) > len(row)
                result += "    <tr>\n"
                for i, field in enumerate(row):
                    escaped_field = html_lib.escape(str(field)) if escape else str(field)
                    if extend_last_field and i == len(row) - 1:
                        result += f'      <td colspan="{len(header)-i}">{escaped_field}</td>\n'
                    else:
                        result += f"      <td>{escaped_field}</td>\n"
                result += "    </tr>\n"
            result += "  </tbody>\n"

        result += "</table>\n"
        ctx.queue.put_nowait(result)
    except ValueError:
        raise
    except Exception as e:
        raise RuntimeError(f"Failed to generate table or add to buffer: {e}") from e
