from .context import buffer


class Controller:
    """Async controller class, but also combined with model"""

    def __init__(self):
        self.poll = False
        self.poll_count = 0
        self.refresh_time = 1

    def state_dict(self, request, d={}):
        d["request"] = request
        d["results"] = buffer()
        if self.poll:
            d["poll_count"] = self.poll_count
            self.poll_count += 1
            d["refresh"] = f'<meta http-equiv="Refresh" content="{self.refresh_time}">'
        else:
            self.poll_count = 0
            d["refresh"] = ""
        return d

    def start_action(self, refresh_time=1):
        self.poll = True

    def end_action(self):
        self.poll = False
