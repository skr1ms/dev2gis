import logging
from typing import Optional
import boto3
from botocore.exceptions import ClientError


logger = logging.getLogger(__name__)


class MinioClient:
    def __init__(self, endpoint: str, access_key: str, secret_key: str, use_ssl: bool = False):
        self.endpoint = endpoint
        self.client = boto3.client(
            's3',
            endpoint_url=f"{'https' if use_ssl else 'http'}://{endpoint}",
            aws_access_key_id=access_key,
            aws_secret_access_key=secret_key,
            region_name='us-east-1'
        )
        logger.info(f"S3 client initialized for endpoint: {endpoint}")

    def download_file(self, bucket: str, object_name: str, local_path: str) -> bool:
        try:
            logger.info(f"Downloading {object_name} from bucket {bucket}")
            self.client.download_file(bucket, object_name, local_path)
            logger.info(f"Successfully downloaded to {local_path}")
            return True
        except ClientError as e:
            logger.error(f"Failed to download file: {e}")
            return False

    def upload_file(self, file_path: str, bucket: str, object_name: str) -> Optional[str]:
        try:
            logger.info(
                f"Uploading {file_path} to bucket {bucket} as {object_name}")

            content_type = 'application/octet-stream'
            if object_name.endswith('.png'):
                content_type = 'image/png'
            elif object_name.endswith('.jpg') or object_name.endswith('.jpeg'):
                content_type = 'image/jpeg'
            elif object_name.endswith('.tif') or object_name.endswith('.tiff'):
                content_type = 'image/tiff'

            self.client.upload_file(
                file_path,
                bucket,
                object_name,
                ExtraArgs={
                    'ContentType': content_type,
                    'ContentDisposition': 'inline'
                }
            )
            logger.info(
                f"Successfully uploaded {object_name} with ContentType={content_type}")
            return object_name
        except ClientError as e:
            logger.error(f"Failed to upload file: {e}")
            return None

    def delete_file(self, bucket: str, object_name: str) -> bool:
        try:
            logger.info(f"Deleting {object_name} from bucket {bucket}")
            self.client.delete_object(Bucket=bucket, Key=object_name)
            logger.info(f"Successfully deleted {object_name}")
            return True
        except ClientError as e:
            logger.error(f"Failed to delete file: {e}")
            return False

    def generate_url(self, bucket: str, object_name: str, endpoint_public: Optional[str] = None) -> str:
        if endpoint_public:
            if endpoint_public.startswith('http://') or endpoint_public.startswith('https://'):
                return f"{endpoint_public}/{bucket}/{object_name}"
            return f"http://{endpoint_public}/{bucket}/{object_name}"
        return f"http://{self.endpoint}/{bucket}/{object_name}"
