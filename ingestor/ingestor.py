import os
import time
import json
import base64

import cv2
from dotenv import load_dotenv
from confluent_kafka import Producer

# ─── Load env from root .env ────────────────────────────────────────────────────────
# expects KAFKA_BROKER, KAFKA_ANALYSIS_TOPIC, VIDEO_SOURCE, CAMERA_ID, FPS

load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))

BROKER       = os.getenv("KAFKA_BROKER",       "kafka:9092")
TOPIC        = os.getenv("KAFKA_ANALYSIS_TOPIC","video.analysis")
SOURCE       = os.getenv("VIDEO_SOURCE",       "0")         # '0' = first webcam, or '/path/to/video.mp4'
CAMERA_ID    = os.getenv("CAMERA_ID",          "cam1")
FPS          = float(os.getenv("INGESTOR_FPS", "5"))        # frames per second

# ─── Kafka producer setup ───────────────────────────────────────────────────────────
producer = Producer({
    "bootstrap.servers": BROKER,
    "message.max.bytes": 5000000
})
print(f"📡 Ingestor: reading from {SOURCE} @ {FPS} FPS → topic {TOPIC}")

# ─── OpenCV capture setup ────────────────────────────────────────────────────────────
try:
    src = int(SOURCE)
except ValueError:
    src = SOURCE
cap = cv2.VideoCapture(src)
if not cap.isOpened():
    raise RuntimeError(f"Cannot open video source: {SOURCE}")

interval = 1.0 / FPS

# ─── Main loop ───────────────────────────────────────────────────────────────────────

try:
    while True:
        t0 = time.time()
        ret, frame = cap.read()
        if not ret:
            print("Frame read failed, rewinding…")
            cap.set(cv2.CAP_PROP_POS_FRAMES, 0)
            time.sleep(interval)
            continue

        # 1) resize to 640×480
        small = cv2.resize(frame, (640, 480))
        # 2) encode at 50% JPEG quality
        _, jpg = cv2.imencode(".jpg", small, [int(cv2.IMWRITE_JPEG_QUALITY), 50])
        jpg_bytes = jpg.tobytes()
        b64 = base64.b64encode(jpg_bytes).decode("utf-8")

        # build message
        msg = {
            "camera":      CAMERA_ID,
            "frame_bytes": b64,
            "timestamp":   int(time.time() * 1000),
        }
        producer.produce(TOPIC, json.dumps(msg).encode("utf-8"))
        producer.flush()

        print(f"▶️  Published frame to {TOPIC} ({CAMERA_ID})")

        # throttle to target FPS
        elapsed = time.time() - t0
        if elapsed < interval:
            time.sleep(interval - elapsed)
except KeyboardInterrupt:
    print("Stopping ingestor…")

finally:
    cap.release()
    print("Ingestor shut down.")

