from heightmap_service.core.logging import logger
from heightmap_service.core.metrics_server import start_metrics_server
from heightmap_service.workers.single_worker import HeightmapWorker
from heightmap_service.workers.batch_worker import BatchHeightmapWorker
from heightmap_service.config import config
import sys
import signal
import threading


def serve():
    logger.info("Starting Python Heightmap Processing Service")

    metrics_server = start_metrics_server(port=8000)

    single_worker = HeightmapWorker()
    batch_worker = BatchHeightmapWorker()

    def signal_handler(sig, frame):
        logger.info("Received shutdown signal, stopping workers...")
        single_worker.should_stop = True
        batch_worker.should_stop = True
        metrics_server.shutdown()
        sys.exit(0)

    signal.signal(signal.SIGTERM, signal_handler)
    signal.signal(signal.SIGINT, signal_handler)

    logger.info("Workers configuration:", extra={
        'rabbitmq_url': config.rabbitmq_url,
        'single_queue': config.rabbitmq_queue_name,
        'batch_queue': f"{config.rabbitmq_queue_name}_batch",
        'metrics_port': 8000,
        'minio_endpoint': config.minio_endpoint,
        'temp_directory': config.temp_dir
    })

    single_thread = threading.Thread(
        target=single_worker.start, name="SingleWorker")
    batch_thread = threading.Thread(
        target=batch_worker.start, name="BatchWorker")

    single_thread.start()
    batch_thread.start()

    single_thread.join()
    batch_thread.join()


if __name__ == '__main__':
    serve()
