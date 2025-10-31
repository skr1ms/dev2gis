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
from osgeo import gdal
from heightmap_service.config import config
from heightmap_service.infrastructure.minio_client import MinioClient
from heightmap_service.services.image_service import ImageProcessor
from heightmap_service.services.heightmap_service import HeightmapGenerator
from heightmap_service.services.batch_service import NodeODMClient
from heightmap_service.core.metrics import (
    record_file_size,
    record_image_dimension,
    heightmap_errors_total
)


logger = logging.getLogger(__name__)


class BatchHeightmapWorker:
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
        logger.info(
            f"Received signal {signum}, stopping batch worker gracefully...")
        self.should_stop = True
        if self.channel:
            self.channel.stop_consuming()

    def connect_rabbitmq(self):
        queue_name = self.config.rabbitmq_queue_name + "_batch"
        logger.info(f"Connecting to RabbitMQ: {self.config.rabbitmq_url}")
        parameters = pika.URLParameters(self.config.rabbitmq_url)
        parameters.heartbeat = 600
        parameters.blocked_connection_timeout = 300

        self.connection = pika.BlockingConnection(parameters)
        self.channel = self.connection.channel()

        self.channel.queue_declare(
            queue=queue_name,
            durable=True
        )

        self.channel.basic_qos(prefetch_count=1)

        logger.info(
            f"Successfully connected to RabbitMQ (batch queue: {queue_name})")

    def connect_database(self):
        try:
            conn = psycopg2.connect(self.config.database_url)
            return conn
        except Exception as e:
            logger.error(f"Failed to connect to database: {e}")
            raise

    def update_batch_job_status(self, batch_job_id: str, status: str, error_message: Optional[str] = None):
        try:
            conn = self.connect_database()
            cur = conn.cursor()

            if error_message:
                cur.execute(
                    "UPDATE batch_heightmap_jobs SET status = %s, error_message = %s, updated_at = CURRENT_TIMESTAMP WHERE id = %s",
                    (status, error_message, batch_job_id)
                )
            else:
                cur.execute(
                    "UPDATE batch_heightmap_jobs SET status = %s, updated_at = CURRENT_TIMESTAMP WHERE id = %s",
                    (status, batch_job_id)
                )

            conn.commit()
            cur.close()
            conn.close()
            logger.info(f"Updated batch job {batch_job_id} status to {status}")
        except Exception as e:
            logger.error(f"Failed to update batch job status: {e}")

    def update_batch_job_result(self, batch_job_id: str, result_url: str, width: int, height: int, processing_time: float, orthophoto_url: str = None):
        try:
            conn = self.connect_database()
            cur = conn.cursor()

            if orthophoto_url:
                cur.execute(
                    """UPDATE batch_heightmap_jobs 
                       SET status = %s, result_url = %s, orthophoto_url = %s, width = %s, height = %s, 
                           processing_time = %s, updated_at = CURRENT_TIMESTAMP 
                       WHERE id = %s""",
                    ('completed', result_url, orthophoto_url, width,
                     height, processing_time, batch_job_id)
                )
            else:
                cur.execute(
                    """UPDATE batch_heightmap_jobs 
                       SET status = %s, result_url = %s, width = %s, height = %s, 
                           processing_time = %s, updated_at = CURRENT_TIMESTAMP 
                       WHERE id = %s""",
                    ('completed', result_url, width,
                     height, processing_time, batch_job_id)
                )

            conn.commit()
            cur.close()
            conn.close()
            logger.info(f"Updated batch job {batch_job_id} with results")
        except Exception as e:
            logger.error(f"Failed to update batch job result: {e}")

    def process_batch_task(self, task: dict):
        batch_job_id = task['batch_job_id']
        logger.info(
            f"Processing batch task: batch_job_id={batch_job_id}, images={len(task['image_urls'])}")

        start_time = time.time()
        input_paths = []
        output_path = None
        dem_output = None

        try:
            self.update_batch_job_status(batch_job_id, 'processing')

            s3_client = MinioClient(
                endpoint=task['minio_endpoint'],
                access_key=task['minio_access_key'],
                secret_key=task['minio_secret_key'],
                use_ssl=task['minio_use_ssl']
            )

            source_objects = []
            for idx, image_url in enumerate(task['image_urls']):
                logger.info(
                    f"Downloading image {idx + 1}/{len(task['image_urls'])}")

                image_url_parts = image_url.split('/')
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

                input_path = os.path.join(
                    self.config.temp_dir, f"{batch_job_id}_input_{idx}.jpg")
                input_paths.append(input_path)
                source_objects.append((bucket_name, object_name))

                if not s3_client.download_file(bucket_name, object_name, input_path):
                    raise Exception(
                        f"Failed to download image {idx + 1} from MinIO")

            logger.info(
                f"All {len(input_paths)} images downloaded")

            fast_mode = task.get('fast_mode', False)
            merge_method = task.get('merge_method', 'average')
            generation_mode = task.get('generation_mode', 'heightmap')

            if fast_mode:
                from PIL import Image
                logger.info(
                    "Fast mode enabled: Resizing images to max 2000px for faster processing")

                resized_count = 0
                for idx, path in enumerate(input_paths):
                    try:
                        with Image.open(path) as img:
                            original_size = img.size
                            max_dimension = max(img.size)

                            if max_dimension > 2000:
                                ratio = 2000 / max_dimension
                                new_size = tuple(int(dim * ratio)
                                                 for dim in img.size)

                                exif = img.info.get('exif')

                                img_resized = img.resize(
                                    new_size, Image.Resampling.LANCZOS)

                                save_kwargs = {'quality': 85, 'optimize': True}
                                if exif:
                                    save_kwargs['exif'] = exif

                                img_resized.save(path, **save_kwargs)
                                resized_count += 1

                                logger.info(
                                    f"Image {idx + 1}/{len(input_paths)}: {original_size} → {new_size} "
                                    f"({int((1 - (new_size[0]*new_size[1])/(original_size[0]*original_size[1]))*100)}% size reduction)")
                            else:
                                logger.info(
                                    f"Image {idx + 1}/{len(input_paths)}: {original_size} (no resize needed)")

                    except Exception as e:
                        logger.warning(
                            f"Failed to resize image {idx + 1}: {e}. Using original size.")

                logger.info(
                    f"Image resizing completed: {resized_count}/{len(input_paths)} images resized")

            logger.info("Checking GPS data in images")
            from heightmap_service.utils.exif_helper import check_images_have_gps
            with_gps, total = check_images_have_gps(input_paths)

            logger.info(
                f"GPS check: {with_gps}/{total} images have GPS coordinates")

            if with_gps == 0:
                logger.warning(
                    "No GPS data found in images! NodeODM will work but results won't be georeferenced")
            elif with_gps < total:
                logger.warning(
                    f"Only {with_gps}/{total} images have GPS. Missing GPS may reduce accuracy")
            else:
                logger.info(
                    "All images have GPS coordinates - NodeODM will produce georeferenced results!")

            logger.info("Starting NodeODM processing")
            logger.info(
                f"Using quality preset: {merge_method}, fast_mode: {fast_mode}, generation_mode: {generation_mode}")

            odm_client = NodeODMClient(base_url=self.config.nodeodm_url)

            dem_output = os.path.join(
                self.config.temp_dir, f"{batch_job_id}_dem.tif")

            success, task_uuid = odm_client.process_batch(
                input_paths, dem_output, quality_preset=merge_method, fast_mode=fast_mode)

            if not success:
                raise Exception("Failed to process batch with NodeODM")

            if task_uuid is None:
                raise Exception(
                    "NodeODM task_uuid is None - cannot extract assets")

            result_url = None
            width = None
            height = None

            if generation_mode in ['heightmap', 'both']:
                logger.info("Converting DEM to heightmap PNG")

                dataset = gdal.Open(dem_output)
                if dataset is None:
                    raise Exception("Failed to open DEM file")

                band = dataset.GetRasterBand(1)
                dem_array = band.ReadAsArray().astype(np.float32)

                nodata_value = band.GetNoDataValue()
                logger.info(
                    f"DEM shape: {dem_array.shape}, NoData value: {nodata_value}")
                logger.info(
                    f"DEM min: {np.nanmin(dem_array)}, max: {np.nanmax(dem_array)}")

                valid_mask = np.isfinite(dem_array)
                if nodata_value is not None:
                    valid_mask &= (dem_array != nodata_value)

                if not valid_mask.any():
                    raise Exception("DEM contains no valid data")

                valid_data = dem_array[valid_mask]
                dem_min = np.min(valid_data)
                dem_max = np.max(valid_data)
                dem_range = dem_max - dem_min

                logger.info(
                    f"DEM valid data - min: {dem_min:.2f}m, max: {dem_max:.2f}m, range: {dem_range:.2f}m")

                dem_cleaned = dem_array.copy()
                dem_cleaned[~valid_mask] = dem_min

                p1, p99 = np.percentile(valid_data, [1, 99])
                logger.info(
                    f"DEM percentiles - 1%: {p1:.2f}m, 99%: {p99:.2f}m")

                dem_clipped = np.clip(dem_cleaned, p1, p99)
                dem_normalized = ((dem_clipped - p1) / (p99 - p1)
                                  * 255).astype(np.uint8)

                logger.info("Applying CLAHE for detail enhancement")
                clahe = cv2.createCLAHE(clipLimit=3.0, tileGridSize=(16, 16))
                dem_enhanced = clahe.apply(dem_normalized)

                logger.info("Applying edge enhancement")
                edges = cv2.Laplacian(dem_enhanced, cv2.CV_64F, ksize=3)
                edges = np.abs(edges)
                edges_norm = ((edges - edges.min()) / (edges.max() -
                              edges.min() + 1e-6) * 255).astype(np.uint8)

                logger.info("Generating multi-angle hillshade")
                dx = cv2.Sobel(dem_enhanced, cv2.CV_64F, 1, 0, ksize=5)
                dy = cv2.Sobel(dem_enhanced, cv2.CV_64F, 0, 1, ksize=5)

                slope = np.sqrt(dx*dx + dy*dy)

                azimuth1 = np.deg2rad(315)
                altitude1 = np.deg2rad(45)
                aspect = np.arctan2(-dy, dx)
                hillshade1 = np.sin(altitude1) * np.cos(slope) + \
                    np.cos(altitude1) * np.sin(slope) * \
                    np.cos(azimuth1 - aspect)

                azimuth2 = np.deg2rad(135)
                altitude2 = np.deg2rad(30)
                hillshade2 = np.sin(altitude2) * np.cos(slope) + \
                    np.cos(altitude2) * np.sin(slope) * \
                    np.cos(azimuth2 - aspect)

                hillshade = hillshade1 * 0.7 + hillshade2 * 0.3
                hillshade = np.clip(hillshade, 0, 1)
                hillshade = (hillshade * 255).astype(np.uint8)

                logger.info("Applying color map (terrain colors)")
                colormap = cv2.applyColorMap(dem_enhanced, cv2.COLORMAP_JET)

                logger.info(
                    "Blending hillshade with color map and edge enhancement")
                hillshade_colored = cv2.cvtColor(hillshade, cv2.COLOR_GRAY2BGR)
                edges_colored = cv2.cvtColor(edges_norm, cv2.COLOR_GRAY2BGR)

                heightmap_rgb = cv2.addWeighted(
                    colormap, 0.5, hillshade_colored, 0.35, 0)
                heightmap_rgb = cv2.addWeighted(
                    heightmap_rgb, 0.95, edges_colored, 0.15, 0)

                heightmap_rgb[~valid_mask] = [0, 0, 0]

                logger.info(
                    "Heightmap RGB created with hillshade visualization")

                width = heightmap_rgb.shape[1]
                height = heightmap_rgb.shape[0]

                record_image_dimension(width, height)

                output_path = os.path.join(
                    self.config.temp_dir, f"{batch_job_id}_batch_heightmap.png")

                logger.info(f"Saving RGB heightmap to {output_path}")
                cv2.imwrite(output_path, heightmap_rgb)

                if not os.path.exists(output_path):
                    raise Exception("Failed to save batch heightmap")

                file_size = os.path.getsize(output_path)
                record_file_size(file_size)

                result_object_name = f"batch-heightmaps/{batch_job_id}_heightmap.png"
                uploaded_name = s3_client.upload_file(
                    output_path, task['output_bucket'], result_object_name)

                if uploaded_name is None:
                    raise Exception(
                        "Failed to upload batch heightmap to MinIO")

                result_url = s3_client.generate_url(
                    task['output_bucket'], result_object_name, task.get('minio_public_url', task.get('minio_endpoint')))

            orthophoto_url = None
            if generation_mode in ['orthophoto', 'both']:
                logger.info("Extracting orthophoto from NodeODM results")
                orthophoto_output = os.path.join(
                    self.config.temp_dir, f"{batch_job_id}_orthophoto.tif")

                if odm_client.extract_orthophoto(task_uuid, orthophoto_output):
                    logger.info("Uploading orthophoto to MinIO")
                    orthophoto_object_name = f"orthophotos/{batch_job_id}_orthophoto.tif"

                    uploaded_orthophoto = s3_client.upload_file(
                        orthophoto_output, 'uav-photoplanes', orthophoto_object_name)

                    if uploaded_orthophoto:
                        orthophoto_url = s3_client.generate_url(
                            'uav-photoplanes', orthophoto_object_name, task.get('minio_public_url', task.get('minio_endpoint')))
                        logger.info(f"Orthophoto uploaded: {orthophoto_url}")
                    else:
                        logger.warning("Failed to upload orthophoto to MinIO")

                    if os.path.exists(orthophoto_output):
                        try:
                            os.remove(orthophoto_output)
                        except Exception as e:
                            logger.warning(
                                f"Failed to remove orthophoto file: {e}")
                else:
                    logger.warning(
                        "Failed to extract orthophoto from NodeODM results")

            processing_time = time.time() - start_time

            self.update_batch_job_result(
                batch_job_id, result_url, width, height, processing_time, orthophoto_url)

            deleted_count = 0
            for bucket_name, object_name in source_objects:
                try:
                    logger.info(
                        f"Deleting source image: bucket={bucket_name}, object={object_name}")
                    s3_client.delete_file(bucket_name, object_name)
                    deleted_count += 1
                except Exception as e:
                    logger.warning(
                        f"Failed to delete source image {object_name}: {e}")
            logger.info(
                f"Deleted {deleted_count}/{len(source_objects)} source images")

            logger.info(
                f"Batch heightmap generation completed: batch_job_id={batch_job_id}, time={processing_time:.2f}s, images={len(input_paths)}")

        except Exception as e:
            logger.error(
                f"Error processing batch task for job {batch_job_id}: {e}")
            logger.error(traceback.format_exc())
            heightmap_errors_total.labels(error_type=type(e).__name__).inc()
            self.update_batch_job_status(batch_job_id, 'failed', str(e))
            raise

        finally:
            for input_path in input_paths:
                if os.path.exists(input_path):
                    try:
                        os.remove(input_path)
                    except Exception as e:
                        logger.warning(f"Failed to remove input file: {e}")

            if output_path and os.path.exists(output_path):
                try:
                    os.remove(output_path)
                except Exception as e:
                    logger.warning(f"Failed to remove output file: {e}")

            if dem_output and os.path.exists(dem_output):
                try:
                    os.remove(dem_output)
                except Exception as e:
                    logger.warning(f"Failed to remove DEM file: {e}")

    def callback(self, ch, method, properties, body):
        try:
            task = json.loads(body)
            logger.info(f"Received batch task: {task['batch_job_id']}")

            self.process_batch_task(task)

            ch.basic_ack(delivery_tag=method.delivery_tag)
            logger.info(f"Batch task {task['batch_job_id']} acknowledged")

        except Exception as e:
            logger.error(f"Error in batch callback: {e}")
            logger.error(traceback.format_exc())
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)

    def start(self):
        retry_delay = 5
        max_retry_delay = 60
        queue_name = self.config.rabbitmq_queue_name + "_batch"

        while not self.should_stop:
            try:
                self.connect_rabbitmq()

                logger.info(
                    f"Starting to consume from batch queue: {queue_name}")
                self.channel.basic_consume(
                    queue=queue_name,
                    on_message_callback=self.callback
                )

                logger.info("Batch worker started, waiting for batch tasks...")
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

        logger.info("Batch worker stopped")


def main():
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )

    worker = BatchHeightmapWorker()
    worker.start()


if __name__ == "__main__":
    main()
