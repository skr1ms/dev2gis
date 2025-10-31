import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from threading import Thread
from heightmap_service.core.metrics import get_metrics, get_metrics_content_type, http_requests_total, http_request_duration
from heightmap_service.core.logging import logger


class MetricsHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        start_time = time.time()
        path = self.path
        method = self.command
        status = 200

        try:
            if path == '/metrics':
                self.send_response(200)
                self.send_header('Content-Type', get_metrics_content_type())
                self.end_headers()
                self.wfile.write(get_metrics())
            elif path == '/health':
                self.send_response(200)
                self.send_header('Content-Type', 'text/plain')
                self.end_headers()
                self.wfile.write(b'OK')
            else:
                status = 404
                self.send_response(404)
                self.end_headers()
        finally:
            duration = time.time() - start_time
            http_requests_total.labels(
                method=method, path=path, status=str(status)).inc()
            http_request_duration.labels(
                method=method, path=path).observe(duration)

            logger.info("HTTP Request", extra={
                'method': method,
                'path': path,
                'status': status,
                'latency': f"{duration:.3f}s"
            })

    def log_message(self, format, *args):
        pass


def start_metrics_server(port: int = 8000):
    server = HTTPServer(('0.0.0.0', port), MetricsHandler)
    thread = Thread(target=server.serve_forever, daemon=True)
    thread.start()
    logger.info("Metrics server started", extra={'port': port})
    return server
