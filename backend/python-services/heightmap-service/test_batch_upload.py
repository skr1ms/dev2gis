import os
import requests
import json
import time
import subprocess
import re
from pathlib import Path


API_BASE_URL = os.getenv("API_BASE_URL", "http://localhost/api")
TEST_EMAIL = "test@example.com"
TEST_PASSWORD = "testpassword123"
TEST_NAME = "Test User"
TEST_DATA_DIR = "tests_data"


def get_nodeodm_progress(batch_job_id):
    try:
        result = subprocess.run(
            ["docker", "logs", "dev-uav-python-workers", "--tail", "100"],
            capture_output=True,
            text=True,
            timeout=5
        )

        logs = result.stdout + result.stderr

        progress_patterns = [
            r'progress["\s:=]+(\d+\.?\d*)',
            r'NodeODM.*?(\d+\.?\d*)%',
            r'status.*?progress.*?(\d+\.?\d*)',
        ]

        for pattern in progress_patterns:
            matches = re.findall(pattern, logs, re.IGNORECASE)
            if matches:
                try:
                    return float(matches[-1])
                except ValueError:
                    continue

        return None
    except Exception:
        return None


def register_or_login():
    register_url = f"{API_BASE_URL}/auth/register"
    login_url = f"{API_BASE_URL}/auth/login"

    register_data = {
        "email": TEST_EMAIL,
        "password": TEST_PASSWORD,
        "name": TEST_NAME
    }

    try:
        response = requests.post(register_url, json=register_data)
        if response.status_code in [200, 201]:
            print(f"Registered new user: {TEST_EMAIL}")
            return response.json()["tokens"]["access_token"]
    except Exception as e:
        print(f"Registration failed (user might exist): {e}")

    login_data = {
        "email": TEST_EMAIL,
        "password": TEST_PASSWORD
    }

    response = requests.post(login_url, json=login_data)
    if response.status_code == 200:
        print(f"Logged in as: {TEST_EMAIL}")
        return response.json()["tokens"]["access_token"]

    raise Exception(
        f"Failed to login: {response.status_code} - {response.text}")


def upload_batch_photos(token, test_data_dir, merge_method="average"):
    upload_url = f"{API_BASE_URL}/heightmaps/batch/upload"

    test_data_path = Path(test_data_dir)
    if not test_data_path.exists():
        raise Exception(f"Test data directory not found: {test_data_dir}")

    image_files = list(test_data_path.glob("*.JPG")) + \
        list(test_data_path.glob("*.jpg"))

    image_files = sorted(set(image_files))

    if not image_files:
        raise Exception(f"No JPG files found in {test_data_dir}")

    print(f"\nFound {len(image_files)} images in {test_data_dir}")
    print("Images:")
    for img in image_files:
        print(f"  - {img.name}")

    files = []
    for img_path in image_files:
        files.append(
            ('files', (img_path.name, open(img_path, 'rb'), 'image/jpeg')))

    headers = {
        "Authorization": f"Bearer {token}"
    }

    data = {
        "merge_method": merge_method
    }

    print(
        f"\nUploading {len(files)} images with merge method: {merge_method}...")

    start_time = time.time()
    response = requests.post(upload_url, files=files,
                             data=data, headers=headers)
    upload_time = time.time() - start_time

    for _, (_, file_obj, _) in files:
        file_obj.close()

    if response.status_code in [200, 202]:
        result = response.json()
        print("\nBatch upload successful!")
        print(f"Upload time: {upload_time:.2f}s")
        print(f"Batch Job ID: {result['id']}")
        print(f"Status: {result['status']}")
        print(f"Image Count: {result['image_count']}")
        print(f"Merge Method: {result['merge_method']}")
        return result['id']
    else:
        raise Exception(
            f"Upload failed: {response.status_code} - {response.text}")


def check_batch_status(token, batch_job_id):
    status_url = f"{API_BASE_URL}/heightmaps/batch/{batch_job_id}"

    headers = {
        "Authorization": f"Bearer {token}"
    }

    response = requests.get(status_url, headers=headers)

    if response.status_code == 200:
        return response.json()
    else:
        raise Exception(
            f"Failed to get status: {response.status_code} - {response.text}")


def wait_for_completion(token, batch_job_id, timeout=600, check_interval=5):
    print(
        f"\nWaiting for batch processing to complete (timeout: {timeout}s)...")

    start_time = time.time()
    consecutive_errors = 0
    max_consecutive_errors = 3

    while time.time() - start_time < timeout:
        try:
            status = check_batch_status(token, batch_job_id)
            consecutive_errors = 0

            current_status = status['status']
            processed = status.get('processed_count', 0)
            total = status.get('image_count', 0)

            elapsed = time.time() - start_time

            nodeodm_progress = get_nodeodm_progress(batch_job_id)
            if nodeodm_progress is not None and current_status == 'processing':
                print(f"[{elapsed:.0f}s] NodeODM Progress: {nodeodm_progress:.1f}%")
            else:
                print(f"[{elapsed:.0f}s] Status: {current_status}")

            if current_status == 'completed':
                print("\nBatch processing completed successfully!")
                print(
                    f"Total processing time: {status.get('processing_time', 0):.2f}s")
                print(
                    f"Result dimensions: {status.get('width')}x{status.get('height')}")
                if status.get('result_url'):
                    print(f"Result URL: {status['result_url']}")
                return status

            elif current_status == 'failed':
                error_msg = status.get('error_message', 'Unknown error')
                raise Exception(f"Batch processing failed: {error_msg}")

            time.sleep(check_interval)

        except (requests.exceptions.ConnectionError, ConnectionResetError, ConnectionAbortedError) as e:
            consecutive_errors += 1
            print(
                f"Connection error (attempt {consecutive_errors}/{max_consecutive_errors}), retrying...")

            if consecutive_errors >= max_consecutive_errors:
                print(
                    "Too many connection errors, checking if job completed via direct query...")
                try:
                    final_status = check_batch_status(token, batch_job_id)
                    if final_status['status'] == 'completed':
                        print(
                            "\nJob completed successfully (verified after connection errors)!")
                        return final_status
                except:
                    pass
                raise Exception(
                    f"Connection lost after {max_consecutive_errors} attempts")

            time.sleep(check_interval * 2)

        except Exception as e:
            print(f"Error checking status: {e}")
            time.sleep(check_interval)

    raise Exception(f"Timeout waiting for batch processing after {timeout}s")


def main():
    print("=" * 60)
    print("Batch Heightmap Upload Test")
    print("=" * 60)

    try:
        print("\nStep 1: Authentication")
        token = register_or_login()

        print("\nStep 2: Batch Upload")
        batch_job_id = upload_batch_photos(
            token, TEST_DATA_DIR, merge_method="average")

        print("\nStep 3: Monitor Processing")
        result = wait_for_completion(token, batch_job_id)

        print("\n" + "=" * 60)
        print("Test completed successfully!")
        print("=" * 60)
        print(json.dumps(result, indent=2))

    except Exception as e:
        print(f"\nTest failed: {e}")
        import traceback
        traceback.print_exc()
        exit(1)


if __name__ == "__main__":
    main()
