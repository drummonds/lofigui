#!/usr/bin/env python3
"""Static file server with SeaweedFS API proxy for WASM demo."""

import http.server
import mimetypes
import os
import socketserver
import urllib.error
import urllib.request

PORT = 1351
MASTER_URL = "http://localhost:9333"
STATIC_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "templates")

mimetypes.add_type("application/wasm", ".wasm")


class Handler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=STATIC_DIR, **kwargs)

    def _proxy(self, target_url):
        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length) if length > 0 else None

        req = urllib.request.Request(target_url, data=body, method=self.command)
        ct = self.headers.get("Content-Type")
        if ct:
            req.add_header("Content-Type", ct)

        try:
            with urllib.request.urlopen(req, timeout=10) as resp:
                data = resp.read()
                self.send_response(resp.status)
                self.send_header(
                    "Content-Type",
                    resp.headers.get("Content-Type", "application/octet-stream"),
                )
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                self.wfile.write(data)
        except urllib.error.HTTPError as e:
            data = e.read()
            self.send_response(e.code)
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)
        except (urllib.error.URLError, TimeoutError) as e:
            msg = str(e).encode()
            self.send_response(502)
            self.send_header("Content-Length", str(len(msg)))
            self.end_headers()
            self.wfile.write(msg)

    def _handle_api(self):
        if self.path.startswith("/api/master/"):
            target = MASTER_URL + self.path[len("/api/master") :]
            self._proxy(target)
            return True
        elif self.path.startswith("/api/vol/"):
            rest = self.path[len("/api/vol/") :]
            target = "http://" + rest
            self._proxy(target)
            return True
        return False

    def do_GET(self):
        if not self._handle_api():
            super().do_GET()

    def do_POST(self):
        if not self._handle_api():
            self.send_error(405)

    def do_DELETE(self):
        if not self._handle_api():
            self.send_error(405)


def main():
    socketserver.TCPServer.allow_reuse_address = True
    with socketserver.TCPServer(("", PORT), Handler) as httpd:
        print(f"\nSeaweedFS WASM Demo")
        print(f"Serving on http://localhost:{PORT}")
        print(f"Proxying SeaweedFS master at {MASTER_URL}")
        print(f"Press Ctrl+C to stop\n")
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\nStopped.")


if __name__ == "__main__":
    main()
