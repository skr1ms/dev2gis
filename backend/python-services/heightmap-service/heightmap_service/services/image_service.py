import logging
import cv2
import numpy as np
from typing import Optional, Tuple


logger = logging.getLogger(__name__)


class ImageProcessor:
    def __init__(self, max_size: int = 8000):
        self.max_size = max_size

    def load_image(self, file_path: str) -> Optional[np.ndarray]:
        try:
            logger.info(f"Loading image from {file_path}")
            image = cv2.imread(file_path)
            if image is None:
                logger.error(f"Failed to load image: {file_path}")
                return None
            logger.info(f"Image loaded successfully: shape={image.shape}")
            return image
        except Exception as e:
            logger.error(f"Error loading image: {e}")
            return None

    def validate_image(self, image: np.ndarray) -> Tuple[bool, str]:
        if image is None:
            return False, "Image is None"

        if len(image.shape) < 2:
            return False, "Invalid image dimensions"

        height, width = image.shape[:2]

        if height < 100 or width < 100:
            return False, f"Image too small: {width}x{height}"

        if height > self.max_size or width > self.max_size:
            return False, f"Image too large: {width}x{height}"

        logger.info(f"Image validation passed: {width}x{height}")
        return True, "OK"

    def resize_if_needed(self, image: np.ndarray, max_size: Optional[int] = None) -> np.ndarray:
        if max_size is None:
            max_size = self.max_size

        height, width = image.shape[:2]

        if height <= max_size and width <= max_size:
            return image

        scale = min(max_size / width, max_size / height)
        new_width = int(width * scale)
        new_height = int(height * scale)

        logger.info(
            f"Resizing image from {width}x{height} to {new_width}x{new_height}")
        resized = cv2.resize(image, (new_width, new_height),
                             interpolation=cv2.INTER_AREA)

        return resized
