import logging
import time
import requests
import os
import json
import zipfile
import tempfile
from typing import List, Optional, Dict

logger = logging.getLogger(__name__)


class NodeODMClient:
    def __init__(self, base_url: str = "http://nodeodm:3000"):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()

    def get_quality_options(self, quality_preset: str = 'max', fast_mode: bool = False) -> List[Dict]:
        if quality_preset == 'max':
            options = [
                {'name': 'dsm', 'value': True},
                {'name': 'dtm', 'value': False},
                {'name': 'pc-quality', 'value': 'ultra'},
                {'name': 'feature-quality', 'value': 'ultra'},
                {'name': 'dem-resolution', 'value': 0.5},
                {'name': 'orthophoto-resolution', 'value': 0.5},
                {'name': 'min-num-features', 'value': 20000},
                {'name': 'mesh-octree-depth', 'value': 12},
                {'name': 'mesh-size', 'value': 500000}
            ]
        elif quality_preset == 'medium':
            options = [
                {'name': 'dsm', 'value': True},
                {'name': 'dtm', 'value': False},
                {'name': 'pc-quality', 'value': 'high'},
                {'name': 'feature-quality', 'value': 'high'},
                {'name': 'dem-resolution', 'value': 1.0},
                {'name': 'orthophoto-resolution', 'value': 1.0},
                {'name': 'min-num-features', 'value': 10000},
                {'name': 'mesh-octree-depth', 'value': 12},
                {'name': 'mesh-size', 'value': 200000}
            ]
        else:
            options = [
                {'name': 'dsm', 'value': True},
                {'name': 'dtm', 'value': False},
                {'name': 'pc-quality', 'value': 'high'},
                {'name': 'feature-quality', 'value': 'high'},
                {'name': 'dem-resolution', 'value': 1.0},
                {'name': 'orthophoto-resolution', 'value': 1.0},
                {'name': 'min-num-features', 'value': 5000},
                {'name': 'mesh-octree-depth', 'value': 12},
                {'name': 'mesh-size', 'value': 100000}
            ]

        if fast_mode:
            logger.info("Fast mode enabled - adding speed optimization flags")
            options.extend([
                # Skip dense reconstruction
                {'name': 'fast-orthophoto', 'value': True},
                # Skip 3D model generation
                {'name': 'skip-3dmodel', 'value': True},
                # Skip PDF report
                {'name': 'skip-report', 'value': True},
                # Delete intermediate files
                {'name': 'optimize-disk-space', 'value': True},
                # Skip geometric estimates
                {'name': 'pc-skip-geometric', 'value': True}
            ])
            # Note: Image resizing to max 2000px is handled in batch_worker.py
            # before NodeODM processing to reduce processing time by ~50-75%

        return options

    def create_task(self, image_paths: List[str], options: Optional[List[Dict]] = None) -> Optional[str]:
        try:
            if not options:
                options = [
                    {'name': 'dsm', 'value': True},
                    {'name': 'dtm', 'value': False},
                    {'name': 'pc-quality', 'value': 'ultra'},
                    {'name': 'feature-quality', 'value': 'ultra'},
                    {'name': 'dem-resolution', 'value': 0.5},
                    {'name': 'orthophoto-resolution', 'value': 0.5},
                    {'name': 'min-num-features', 'value': 20000},
                    {'name': 'mesh-octree-depth', 'value': 12},
                    {'name': 'mesh-size', 'value': 500000}
                ]

            files = []
            for img_path in image_paths:
                if os.path.exists(img_path):
                    files.append(
                        ('images', (os.path.basename(img_path), open(img_path, 'rb'), 'image/jpeg')))

            data = {'options': json.dumps(options)}

            response = self.session.post(
                f"{self.base_url}/task/new",
                files=files,
                data=data,
                timeout=30
            )

            logger.info(f"NodeODM response status: {response.status_code}")

            if response.status_code == 200:
                response_data = response.json()
                logger.info(f"NodeODM response: {response_data}")

                if 'uuid' in response_data:
                    task_uuid = response_data['uuid']
                    logger.info(f"Created NodeODM task: {task_uuid}")
                    return task_uuid
                else:
                    logger.error(f"No 'uuid' in response: {response_data}")
                    return None
            else:
                logger.error(
                    f"Failed to create task: {response.status_code} - {response.text}")
                return None
        except Exception as e:
            logger.error(f"Error creating task: {e}")
            import traceback
            logger.error(traceback.format_exc())
            return None

    def get_task_info(self, task_uuid: str) -> Optional[Dict]:
        try:
            response = self.session.get(
                f"{self.base_url}/task/{task_uuid}/info")
            if response.status_code == 200:
                return response.json()
            return None
        except Exception as e:
            logger.error(f"Error getting task info: {e}")
            return None

    def wait_for_completion(self, task_uuid: str, timeout: int = 3600) -> bool:
        start_time = time.time()
        check_interval = 15

        while time.time() - start_time < timeout:
            info = self.get_task_info(task_uuid)

            if not info:
                time.sleep(check_interval)
                continue

            status = info.get('status', {}).get('code')
            progress = info.get('progress', 0)

            logger.info(
                f"NodeODM Task {task_uuid}: status={status}, progress={progress}%")

            if status == 40:
                logger.info("Task completed successfully")
                return True
            elif status in [30, 50]:
                logger.error(
                    f"Task failed: {info.get('status', {}).get('errorMessage', 'Unknown error')}")
                return False

            time.sleep(check_interval)

        logger.error("Task timed out")
        return False

    def download_asset(self, task_uuid: str, asset_name: str, output_path: str) -> bool:
        try:
            response = self.session.get(
                f"{self.base_url}/task/{task_uuid}/download/{asset_name}",
                stream=True
            )

            if response.status_code == 200:
                with open(output_path, 'wb') as f:
                    for chunk in response.iter_content(chunk_size=8192):
                        f.write(chunk)
                logger.info(f"Downloaded {asset_name} to {output_path}")
                return True
            else:
                logger.error(
                    f"Failed to download {asset_name}: {response.status_code}")
                return False
        except Exception as e:
            logger.error(f"Error downloading asset: {e}")
            return False

    def process_batch(self, image_paths: List[str], output_path: str, quality_preset: str = 'max', fast_mode: bool = False) -> tuple[bool, Optional[str]]:
        logger.info(
            f"Starting NodeODM processing for {len(image_paths)} images with quality: {quality_preset}, fast_mode: {fast_mode}")

        options = self.get_quality_options(quality_preset, fast_mode)
        task_uuid = self.create_task(image_paths, options)
        if not task_uuid:
            return False, None

        if not self.wait_for_completion(task_uuid):
            return False, task_uuid

        with tempfile.NamedTemporaryFile(suffix='.zip', delete=False) as tmp_zip:
            zip_path = tmp_zip.name

        try:
            if not self.download_asset(task_uuid, 'all.zip', zip_path):
                return False

            logger.info(f"Extracting results from {zip_path}")

            with zipfile.ZipFile(zip_path, 'r') as zip_ref:
                for file_info in zip_ref.filelist:
                    if 'dsm.tif' in file_info.filename.lower():
                        logger.info(f"Found DSM file: {file_info.filename}")
                        with zip_ref.open(file_info.filename) as source:
                            with open(output_path, 'wb') as target:
                                target.write(source.read())
                        logger.info(f"Extracted DSM to {output_path}")
                        return True, task_uuid

            logger.error("DSM file not found in results archive")
            return False, task_uuid

        finally:
            if os.path.exists(zip_path):
                os.remove(zip_path)
                logger.info(f"Cleaned up temporary zip: {zip_path}")

    def extract_orthophoto(self, task_uuid: str, output_path: str) -> bool:
        try:
            with tempfile.NamedTemporaryFile(suffix='.zip', delete=False) as tmp_zip:
                zip_path = tmp_zip.name

            logger.info(f"Downloading orthophoto archive for task {task_uuid}")

            if not self.download_asset(task_uuid, 'all.zip', zip_path):
                return False

            logger.info(f"Extracting orthophoto from {zip_path}")

            with zipfile.ZipFile(zip_path, 'r') as zip_ref:
                for file_info in zip_ref.filelist:
                    if 'odm_orthophoto' in file_info.filename.lower() and file_info.filename.lower().endswith('.tif'):
                        logger.info(
                            f"Found orthophoto file: {file_info.filename}")
                        with zip_ref.open(file_info.filename) as source:
                            with open(output_path, 'wb') as target:
                                target.write(source.read())
                        logger.info(f"Extracted orthophoto to {output_path}")
                        return True

            logger.error("Orthophoto file not found in results archive")
            return False

        except Exception as e:
            logger.error(f"Error extracting orthophoto: {e}")
            return False
        finally:
            if os.path.exists(zip_path):
                os.remove(zip_path)
                logger.info(f"Cleaned up orthophoto zip: {zip_path}")
