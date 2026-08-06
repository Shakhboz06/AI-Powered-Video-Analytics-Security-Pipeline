
import json
import os
import time

import cv2
import httpx
from confluent_kafka import Consumer, KafkaError, Producer
from dotenv import load_dotenv

load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))

BROKER        = os.getenv("KAFKA_BROKER", "kafka:9092")
UPLOADS_TOPIC = os.getenv("KAFKA_UPLOADS_TOPIC", "video.uploads")
ANALYSIS_T    = os.getenv("KAFKA_ANALYSIS_TOPIC", "video.analysis")
GROUP_ID      = os.getenv("UPLOAD_GROUP_ID", "upload-ingestor")
PUBLISH_FPS   = float(os.getenv("INGEST_FPS", "5"))
MAX_SECONDS   = float(os.getenv("UPLOAD_MAX_SECONDS", "180"))
FRAME_DELAY   = float(os.getenv("UPLOAD_FRAME_DELAY", "0.05"))
KEEP_FILES    = os.getenv("UPLOAD_KEEP_FILES", "false").lower() == "true"

DASHBOARD_URL = os.getenv("DASHBOARD_INTERNAL_URL", "http://dashboard:8081")
API_KEY       = os.getenv("AUTH_API_KEY", "")

consumer = Consumer({
    "bootstrap.servers": BROKER,
    "group.id": GROUP_ID,
    "auto.offset.reset": "earliest",
})
consumer.subscribe([UPLOADS_TOPIC])

producer = Producer({"bootstrap.servers": BROKER})


def report(job_id, status, progress, error=None, last_frame_ms=None):
    payload = {"status": status, "progress": int(progress)}
    if error:
        payload["error"] = str(error)[:500]
    if last_frame_ms is not None:
        payload["last_frame_ms"] = int(last_frame_ms)
    try:
        resp = httpx.patch(
            f"{DASHBOARD_URL}/api/v1/internal/uploads/{job_id}",
            json=payload,
            headers={"X-API-Key": API_KEY},
            timeout=5.0,
        )
        if resp.status_code != 200:
            print(f"status callback for {job_id} returned {resp.status_code}: {resp.text}")
    except Exception as e:
        print(f"status callback for {job_id} failed: {e}")


def process_job(job):
    job_id = job["job_id"]
    stream_id = job["stream_id"]
    path = job["path"]

    cap = cv2.VideoCapture(path)
    if not cap.isOpened():
        report(job_id, "failed", 0, error="could not open the uploaded video")
        return

    source_fps = cap.get(cv2.CAP_PROP_FPS) or 0
    total_frames = cap.get(cv2.CAP_PROP_FRAME_COUNT) or 0
    if source_fps <= 0:
        source_fps = 30.0

    duration_s = total_frames / source_fps if total_frames > 0 else MAX_SECONDS
    analyzed_s = min(duration_s, MAX_SECONDS)

    skip_ratio = max(source_fps / PUBLISH_FPS, 1.0)
    skip_accum = 0.0

    base_ms = int(time.time() * 1000)

    report(job_id, "processing", 0)

    frame_idx = 0
    published = 0
    last_ts = base_ms
    last_report = time.time()

    try:
        while True:
            if skip_ratio > 1.0:
                skip_accum += skip_ratio - 1.0
                while skip_accum >= 1.0:
                    if not cap.grab():
                        break
                    frame_idx += 1
                    skip_accum -= 1.0

            ok, frame = cap.read()
            if not ok:
                break
            frame_idx += 1

            position_s = frame_idx / source_fps
            if position_s > MAX_SECONDS:
                print(f"job {job_id}: reached {MAX_SECONDS}s analysis cap")
                break

            ok, buf = cv2.imencode(".jpg", frame)
            if not ok:
                continue

            last_ts = base_ms + int(position_s * 1000)
            producer.produce(
                ANALYSIS_T,
                key=stream_id.encode("utf-8"),
                value=buf.tobytes(),
                headers=[("timestamp", str(last_ts))],
            )
            producer.poll(0)
            published += 1

            if time.time() - last_report >= 2.0:
                progress = min(int(position_s / analyzed_s * 100), 99) if analyzed_s > 0 else 0
                report(job_id, "processing", progress)
                last_report = time.time()

            if FRAME_DELAY > 0:
                time.sleep(FRAME_DELAY)

    except Exception as e:
        cap.release()
        producer.flush()
        report(job_id, "failed", 0, error=f"analysis failed: {e}")
        print(f"❌ job {job_id} failed: {e}")
        return
    finally:
        cap.release()

    producer.flush()

    if published == 0:
        report(job_id, "failed", 0, error="no readable frames in the uploaded video")
        print(f"❌ job {job_id}: no frames published")
        return

    report(job_id, "finalizing", 99, last_frame_ms=last_ts)
    print(f"job {job_id}: published {published} frames, awaiting worker to finish")

    if not KEEP_FILES:
        try:
            os.remove(path)
        except OSError as e:
            print(f"could not remove {path}: {e}")


try:
    while True:
        msg = consumer.poll(1.0)
        if msg is None:
            continue
        if msg.error():
            if msg.error().code() != KafkaError._PARTITION_EOF:
                print(f"Consumer error: {msg.error()}")
            continue

        try:
            job = json.loads(msg.value())
        except (json.JSONDecodeError, TypeError) as e:
            print(f"Skipping malformed job message: {e}")
            continue

        if not all(k in job for k in ("job_id", "stream_id", "path")):
            print(f"Skipping incomplete job message: {job}")
            continue

        process_job(job)

except KeyboardInterrupt:
    pass
finally:
    consumer.close()
    producer.flush()
    print("Upload ingestor stopped.")
