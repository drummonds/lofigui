import markdown as mkdwn

from .context import _ctx


def markdown(msg="", ctx=None):
    if ctx is None:
        ctx = _ctx
    md = mkdwn.markdown(msg)
    ctx.queue.put_nowait(md)


def html(msg="", ctx=None):
    if ctx is None:
        ctx = _ctx
    ctx.queue.put_nowait(msg)


def table(table, header=[], ctx=None):
    if ctx is None:
        ctx = _ctx
    result = '<table class="table is-bordered is-striped">\n'
    if header:
        result += "  <thead><tr>\n"
        for field in header:
            result += f"    <th>{field}</th>\n"
        result += "  </tr></thead>\n"
    if table:
        result += "  <tbody>\n"
        for row in table:
            # Make last field expand eg use one field to go alway across
            extend_last_field = header and len(header) > len(row)
            result += "    <tr>\n"
            for i, field in enumerate(row):
                if extend_last_field and i == len(row) - 1:
                    result += f'      <td colspan="{len(header)-i}">{field}</td>\n'
                else:
                    result += f"      <td>{field}</td>\n"
            result += "    </tr>\n"
        result += "  </tbody>\n"
    result += "</table>\n"
    ctx.queue.put_nowait(result)
