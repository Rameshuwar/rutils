#!/usr/bin/env python3
import http.server
import socketserver
import sys
import signal

try:
    signal.signal(signal.SIGHUP, signal.SIG_IGN)
except Exception:
    pass

class NoCacheHandler(http.server.SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header('Cache-Control', 'no-store, no-cache, must-revalidate, max-age=0')
        self.send_header('Pragma', 'no-cache')
        self.send_header('Expires', '0')
        super().end_headers()

def run():
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8087
    web_dir = sys.argv[2] if len(sys.argv) > 2 else "build/web"

    socketserver.TCPServer.allow_reuse_address = True
    handler = lambda *args, **kwargs: NoCacheHandler(*args, directory=web_dir, **kwargs)

    with socketserver.TCPServer(('', port), handler) as httpd:
        print(f"Serving {web_dir} on http://0.0.0.0:{port} with no-cache headers...")
        sys.stdout.flush()
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            pass

if __name__ == '__main__':
    run()
