#!/usr/bin/env python3
"""
Simple HTTP server for testing the Go WASM example locally.
Sets correct MIME type for .wasm files.
"""

import http.server
import socketserver

PORT = 8000


class WASMHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    """Custom handler to set correct MIME types for WASM"""

    def end_headers(self):
        # Set WASM MIME type
        if self.path.endswith('.wasm'):
            self.send_header('Content-Type', 'application/wasm')

        # Enable CORS for local development
        self.send_header("Access-Control-Allow-Origin", "*")
        super().end_headers()


def main():
    handler = WASMHTTPRequestHandler

    with socketserver.TCPServer(("", PORT), handler) as httpd:
        print(f"\n{'='*60}")
        print(f"Lofigui Go WASM Example - Test Server")
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


if __name__ == "__main__":
    main()
