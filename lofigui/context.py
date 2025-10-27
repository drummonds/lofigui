import asyncio
from typing import Optional, Any


class PrintContext:
    """Context manager for buffering HTML output.

    This class manages a queue for collecting HTML fragments and a buffer
    for accumulated output. Supports both single-threaded and async usage.

    The context can be used as a context manager for automatic cleanup:
        >>> with PrintContext() as ctx:
        ...     # Use ctx
        ...     pass  # Buffer is automatically reset on exit

    Attributes:
        queue: AsyncIO queue for collecting HTML fragments.
        buffer: Accumulated HTML output as a string.
    """

    def __init__(self, max_buffer_size: Optional[int] = None):
        """Initialize a new PrintContext.

        Args:
            max_buffer_size: Optional maximum buffer size in characters. If the buffer
                           exceeds this size, a warning will be logged (but not enforced).
                           Use None for unlimited buffer (default).
        """
        self.queue: asyncio.Queue = asyncio.Queue()
        self.buffer: str = ""
        self.max_buffer_size = max_buffer_size

    def read(self) -> None:
        """Drain the queue and append all items to the buffer.

        This method reads all available items from the queue without blocking
        and appends them to the internal buffer string.
        """
        if self.queue.empty():
            return

        response = ""
        try:
            while not self.queue.empty():
                # Get a "work item" out of the queue.
                response += self.queue.get_nowait()
                self.queue.task_done()
        except asyncio.QueueEmpty:
            # This shouldn't happen due to the empty() check, but handle it gracefully
            pass
        except Exception as e:
            raise RuntimeError(f"Failed to read from queue: {e}") from e

        self.buffer += response

        # Warn if buffer is getting large (optional feature)
        if self.max_buffer_size and len(self.buffer) > self.max_buffer_size:
            import warnings

            warnings.warn(
                f"Buffer size ({len(self.buffer)} chars) exceeds max_buffer_size "
                f"({self.max_buffer_size} chars). Consider calling reset() more frequently.",
                RuntimeWarning,
            )

    def __enter__(self) -> "PrintContext":
        """Enter context manager."""
        return self

    def __exit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        """Exit context manager and reset buffer."""
        self.buffer = ""


# Default global context for single-threaded use
_ctx = PrintContext()


def buffer(ctx: Optional[PrintContext] = None) -> str:
    """Get the accumulated buffer content.

    Drains the queue and returns all accumulated HTML output.

    Args:
        ctx: Optional PrintContext to use. If None, uses the default global context.

    Returns:
        The accumulated HTML buffer as a string.

    Example:
        >>> import lofigui as lg
        >>> lg.print("Hello")
        >>> content = lg.buffer()
        >>> print(content)
        <p>Hello</p>
    """
    if ctx is None:
        ctx = _ctx
    ctx.read()  # drain buffer
    return ctx.buffer


def reset(ctx: Optional[PrintContext] = None) -> None:
    """Clear the buffer.

    Resets the buffer to an empty string. Does not clear the queue.

    Args:
        ctx: Optional PrintContext to use. If None, uses the default global context.

    Example:
        >>> import lofigui as lg
        >>> lg.print("Hello")
        >>> lg.reset()
        >>> lg.buffer()  # Returns empty string
        ''
    """
    if ctx is None:
        ctx = _ctx
    ctx.buffer = ""
