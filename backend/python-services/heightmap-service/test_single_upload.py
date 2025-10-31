import requests
import time
import json
from pathlib import Path

API_URL = "http://localhost/api"
EMAIL = "test@example.com"
PASSWORD = "testpassword123"


def main():
    print("=" * 60)
    print("Single Image Heightmap Test")
    print("=" * 60)

    print("\nStep 1: Authentication")
    response = requests.post(f"{API_URL}/auth/login", json={
        "email": EMAIL,
        "password": PASSWORD
    })

    if response.status_code != 200:
        print(f"Login failed: {response.status_code} - {response.text}")
        return

    response_data = response.json()

    if "tokens" in response_data and "access_token" in response_data["tokens"]:
        token = response_data["tokens"]["access_token"]
    elif "token" in response_data:
        token = response_data["token"]
    else:
        print(f"No token in response: {response_data}")
        return

    print(f"Logged in as: {EMAIL}")

    print("\nStep 2: Upload Single Image")

    image_path = Path("tests_data/DJI_0161.JPG")

    with open(image_path, "rb") as f:
        response = requests.post(
            f"{API_URL}/heightmaps/upload",
            headers={"Authorization": f"Bearer {token}"},
            files={"file": (image_path.name, f, "image/jpeg")}
        )

    if response.status_code not in [200, 202]:
        print(f"Upload failed: {response.status_code}")
        print(response.text)
        return

    result = response.json()
    job_id = result["id"]

    print("Upload successful!")
    print(f"Job ID: {job_id}")
    print(f"Status: {result['status']}")

    print("\nStep 3: Wait for Processing")

    max_wait = 60
    check_interval = 2
    consecutive_errors = 0

    for elapsed in range(0, max_wait, check_interval):
        time.sleep(check_interval)

        try:
            response = requests.get(
                f"{API_URL}/heightmaps/{job_id}",
                headers={"Authorization": f"Bearer {token}"}
            )

            if response.status_code != 200:
                print(f"Error: {response.status_code}")
                continue

            consecutive_errors = 0
            status = response.json()
            state = status.get("status")

            print(f"[{elapsed}s] Status: {state}")

        except (requests.exceptions.ConnectionError, ConnectionResetError) as e:
            consecutive_errors += 1
            if consecutive_errors < 3:
                print(f"[{elapsed}s] Connection error, retrying...")
                continue
            else:
                print("Too many connection errors, but processing may have completed")
                break

        if state == "completed":
            print("\n" + "=" * 60)
            print("SUCCESS! Heightmap generated with MiDaS")
            print("=" * 60)
            print(json.dumps(status, indent=2))

            result_url = status.get("result_url", "")
            print(f"\nResult URL: {result_url}")
            print(f"Dimensions: {status['width']}x{status['height']}")
            print(f"Processing time: {status['processing_time']:.2f}s")

            print("\nCheck MinIO at http://localhost:9001")
            print("Bucket: uav-models")
            print(f"Path: heightmaps/{job_id}_heightmap.png")

            return

        elif state == "failed":
            print("\nProcessing FAILED!")
            print(json.dumps(status, indent=2))
            return

    print("\nTimeout waiting for processing")


if __name__ == "__main__":
    main()
