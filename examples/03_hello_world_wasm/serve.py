#!/usr/bin/env python3
"""
Simple HTTP server for testing the WASM example locally.
Run this to serve the files and test in your browser.
"""

import http.server
import socketserver
import sys
from pathlib import Path

PORT = 8000


class MyHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    """Custom handler to set correct MIME types"""

    def end_headers(self):
        # Enable CORS for local development
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "*")
        super().end_headers()


def main():
    # Change to the directory containing this script
    script_dir = Path(__file__).parent
    os.chdir(script_dir) if hasattr(Path, "chdir") else None

    handler = MyHTTPRequestHandler

    with socketserver.TCPServer(("", PORT), handler) as httpd:
        print(f"\n{'='*60}")
        print(f"Lofigui WASM Example - Test Server")
        print(f"{'='*60}")
        print(f"\nServing on port {PORT}")
        print(f"Open your browser to:\n")
        print(f"    http://localhost:{PORT}")
        print(f"    http://127.0.0.1:{PORT}")
        print(f"\nPress Ctrl+C to stop the server")
        print(f"{'='*60}\n")

        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\n\nServer stopped.")
            sys.exit(0)


if __name__ == "__main__":
    main()
