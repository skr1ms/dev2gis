import os
from typing import List, Tuple


SUPPORTED_IMAGE_FORMATS = {'.jpg', '.jpeg', '.png', '.tif', '.tiff', '.bmp'}


def is_valid_image_file(file_path: str) -> Tuple[bool, str]:
    if not os.path.exists(file_path):
        return False, "File does not exist"

    if not os.path.isfile(file_path):
        return False, "Path is not a file"

    ext = os.path.splitext(file_path)[1].lower()
    if ext not in SUPPORTED_IMAGE_FORMATS:
        return False, f"Unsupported format: {ext}"

    file_size = os.path.getsize(file_path)
    if file_size == 0:
        return False, "File is empty"

    if file_size > 500 * 1024 * 1024:
        return False, f"File too large: {file_size / (1024*1024):.2f}MB"

    return True, "OK"


def validate_image_dimensions(width: int, height: int, min_size: int = 100, max_size: int = 8000) -> Tuple[bool, str]:
    if width < min_size or height < min_size:
        return False, f"Image too small: {width}x{height}"

    if width > max_size or height > max_size:
        return False, f"Image too large: {width}x{height}"

    return True, "OK"


def validate_batch_images(image_paths: List[str]) -> Tuple[bool, str]:
    if not image_paths:
        return False, "No images provided"

    if len(image_paths) < 3:
        return False, "Batch requires at least 3 images"

    if len(image_paths) > 1000:
        return False, "Too many images in batch"

    for img_path in image_paths:
        valid, msg = is_valid_image_file(img_path)
        if not valid:
            return False, f"{os.path.basename(img_path)}: {msg}"

    return True, "OK"
