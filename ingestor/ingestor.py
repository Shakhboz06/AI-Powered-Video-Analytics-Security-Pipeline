import logging
import os
import threading
import time
import cv2
from dotenv import load_dotenv
from confluent_kafka import Producer
import httpx

# ─── Load env from root .env ────────────────────────────────────────────────────────
# expects KAFKA_BROKER, KAFKA_ANALYSIS_TOPIC, VIDEO_SOURCE, CAMERA_ID, FPS

load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))

BROKER       = os.getenv("KAFKA_BROKER",       "kafka:9092")
TOPIC        = os.getenv("KAFKA_ANALYSIS_TOPIC","video.analysis")
CAMERA_URL      = os.getenv("CAMERA_URL", "")   
AUTH_API_KEY = os.getenv("AUTH_API_KEY","")    
FPS          = float(os.getenv("INGESTOR_FPS", "5"))        # frames per second

# ─── Kafka producer setup ───────────────────────────────────────────────────────────
producer = Producer({
    "bootstrap.servers": BROKER,
    "message.max.bytes": 5000000
})
print(f"📡 Ingestor: reading from {CAMERA_URL} @ {FPS} FPS → topic {TOPIC}")

# ─── OpenCV capture setup ────────────────────────────────────────────────────────────

interval = 1.0 / FPS
def fetch_cameras():
    response = httpx.get(CAMERA_URL, headers={"X-API-Key": AUTH_API_KEY})
    response.raise_for_status()
    return response.json()['cameras']

def run_camera(camera, stop_event):
    cap = cv2.VideoCapture(camera['video_source'])
    if not cap.isOpened():
        logging.error("Failed to open camera/video source")
        return
    
    try:
        while not stop_event.is_set():
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

            # build message

            timestamp = int(time.time() * 1000)

            producer.produce(TOPIC, key=camera['camera_name'].encode("utf-8"), value=jpg_bytes, headers=[("timestamp", str(timestamp).encode("utf-8"))])        
            producer.flush()

            print(f"▶️  Published frame to {TOPIC} ({camera['camera_id']})")

            # throttle to target FPS
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
            
            active_ids = {c['camera_id'] for c in cameras}

            for camera in cameras:
                if camera['camera_id'] not in threads:
                    stop = threading.Event()
                    t = threading.Thread(target=run_camera, args=(camera, stop), daemon=True)
                    threads[camera['camera_id']] = (t, stop)
                    t.start()
            
            # Stop removed ones
            for camera_id in list(threads.keys()):
                if camera_id not in active_ids:
                    threads[camera_id][1].set()  # signal stop
                    threads[camera_id][0].join()  # wait
                    del threads[camera_id]
            
            time.sleep(30)
    except KeyboardInterrupt:
        print("Stopping ingestor…")
        print("Shutting down all camera threads...")
        for camera_id, (t, stop) in threads.items():
            stop.set()
        for camera_id, (t, stop) in threads.items():
            t.join()

if __name__ == "__main__": main()