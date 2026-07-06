import os, time
import cv2
from confluent_kafka import Producer
from dotenv import load_dotenv

load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))

BROKER   = os.getenv("KAFKA_BROKER", "kafka:9092")
TOPIC    = os.getenv("KAFKA_ANALYSIS_TOPIC", "video.analysis")
SOURCE   = os.getenv("VIDEO_SOURCE", "0")
FPS      = float(os.getenv("INGEST_FPS", "5"))
STREAMID = os.getenv("STREAM_ID", "cam1")

producer = Producer({"bootstrap.servers": BROKER})

src = int(SOURCE) if SOURCE == "0" else SOURCE
cap = cv2.VideoCapture(src)

if not cap.isOpened():
    raise RuntimeError(f"Could not open video source: {SOURCE}")

loop = SOURCE != "0"
interval = 1.0 / FPS


is_file = SOURCE != "0"
source_fps = cap.get(cv2.CAP_PROP_FPS) if is_file else 0
skip_ratio = (source_fps / FPS) if (is_file and source_fps > 0) else 1.0
skip_accum = 0.0
print(f"source_fps={source_fps}, publish_fps={FPS}, skip_ratio={skip_ratio:.2f}")

print(f"Ingesting from {SOURCE} -> topic {TOPIC} at {FPS} fps (stream_id={STREAMID})")

try:
    while True:
        if is_file and skip_ratio > 1.0:
            skip_accum += skip_ratio - 1.0
            while skip_accum >= 1.0:
                if not cap.grab():         
                    break                  
                skip_accum -= 1.0

        ok, frame = cap.read()

        if not ok:
            if loop:
                cap.set(cv2.CAP_PROP_POS_FRAMES, 0)
                continue
            else:
                print("Frame read failed, stopping.")
                break

        ok, buf = cv2.imencode(".jpg", frame)
        if not ok:
            print("Frame encode failed, skipping.")
            continue

        producer.produce(
            TOPIC,
            key=STREAMID.encode("utf-8"),
            value=buf.tobytes(),
            headers=[("timestamp", str(int(time.time() * 1000)))],
        )
        producer.poll(0)

        time.sleep(interval)

except KeyboardInterrupt:
    pass
finally:
    cap.release()
    producer.flush()
    print("Ingestor stopped.")