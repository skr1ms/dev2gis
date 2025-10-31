import logging
import cv2
import numpy as np
from typing import Tuple, Optional
from PIL import Image


logger = logging.getLogger(__name__)


class HeightmapGenerator:
    def __init__(self):
        self.depth_model = None
        self.transform = None
        self.device = None
        self._load_midas_model()

    def _load_midas_model(self):
        try:
            import torch
            logger.info("Loading MiDaS depth estimation model...")

            model_type = "DPT_Large"
            self.depth_model = torch.hub.load("intel-isl/MiDaS", model_type)
            self.depth_model.eval()

            midas_transforms = torch.hub.load("intel-isl/MiDaS", "transforms")
            self.transform = midas_transforms.dpt_transform

            self.device = torch.device(
                "cuda" if torch.cuda.is_available() else "cpu")
            self.depth_model.to(self.device)

            if torch.cuda.is_available():
                self.depth_model = self.depth_model.half()
                logger.info("Using FP16 precision for faster processing")

            logger.info(
                f"MiDaS {model_type} model loaded successfully on {self.device}")
        except Exception as e:
            logger.error(f"Failed to load MiDaS model: {e}")
            raise RuntimeError(f"Cannot initialize heightmap generator: {e}")

    def generate(self, image: np.ndarray) -> Tuple[Optional[np.ndarray], int, int]:
        return self._generate_midas(image)

    def _generate_midas(self, image: np.ndarray) -> Tuple[Optional[np.ndarray], int, int]:
        try:
            import torch
            logger.info("Starting MiDaS depth estimation")

            img_rgb = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)

            input_batch = self.transform(img_rgb).to(self.device)

            with torch.no_grad():
                prediction = self.depth_model(input_batch)

                prediction = torch.nn.functional.interpolate(
                    prediction.unsqueeze(1),
                    size=img_rgb.shape[:2],
                    mode="bicubic",
                    align_corners=False,
                ).squeeze()

            depth_map = prediction.cpu().numpy()

            depth_map = cv2.normalize(depth_map, None, 0, 255, cv2.NORM_MINMAX)
            depth_map = np.uint8(depth_map)

            depth_map = 255 - depth_map

            height, width = depth_map.shape
            logger.info(f"Generated MiDaS heightmap: {width}x{height}")

            return depth_map, width, height

        except Exception as e:
            logger.error(f"Error generating MiDaS heightmap: {e}")
            return None, 0, 0

    def save_heightmap(self, heightmap: np.ndarray, output_path: str) -> bool:
        try:
            logger.info(f"Saving heightmap to {output_path}")

            img = Image.fromarray(heightmap.astype(np.uint8))
            img.save(output_path, format='PNG')

            logger.info(f"Heightmap saved successfully: {output_path}")
            return True

        except Exception as e:
            logger.error(f"Error saving heightmap: {e}")
            return False

    def save_heightmap_with_colormap(self, heightmap: np.ndarray, output_path: str, colormap: int = cv2.COLORMAP_JET) -> bool:
        try:
            logger.info(f"Saving colored heightmap to {output_path}")

            colored = cv2.applyColorMap(heightmap.astype(np.uint8), colormap)

            cv2.imwrite(output_path, colored)

            logger.info(f"Colored heightmap saved successfully: {output_path}")
            return True

        except Exception as e:
            logger.error(f"Error saving colored heightmap: {e}")
            return False

    def generate_batch(self, images: list, merge_method: str = "average") -> Tuple[Optional[np.ndarray], int, int]:
        try:
            logger.info(
                f"Starting batch heightmap generation for {len(images)} images with method: {merge_method}")

            heightmaps = []
            max_width = 0
            max_height = 0

            for idx, image in enumerate(images):
                logger.info(f"Processing image {idx + 1}/{len(images)}")
                heightmap, width, height = self.generate(image)

                if heightmap is not None:
                    heightmaps.append(heightmap)
                    max_width = max(max_width, width)
                    max_height = max(max_height, height)
                else:
                    logger.warning(
                        f"Failed to generate heightmap for image {idx + 1}")

            if not heightmaps:
                logger.error("No heightmaps generated from batch")
                return None, 0, 0

            logger.info(
                f"Successfully generated {len(heightmaps)} heightmaps, merging with method: {merge_method}")

            resized_heightmaps = []
            for hm in heightmaps:
                if hm.shape != (max_height, max_width):
                    resized = cv2.resize(
                        hm, (max_width, max_height), interpolation=cv2.INTER_LINEAR)
                    resized_heightmaps.append(resized)
                else:
                    resized_heightmaps.append(hm)

            if merge_method == "average":
                merged = np.mean(resized_heightmaps, axis=0).astype(np.uint8)
            elif merge_method == "max":
                merged = np.max(resized_heightmaps, axis=0).astype(np.uint8)
            elif merge_method == "median":
                merged = np.median(resized_heightmaps, axis=0).astype(np.uint8)
            else:
                logger.warning(
                    f"Unknown merge method: {merge_method}, using average")
                merged = np.mean(resized_heightmaps, axis=0).astype(np.uint8)

            logger.info(
                f"Batch heightmap merge complete: {max_width}x{max_height}")
            return merged, max_width, max_height

        except Exception as e:
            logger.error(f"Error in batch heightmap generation: {e}")
            return None, 0, 0
