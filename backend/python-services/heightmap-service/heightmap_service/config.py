import os


class Config:
    def __init__(self):
        self.rabbitmq_url = os.getenv(
            "RABBITMQ_URL", "amqp://admin:admin@rabbitmq:5672/")
        self.rabbitmq_queue_name = os.getenv(
            "RABBITMQ_QUEUE_NAME", "heightmap.tasks")
        self.rabbitmq_prefetch_count = int(
            os.getenv("RABBITMQ_PREFETCH_COUNT", "1"))

        self.database_url = os.getenv("DATABASE_URL", "")

        self.minio_endpoint = os.getenv("MINIO_ENDPOINT", "localhost:9000")
        self.minio_access_key = os.getenv("MINIO_ACCESS_KEY", "minioadmin")
        self.minio_secret_key = os.getenv("MINIO_SECRET_KEY", "minioadmin")
        self.minio_use_ssl = os.getenv(
            "MINIO_USE_SSL", "false").lower() == "true"
        self.minio_input_bucket = os.getenv("MINIO_INPUT_BUCKET", "uav-data")
        self.minio_output_bucket = os.getenv(
            "MINIO_OUTPUT_BUCKET", "uav-models")

        self.nodeodm_url = os.getenv("NODEODM_URL", "http://nodeodm:3000")
        self.temp_dir = os.getenv("TEMP_DIR", "/tmp/heightmap_processing")
        self.log_level = os.getenv("LOG_LEVEL", "INFO")

        self._validate()
        self._ensure_temp_dir()

    def _validate(self):
        if not self.rabbitmq_url:
            raise ValueError("RABBITMQ_URL is required")
        if not self.database_url:
            raise ValueError("DATABASE_URL is required")
        if not self.minio_endpoint:
            raise ValueError("MINIO_ENDPOINT is required")
        if not self.minio_access_key:
            raise ValueError("MINIO_ACCESS_KEY is required")
        if not self.minio_secret_key:
            raise ValueError("MINIO_SECRET_KEY is required")

    def _ensure_temp_dir(self):
        os.makedirs(self.temp_dir, exist_ok=True)


config = Config()
