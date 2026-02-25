"""Tests for lofigui.context module."""

import pytest
import warnings
from lofigui.context import PrintContext, buffer, reset, _ctx
from lofigui import print as lg_print


class TestPrintContext:
    """Test the PrintContext class."""

    def test_context_init(self):
        """Test PrintContext initialization."""
        ctx = PrintContext()
        assert ctx.buffer == ""
        assert ctx.queue.empty()

    def test_context_init_with_max_size(self):
        """Test PrintContext initialization with max buffer size."""
        ctx = PrintContext(max_buffer_size=1000)
        assert ctx.max_buffer_size == 1000

    def test_context_read_empty(self):
        """Test reading from empty queue."""
        ctx = PrintContext()
        ctx.read()
        assert ctx.buffer == ""

    def test_context_read_with_data(self):
        """Test reading from queue with data."""
        ctx = PrintContext()
        ctx.queue.put_nowait("test1")
        ctx.queue.put_nowait("test2")
        ctx.read()
        assert ctx.buffer == "test1test2"

    def test_context_read_accumulates(self):
        """Test that multiple reads accumulate."""
        ctx = PrintContext()
        ctx.queue.put_nowait("first")
        ctx.read()
        ctx.queue.put_nowait("second")
        ctx.read()
        assert ctx.buffer == "firstsecond"

    def test_context_manager(self):
        """Test context manager protocol."""
        with PrintContext() as ctx:
            lg_print("test", ctx=ctx)
            assert "test" in buffer(ctx)
        # Buffer should be reset after exit
        assert ctx.buffer == ""

    def test_context_manager_with_exception(self):
        """Test context manager resets buffer even with exception."""
        ctx = PrintContext()
        try:
            with ctx:
                lg_print("test", ctx=ctx)
                raise ValueError("Test exception")
        except ValueError:
            pass
        assert ctx.buffer == ""

    def test_buffer_size_warning(self):
        """Test warning when buffer exceeds max size."""
        ctx = PrintContext(max_buffer_size=10)
        ctx.queue.put_nowait("This is a very long string that exceeds the max buffer size")

        with warnings.catch_warnings(record=True) as w:
            warnings.simplefilter("always")
            ctx.read()
            assert len(w) == 1
            assert issubclass(w[0].category, RuntimeWarning)
            assert "exceeds max_buffer_size" in str(w[0].message)


class TestBufferFunction:
    """Test the buffer function."""

    def test_buffer_with_context(self):
        """Test buffer function with custom context."""
        ctx = PrintContext()
        lg_print("test", ctx=ctx)
        result = buffer(ctx)
        assert "test" in result

    def test_buffer_default_context(self):
        """Test buffer function with default context."""
        reset()
        lg_print("test")
        result = buffer()
        assert "test" in result

    def test_buffer_drains_queue(self):
        """Test that buffer drains the queue."""
        ctx = PrintContext()
        ctx.queue.put_nowait("item1")
        ctx.queue.put_nowait("item2")
        result = buffer(ctx)
        assert result == "item1item2"
        assert ctx.queue.empty()


class TestResetFunction:
    """Test the reset function."""

    def test_reset_with_context(self):
        """Test reset function with custom context."""
        ctx = PrintContext()
        lg_print("test", ctx=ctx)
        buffer(ctx)
        reset(ctx)
        assert ctx.buffer == ""

    def test_reset_default_context(self):
        """Test reset function with default context."""
        lg_print("test")
        buffer()
        reset()
        assert _ctx.buffer == ""

    def test_reset_drains_queue(self):
        """Test that reset clears buffer and drains the queue."""
        ctx = PrintContext()
        ctx.queue.put_nowait("item")
        reset(ctx)
        assert ctx.buffer == ""
        assert ctx.queue.empty()


class TestGlobalContext:
    """Test the global _ctx context."""

    def test_global_context_exists(self):
        """Test that global context is initialized."""
        assert _ctx is not None
        assert isinstance(_ctx, PrintContext)

    def test_global_context_shared(self):
        """Test that global context is shared across calls."""
        reset()
        lg_print("first")
        lg_print("second")
        result = buffer()
        assert "first" in result
        assert "second" in result
