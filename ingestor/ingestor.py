# import os, time
# import cv2
# from confluent_kafka import Producer
# from dotenv import load_dotenv

# load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))

# BROKER   = os.getenv("KAFKA_BROKER", "kafka:9092")
# TOPIC    = os.getenv("KAFKA_ANALYSIS_TOPIC", "video.analysis")
# SOURCE   = os.getenv("VIDEO_SOURCE", "0")
# FPS      = float(os.getenv("INGEST_FPS", "5"))
# STREAMID = os.getenv("STREAM_ID", "cam1")

# producer = Producer({"bootstrap.servers": BROKER})

# src = int(SOURCE) if SOURCE == "0" else SOURCE
# cap = cv2.VideoCapture(src)

# if not cap.isOpened():
#     raise RuntimeError(f"Could not open video source: {SOURCE}")

# loop = SOURCE != "0"
# interval = 1.0 / FPS


# is_file = SOURCE != "0"
# source_fps = cap.get(cv2.CAP_PROP_FPS) if is_file else 0
# skip_ratio = (source_fps / FPS) if (is_file and source_fps > 0) else 1.0
# skip_accum = 0.0
# print(f"source_fps={source_fps}, publish_fps={FPS}, skip_ratio={skip_ratio:.2f}")

# print(f"Ingesting from {SOURCE} -> topic {TOPIC} at {FPS} fps (stream_id={STREAMID})")

# try:
#     while True:
#         if is_file and skip_ratio > 1.0:
#             skip_accum += skip_ratio - 1.0
#             while skip_accum >= 1.0:
#                 if not cap.grab():         
#                     break                  
#                 skip_accum -= 1.0

#         ok, frame = cap.read()

#         if not ok:
#             if loop:
#                 cap.set(cv2.CAP_PROP_POS_FRAMES, 0)
#                 continue
#             else:
#                 print("Frame read failed, stopping.")
#                 break

#         ok, buf = cv2.imencode(".jpg", frame)
#         if not ok:
#             print("Frame encode failed, skipping.")
#             continue

#         producer.produce(
#             TOPIC,
#             key=STREAMID.encode("utf-8"),
#             value=buf.tobytes(),
#             headers=[("timestamp", str(int(time.time() * 1000)))],
#         )
#         producer.poll(0)

#         time.sleep(interval)

# except KeyboardInterrupt:
#     pass
# finally:
#     cap.release()
#     producer.flush()
#     print("Ingestor stopped.")



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

print(f"📡 Ingestor: reading from {CAMERA_URL} @ {FPS} FPS → topic {TOPIC}")

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

    if not cap.isOpened():
        logging.error("Failed to open camera/video source")
        return

    try:
        while not stop_event.is_set():
            t0 = time.time()

            ret, frame = cap.read()

            if not ret:
                print("Frame read failed, reconnecting…")

                cap.release()
                cap = cv2.VideoCapture(camera["video_source"])

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

            producer.flush()

            print(f"▶️  Published frame to {TOPIC} ({camera['camera_id']})")

            elapsed = time.time() - t0

            if elapsed < interval:
                time.sleep(interval - elapsed)

    finally:
        cap.release()
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