from PIL import Image
from PIL.ExifTags import TAGS, GPSTAGS
from typing import Optional, Dict, Tuple
import logging

logger = logging.getLogger(__name__)


def get_exif_data(image_path: str) -> Dict:
    try:
        image = Image.open(image_path)
        exif_data = {}

        info = image._getexif()
        if info:
            for tag, value in info.items():
                tag_name = TAGS.get(tag, tag)
                exif_data[tag_name] = value

        return exif_data
    except Exception as e:
        logger.error(f"Error reading EXIF from {image_path}: {e}")
        return {}


def get_gps_data(exif_data: Dict) -> Dict:
    gps_info = {}

    if 'GPSInfo' in exif_data:
        for key in exif_data['GPSInfo'].keys():
            decode = GPSTAGS.get(key, key)
            gps_info[decode] = exif_data['GPSInfo'][key]

    return gps_info


def convert_to_degrees(value: Tuple[float, float, float]) -> float:
    d, m, s = value
    return d + (m / 60.0) + (s / 3600.0)


def get_gps_coordinates(image_path: str) -> Optional[Dict[str, float]]:
    exif = get_exif_data(image_path)
    if not exif:
        return None

    gps = get_gps_data(exif)
    if not gps:
        return None

    if 'GPSLatitude' not in gps or 'GPSLongitude' not in gps:
        return None

    try:
        lat = convert_to_degrees(gps['GPSLatitude'])
        if gps.get('GPSLatitudeRef') == 'S':
            lat = -lat

        lon = convert_to_degrees(gps['GPSLongitude'])
        if gps.get('GPSLongitudeRef') == 'W':
            lon = -lon

        alt = gps.get('GPSAltitude', 0)
        if isinstance(alt, tuple):
            alt = alt[0] / alt[1] if alt[1] != 0 else 0

        return {
            'latitude': lat,
            'longitude': lon,
            'altitude': float(alt),
        }
    except Exception as e:
        logger.error(f"Error parsing GPS coordinates: {e}")
        return None


def check_images_have_gps(image_paths: list) -> Tuple[int, int]:
    total = len(image_paths)
    with_gps = 0

    for path in image_paths:
        coords = get_gps_coordinates(path)
        if coords:
            with_gps += 1

    return with_gps, total


def get_camera_info(image_path: str) -> Optional[Dict[str, str]]:
    exif = get_exif_data(image_path)
    if not exif:
        return None

    return {
        'make': exif.get('Make', 'Unknown'),
        'model': exif.get('Model', 'Unknown'),
        'datetime': exif.get('DateTime', 'Unknown'),
    }
