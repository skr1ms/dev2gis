import json
import logging
import os
import time
import traceback
import signal
import cv2
import numpy as np
import pika
import psycopg2
from typing import Optional
from heightmap_service.config import config
from heightmap_service.infrastructure.minio_client import MinioClient
from heightmap_service.services.image_service import ImageProcessor
from heightmap_service.services.heightmap_service import HeightmapGenerator
from heightmap_service.core.metrics import (
    record_file_size,
    record_image_dimension,
    heightmap_errors_total
)


logger = logging.getLogger(__name__)


class HeightmapWorker:
    def __init__(self):
        self.config = config
        self.image_processor = ImageProcessor()
        self.heightmap_generator = HeightmapGenerator()
        self.connection: Optional[pika.BlockingConnection] = None
        self.channel: Optional[pika.channel.Channel] = None
        self.should_stop = False

        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)

    def _signal_handler(self, signum, frame):
        logger.info(f"Received signal {signum}, stopping worker gracefully...")
        self.should_stop = True
        if self.channel:
            self.channel.stop_consuming()

    def connect_rabbitmq(self):
        logger.info(f"Connecting to RabbitMQ: {self.config.rabbitmq_url}")
        parameters = pika.URLParameters(self.config.rabbitmq_url)
        parameters.heartbeat = 600
        parameters.blocked_connection_timeout = 300

        self.connection = pika.BlockingConnection(parameters)
        self.channel = self.connection.channel()

        self.channel.queue_declare(
            queue=self.config.rabbitmq_queue_name,
            durable=True
        )

        self.channel.basic_qos(
            prefetch_count=self.config.rabbitmq_prefetch_count)

        logger.info("Successfully connected to RabbitMQ")

    def connect_database(self):
        try:
            conn = psycopg2.connect(self.config.database_url)
            return conn
        except Exception as e:
            logger.error(f"Failed to connect to database: {e}")
            raise

    def update_job_status(self, job_id: str, status: str, error_message: Optional[str] = None):
        try:
            conn = self.connect_database()
            cur = conn.cursor()

            if error_message:
                cur.execute(
                    "UPDATE heightmap_jobs SET status = %s, error_message = %s, updated_at = CURRENT_TIMESTAMP WHERE id = %s",
                    (status, error_message, job_id)
                )
            else:
                cur.execute(
                    "UPDATE heightmap_jobs SET status = %s, updated_at = CURRENT_TIMESTAMP WHERE id = %s",
                    (status, job_id)
                )

            conn.commit()
            cur.close()
            conn.close()
            logger.info(f"Updated job {job_id} status to {status}")
        except Exception as e:
            logger.error(f"Failed to update job status: {e}")

    def update_job_result(self, job_id: str, result_url: str, width: int, height: int, processing_time: float):
        try:
            conn = self.connect_database()
            cur = conn.cursor()

            cur.execute(
                """UPDATE heightmap_jobs 
                   SET status = %s, result_url = %s, width = %s, height = %s, 
                       processing_time = %s, updated_at = CURRENT_TIMESTAMP 
                   WHERE id = %s""",
                ('completed', result_url, width, height, processing_time, job_id)
            )

            conn.commit()
            cur.close()
            conn.close()
            logger.info(f"Updated job {job_id} with results")
        except Exception as e:
            logger.error(f"Failed to update job result: {e}")

    def process_task(self, task: dict):
        job_id = task['job_id']
        logger.info(f"Processing task: job_id={job_id}")

        start_time = time.time()
        input_path = None
        output_path = None

        try:
            self.update_job_status(job_id, 'processing')

            s3_client = MinioClient(
                endpoint=task['minio_endpoint'],
                access_key=task['minio_access_key'],
                secret_key=task['minio_secret_key'],
                use_ssl=task['minio_use_ssl']
            )

            image_url_parts = task['image_url'].split('/')
            bucket_index = -1
            for i, part in enumerate(image_url_parts):
                if part in ['uav-data', 'uav-models']:
                    bucket_index = i
                    break

            if bucket_index != -1:
                bucket_name = image_url_parts[bucket_index]
                object_name = '/'.join(image_url_parts[bucket_index + 1:])
            else:
                bucket_name = task['output_bucket']
                object_name = image_url_parts[-1]

            logger.info(
                f"Downloading image: bucket={bucket_name}, object={object_name}")

            input_path = os.path.join(
                self.config.temp_dir, f"{job_id}_input.jpg")

            if not s3_client.download_file(bucket_name, object_name, input_path):
                raise Exception("Failed to download image from MinIO")

            image = self.image_processor.load_image(input_path)
            if image is None:
                raise Exception("Failed to load image")

            valid, msg = self.image_processor.validate_image(image)
            if not valid:
                raise Exception(f"Image validation failed: {msg}")

            image = self.image_processor.resize_if_needed(image, max_size=4000)

            heightmap, width, height = self.heightmap_generator.generate(image)
            if heightmap is None:
                raise Exception("Failed to generate heightmap")

            logger.info(
                f"Raw heightmap stats - min: {heightmap.min()}, max: {heightmap.max()}, mean: {heightmap.mean():.2f}")

            clahe = cv2.createCLAHE(clipLimit=2.5, tileGridSize=(8, 8))
            heightmap_enhanced = clahe.apply(heightmap)

            logger.info(
                f"Enhanced heightmap stats - min: {heightmap_enhanced.min()}, max: {heightmap_enhanced.max()}, mean: {heightmap_enhanced.mean():.2f}")

            logger.info("Generating hillshade for better visualization")
            dx = cv2.Sobel(heightmap_enhanced, cv2.CV_64F, 1, 0, ksize=5)
            dy = cv2.Sobel(heightmap_enhanced, cv2.CV_64F, 0, 1, ksize=5)

            slope = np.sqrt(dx*dx + dy*dy)
            azimuth = np.deg2rad(315)
            altitude = np.deg2rad(45)

            aspect = np.arctan2(-dy, dx)

            hillshade = np.sin(altitude) * np.cos(slope) + \
                np.cos(altitude) * np.sin(slope) * np.cos(azimuth - aspect)

            hillshade = np.clip(hillshade, 0, 1)
            hillshade = (hillshade * 255).astype(np.uint8)

            logger.info("Applying color map (terrain colors)")
            colormap = cv2.applyColorMap(heightmap_enhanced, cv2.COLORMAP_JET)

            logger.info("Blending hillshade with color map")
            hillshade_colored = cv2.cvtColor(hillshade, cv2.COLOR_GRAY2BGR)
            heightmap_rgb = cv2.addWeighted(
                colormap, 0.6, hillshade_colored, 0.4, 0)

            logger.info("Heightmap RGB created with hillshade visualization")

            record_image_dimension(width, height)

            output_path = os.path.join(
                self.config.temp_dir, f"{job_id}_heightmap.png")

            logger.info(f"Saving RGB heightmap to {output_path}")
            cv2.imwrite(output_path, heightmap_rgb)

            if not os.path.exists(output_path):
                raise Exception("Failed to save heightmap")

            file_size = os.path.getsize(output_path)
            record_file_size(file_size)

            result_object_name = f"heightmaps/{job_id}_heightmap.png"
            uploaded_name = s3_client.upload_file(
                output_path, task['output_bucket'], result_object_name)

            if uploaded_name is None:
                raise Exception("Failed to upload heightmap to MinIO")

            minio_public = task.get(
                'minio_public_url', task.get('minio_endpoint'))
            logger.info(
                f"Generating URL with minio_public_url: {minio_public}")

            result_url = s3_client.generate_url(
                task['output_bucket'], result_object_name, minio_public)

            logger.info(f"Generated result_url: {result_url}")
            processing_time = time.time() - start_time

            self.update_job_result(
                job_id, result_url, width, height, processing_time)

            try:
                logger.info(
                    f"Deleting source image: bucket={bucket_name}, object={object_name}")
                s3_client.delete_file(bucket_name, object_name)
                logger.info("Source image deleted successfully")
            except Exception as e:
                logger.warning(f"Failed to delete source image: {e}")

            logger.info(
                f"Heightmap generation completed: job_id={job_id}, time={processing_time:.2f}s")

        except Exception as e:
            logger.error(f"Error processing task for job {job_id}: {e}")
            logger.error(traceback.format_exc())
            heightmap_errors_total.labels(error_type=type(e).__name__).inc()
            self.update_job_status(job_id, 'failed', str(e))
            raise

        finally:
            if input_path and os.path.exists(input_path):
                try:
                    os.remove(input_path)
                except Exception as e:
                    logger.warning(f"Failed to remove input file: {e}")

            if output_path and os.path.exists(output_path):
                try:
                    os.remove(output_path)
                except Exception as e:
                    logger.warning(f"Failed to remove output file: {e}")

    def callback(self, ch, method, properties, body):
        try:
            task = json.loads(body)
            logger.info(f"Received task: {task['job_id']}")

            self.process_task(task)

            ch.basic_ack(delivery_tag=method.delivery_tag)
            logger.info(f"Task {task['job_id']} acknowledged")

        except Exception as e:
            logger.error(f"Error in callback: {e}")
            logger.error(traceback.format_exc())
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)

    def start(self):
        retry_delay = 5
        max_retry_delay = 60

        while not self.should_stop:
            try:
                self.connect_rabbitmq()

                logger.info(
                    f"Starting to consume from queue: {self.config.rabbitmq_queue_name}")
                self.channel.basic_consume(
                    queue=self.config.rabbitmq_queue_name,
                    on_message_callback=self.callback
                )

                logger.info("Worker started, waiting for tasks...")
                self.channel.start_consuming()

            except KeyboardInterrupt:
                logger.info("Interrupted by user, shutting down...")
                self.should_stop = True
                break

            except pika.exceptions.AMQPConnectionError as e:
                logger.error(f"RabbitMQ connection error: {e}")
                if not self.should_stop:
                    logger.info(f"Reconnecting in {retry_delay} seconds...")
                    time.sleep(retry_delay)
                    retry_delay = min(retry_delay * 2, max_retry_delay)

            except Exception as e:
                logger.error(f"Unexpected error: {e}")
                logger.error(traceback.format_exc())
                if not self.should_stop:
                    time.sleep(retry_delay)

            finally:
                if self.channel and self.channel.is_open:
                    self.channel.close()
                if self.connection and self.connection.is_open:
                    self.connection.close()

        logger.info("Worker stopped")
