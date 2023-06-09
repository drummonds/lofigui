from .context import _ctx

def print(msg="", ctx=None, end="\n"):
    if ctx is None:
        ctx = _ctx
    if end == "\n":
        ctx.queue.put_nowait(f"<p>{msg}</p>\n")
    else:
        ctx.queue.put_nowait(f"&nbsp;{msg}&nbsp;")
    # await asyncio.sleep(0)  # Allow breaks


