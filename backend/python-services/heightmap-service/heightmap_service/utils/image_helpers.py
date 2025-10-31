import cv2
import numpy as np
from typing import Tuple, Optional


def normalize_array(arr: np.ndarray, min_val: int = 0, max_val: int = 255) -> np.ndarray:
    return cv2.normalize(arr, None, min_val, max_val, cv2.NORM_MINMAX).astype(np.uint8)


def get_image_dimensions(image_path: str) -> Optional[Tuple[int, int]]:
    try:
        img = cv2.imread(image_path)
        if img is None:
            return None
        height, width = img.shape[:2]
        return width, height
    except Exception:
        return None


def convert_to_grayscale(image: np.ndarray) -> np.ndarray:
    if len(image.shape) == 2:
        return image
    return cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)


def resize_to_dimensions(image: np.ndarray, width: int, height: int) -> np.ndarray:
    return cv2.resize(image, (width, height), interpolation=cv2.INTER_LINEAR)


def calculate_aspect_ratio(width: int, height: int) -> float:
    return width / height if height > 0 else 1.0
