import time
from prometheus_client import Counter, Histogram, Gauge, generate_latest, CONTENT_TYPE_LATEST
from functools import wraps


http_requests_total = Counter(
    'http_requests_total',
    'Total number of HTTP requests',
    ['method', 'path', 'status']
)

http_request_duration = Histogram(
    'http_request_duration_seconds',
    'HTTP request duration in seconds',
    ['method', 'path']
)

heightmap_requests_total = Counter(
    'heightmap_requests_total',
    'Total number of heightmap generation requests',
    ['status']
)

heightmap_processing_duration = Histogram(
    'heightmap_processing_duration_seconds',
    'Time spent processing heightmap',
    buckets=[0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0]
)

heightmap_processing_active = Gauge(
    'heightmap_processing_active',
    'Number of heightmaps currently being processed'
)

heightmap_errors_total = Counter(
    'heightmap_errors_total',
    'Total number of heightmap processing errors',
    ['error_type']
)

heightmap_file_size_bytes = Histogram(
    'heightmap_file_size_bytes',
    'Size of processed heightmap files in bytes',
    buckets=[1024, 10240, 102400, 1048576, 10485760, 104857600]
)

heightmap_image_dimensions = Histogram(
    'heightmap_image_dimensions_pixels',
    'Image dimensions in pixels',
    ['dimension'],
    buckets=[100, 500, 1000, 2000, 4000, 8000]
)


def track_processing_time(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        heightmap_processing_active.inc()
        start_time = time.time()
        try:
            result = func(*args, **kwargs)
            duration = time.time() - start_time
            heightmap_processing_duration.observe(duration)
            heightmap_requests_total.labels(status='success').inc()
            return result
        except Exception as e:
            duration = time.time() - start_time
            heightmap_processing_duration.observe(duration)
            heightmap_requests_total.labels(status='failed').inc()
            heightmap_errors_total.labels(error_type=type(e).__name__).inc()
            raise
        finally:
            heightmap_processing_active.dec()
    return wrapper


def record_file_size(size_bytes: int):
    heightmap_file_size_bytes.observe(size_bytes)


def record_image_dimension(width: int, height: int):
    heightmap_image_dimensions.labels(dimension='width').observe(width)
    heightmap_image_dimensions.labels(dimension='height').observe(height)


def get_metrics():
    return generate_latest()


def get_metrics_content_type():
    return CONTENT_TYPE_LATEST
