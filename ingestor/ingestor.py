import logging
import os
import threading
import time
import random

os.environ["OPENCV_FFMPEG_CAPTURE_OPTIONS"] = "rtsp_transport;tcp|stimeout;10000000"

import cv2
from dotenv import load_dotenv
from confluent_kafka import Producer
import httpx
import math

load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))

BROKER = os.getenv("KAFKA_BROKER", "kafka:9092")
TOPIC = os.getenv("KAFKA_ANALYSIS_TOPIC", "video.analysis")
CAMERA_URL = os.getenv("CAMERA_URL", "")
AUTH_API_KEY = os.getenv("AUTH_API_KEY", "")
FPS = float(os.getenv("INGESTOR_FPS", "5"))
producer = Producer({
    "bootstrap.servers": BROKER,
    "message.max.bytes": 5000000,
})

interval = 1.0 / FPS
base_backoff = 2
max_backoff = 10


def fetch_cameras():
    response = httpx.get(CAMERA_URL, headers={"X-API-Key": AUTH_API_KEY})
    response.raise_for_status()

    return response.json()["cameras"]


def run_camera(camera, stop_event):
    cap = cv2.VideoCapture(camera["video_source"])
    backoff = base_backoff
    cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)
    src_fps = cap.get(cv2.CAP_PROP_FPS)
    drain_count = max(1, math.ceil(src_fps / FPS)) if src_fps > 0 else 30  
    
    if not cap.isOpened():
        logging.error("Failed to open camera/video source")
        return

    try:
        while not stop_event.is_set():
            t0 = time.time()

            ret = False
            
            for _ in range(drain_count):
                
                if not cap.grab():
                    break
                    
            ret, frame = cap.retrieve()
                
            if not ret:
                print("Frame read failed, reconnecting…")

                cap.release()
                cap = cv2.VideoCapture(camera["video_source"])
                cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)
                
                if not cap.isOpened():
                    sleep_time = backoff * random.uniform(0.8, 1.2)
                    time.sleep(sleep_time)
                    backoff = min(backoff * 2, max_backoff)
                    continue

                backoff = base_backoff
                continue

            small = cv2.resize(frame, (640, 480))

            _, jpg = cv2.imencode(
                ".jpg",
                small,
                [int(cv2.IMWRITE_JPEG_QUALITY), 50],
            )

            jpg_bytes = jpg.tobytes()

            timestamp = int(time.time() * 1000)

            producer.produce(
                TOPIC,
                key=camera["camera_name"].encode("utf-8"),
                value=jpg_bytes,
                headers=[
                    ("timestamp", str(timestamp).encode("utf-8")),
                ],
            )

            producer.poll(0)

            elapsed = time.time() - t0

            if elapsed < interval:
                time.sleep(interval - elapsed)

    finally:
        cap.release()
        producer.flush()
        print("Ingestor shut down.")


def main():
    threads = {}

    try:
        while True:
            try:
                cameras = fetch_cameras()

            except Exception as e:
                logging.error(f"failed to fetch cameras: {e}")
                time.sleep(30)
                continue
            
            if cameras is not None:
                active_ids = {c["camera_id"] for c in cameras}
                
                for camera in cameras:
                    if camera["camera_id"] not in threads:
                        stop = threading.Event()

                        t = threading.Thread(
                            target=run_camera,
                            args=(camera, stop),
                            daemon=True,
                        )

                        threads[camera["camera_id"]] = (t, stop)
                        t.start()

                for camera_id in list(threads.keys()):
                    if camera_id not in active_ids:
                        threads[camera_id][1].set()
                        threads[camera_id][0].join()
                        del threads[camera_id]

            time.sleep(30)

    except KeyboardInterrupt:
        print("Stopping ingestor…")
        print("Shutting down all camera threads...")

        for camera_id, (t, stop) in threads.items():
            stop.set()

        for camera_id, (t, stop) in threads.items():
            t.join()


if __name__ == "__main__":
    main()
