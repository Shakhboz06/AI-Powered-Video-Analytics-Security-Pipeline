import json
import os
import tempfile
import time

import cv2
import httpx
from confluent_kafka import Consumer, KafkaError, Producer
from dotenv import load_dotenv

load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))

BROKER      = os.getenv("KAFKA_BROKER", "kafka:9092")
JOBS_TOPIC  = os.getenv("KAFKA_UPLOAD_JOBS_TOPIC", "video.upload_jobs")
ANALYSIS_T  = os.getenv("KAFKA_ANALYSIS_TOPIC", "video.analysis")
GROUP_ID    = os.getenv("UPLOAD_INGESTOR_GROUP_ID", "upload-ingestor-group")
FPS         = float(os.getenv("INGEST_FPS", "5"))

DASHBOARD_API = os.getenv("DASHBOARD_API_URL", "http://dashboard:8081").rstrip("/")
API_KEY       = os.getenv("AUTH_API_KEY", "")

consumer = Consumer({
    "bootstrap.servers": BROKER,
    "group.id": GROUP_ID,
    "auto.offset.reset": "earliest",
})
consumer.subscribe([JOBS_TOPIC])

producer = Producer({"bootstrap.servers": BROKER})


def get_job_status(job_id):
    try:
        resp = httpx.get(f"{DASHBOARD_API}/api/uploads/{job_id}", timeout=5.0)
        if resp.status_code != 200:
            print(f"status lookup for {job_id} returned {resp.status_code}")
            return None
        return resp.json().get("job", {}).get("status")
    except Exception as e:
        print(f"status lookup for {job_id} failed: {e}")
        return None


def set_job_status(job_id, status):
    try:
        resp = httpx.patch(
            f"{DASHBOARD_API}/api/uploads/{job_id}/status",
            json={"status": status},
            headers={"X-API-Key": API_KEY},
            timeout=5.0,
        )
        if resp.status_code != 200:
            print(f"status update {job_id} → {status} returned {resp.status_code}: {resp.text}")
    except Exception as e:
        print(f"status update {job_id} → {status} failed: {e}")


def open_video(video_url):
  
    cap = cv2.VideoCapture(video_url)
    if cap.isOpened():
        return cap, None

    cap.release()
    print("direct URL open failed, downloading to a temp file")

    tmp = tempfile.NamedTemporaryFile(suffix=".mp4", delete=False)
    try:
        with httpx.stream("GET", video_url, timeout=120.0) as resp:
            resp.raise_for_status()
            for chunk in resp.iter_bytes():
                tmp.write(chunk)
        tmp.close()
    except Exception:
        tmp.close()
        os.unlink(tmp.name)
        raise

    cap = cv2.VideoCapture(tmp.name)
    if not cap.isOpened():
        cap.release()
        os.unlink(tmp.name)
        raise RuntimeError("could not open the downloaded video")

    return cap, tmp.name


def process_job(job_id, video_url):
    set_job_status(job_id, "processing")

    cap, temp_path = open_video(video_url)

    interval = 1.0 / FPS
    source_fps = cap.get(cv2.CAP_PROP_FPS)
    skip_ratio = (source_fps / FPS) if source_fps > 0 else 1.0
    skip_accum = 0.0

    frames = 0
    try:
        while True:
            if skip_ratio > 1.0:
                skip_accum += skip_ratio - 1.0
                while skip_accum >= 1.0:
                    if not cap.grab():
                        break
                    skip_accum -= 1.0

            ok, frame = cap.read()
            if not ok:
                break

            ok, buf = cv2.imencode(".jpg", frame)
            if not ok:
                print("Frame encode failed, skipping.")
                continue

            producer.produce(
                ANALYSIS_T,
                key=job_id.encode("utf-8"),
                value=buf.tobytes(),
                headers=[("timestamp", str(int(time.time() * 1000)))],
            )
            producer.poll(0)
            frames += 1

            time.sleep(interval)
    finally:
        cap.release()
        producer.flush()
        if temp_path:
            try:
                os.unlink(temp_path)
            except OSError:
                pass

    if frames == 0:
        raise RuntimeError("no readable frames in the video")


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
            job_id = job["job_id"]
            video_url = job["video_url"]
        except (json.JSONDecodeError, KeyError, TypeError) as e:
            print(f"Skipping malformed job message: {e}")
            continue

        status = get_job_status(job_id)
        if status != "queued":
            print(f"job {job_id}: status is {status!r}, skipping (redelivery?)")
            continue

        try:
            process_job(job_id, video_url)
            set_job_status(job_id, "done")
        except Exception as e:
            # jobs must never hang in 'processing' forever
            set_job_status(job_id, "failed")
            print(f"job {job_id} failed: {e}")

except KeyboardInterrupt:
    pass
finally:
    consumer.close()
    producer.flush()
    print("Upload ingestor stopped.")
