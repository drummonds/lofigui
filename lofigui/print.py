import html
from typing import Optional

from .context import _ctx, PrintContext


def print(
    msg: str = "", ctx: Optional[PrintContext] = None, end: str = "\n", escape: bool = True
) -> None:
    """Print text to the buffer as HTML.

    Args:
        msg: The message to print. Will be wrapped in <p> tags for newline, or &nbsp; for inline.
        ctx: Optional PrintContext to use. If None, uses the default global context.
        end: The end character. Use "\\n" for paragraph mode (default), or "" for inline mode.
        escape: Whether to escape HTML entities in the message (default: True). Set to False
                to allow raw HTML, but be careful of XSS vulnerabilities.

    Example:
        >>> import lofigui as lg
        >>> lg.print("Hello world")
        >>> lg.print("Inline text", end="")
        >>> lg.print("<script>alert('safe')</script>")  # Escaped by default
        >>> lg.print("<b>Bold</b>", escape=False)  # Raw HTML (use with caution)
    """
    if ctx is None:
        ctx = _ctx

    if escape:
        msg = html.escape(str(msg))

    try:
        if end == "\n":
            ctx.queue.put_nowait(f"<p>{msg}</p>\n")
        else:
            ctx.queue.put_nowait(f"&nbsp;{msg}&nbsp;")
    except Exception as e:
        raise RuntimeError(f"Failed to add message to print buffer: {e}") from e
