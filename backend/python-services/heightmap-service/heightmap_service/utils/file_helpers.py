import os
import shutil
import tempfile
from contextlib import contextmanager


def ensure_dir(directory: str) -> None:
    os.makedirs(directory, exist_ok=True)


def cleanup_temp_files(*file_paths: str) -> None:
    for file_path in file_paths:
        try:
            if os.path.exists(file_path):
                os.remove(file_path)
        except Exception:
            pass


def get_file_size(file_path: str) -> int:
    return os.path.getsize(file_path) if os.path.exists(file_path) else 0


@contextmanager
def temp_directory():
    tmp_dir = tempfile.mkdtemp()
    try:
        yield tmp_dir
    finally:
        shutil.rmtree(tmp_dir, ignore_errors=True)


def clean_old_files(directory: str, max_age_seconds: int = 3600) -> int:
    import time
    cleaned = 0
    current_time = time.time()

    if not os.path.exists(directory):
        return 0

    for filename in os.listdir(directory):
        file_path = os.path.join(directory, filename)
        if os.path.isfile(file_path):
            file_age = current_time - os.path.getmtime(file_path)
            if file_age > max_age_seconds:
                try:
                    os.remove(file_path)
                    cleaned += 1
                except Exception:
                    pass

    return cleaned
