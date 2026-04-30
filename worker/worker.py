import os, time, json
import numpy as np
import cv2
from dotenv import load_dotenv
from confluent_kafka import Consumer, Producer, KafkaError
from ultralytics import YOLO
import supervision as sv

# ─── Load environment ──────────────────────────────────────────────────────────────
load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))
BROKER       = os.getenv("KAFKA_BROKER", "kafka:9092")
ANALYSIS_T   = os.getenv("KAFKA_ANALYSIS_TOPIC",  "video.analysis")
RESULTS_T    = os.getenv("KAFKA_RESULTS_TOPIC",   "video.results")
GROUP_ID     = os.getenv("WORKER_GROUP_ID",       "worker-group")

# ─── Kafka setup ───────────────────────────────────────────────────────────────────
consumer = Consumer({
    "bootstrap.servers": BROKER,
    "group.id":          GROUP_ID,
    "auto.offset.reset": "latest"
})
consumer.subscribe([ANALYSIS_T])

producer = Producer({"bootstrap.servers": BROKER})

print(f"Listening on {ANALYSIS_T}, producing to {RESULTS_T}")



def decode_header(headers, key):
    val = None
    if headers is not None:
        headers = dict(headers)
        val = headers.get(key)
    return val.decode("utf-8") if val else None

# ─── Frame processing function ────────────────────────────────────────────────────
model = YOLO("yolo11n.pt")
# tracker = sv.ByteTrack()
trackers = {}

def process_frame(frame_bytes, stream_id):

    tracker = trackers[stream_id]
    arr = np.frombuffer(frame_bytes, dtype=np.uint8)
    img = cv2.imdecode(arr, cv2.IMREAD_COLOR)
    if img is None:
        return {"detections": [], "latency_ms": 0.0}

    print(f"🔍 decoded image shape: {img.shape}")

    start = time.time()
    results = model.predict(img, conf=0.25)

    detections = sv.Detections.from_ultralytics(results[0])
    detections = tracker.update_with_detections(detections)
    

    list_det = []
    
    for item in detections: 
            bound_box = item[0].tolist()
            conf_score = item[2].tolist()
            class_id = item[3].tolist()
            tracker_id = item[4].tolist()
            list_det.append({
                "class_id": class_id,
                "conf_score": conf_score,
                "bound_box": bound_box,
                "tracker_id": tracker_id,
                "label": str(item[5]['class_name'])
            })


    latency = (time.time() - start) * 1000.0

    return {"detections": list_det, "latency_ms": latency}



# ─── Main loop ─────────────────────────────────────────────────────────────────────
try:
    while True:
        msg = consumer.poll(1.0)
        if msg is None:
            continue
        if msg.error():
            if msg.error().code() != KafkaError._PARTITION_EOF:
                print(f"Consumer error: {msg.error()}")
            continue

        job = msg.value()
        stream_id = msg.key().decode("utf-8")

        if stream_id not in trackers:
            trackers[stream_id] = sv.ByteTrack()

        res = process_frame(job,  stream_id)
        
        output = {
            "camera":     msg.key().decode("utf-8"),
            "detections": res["detections"],
            "latency_ms": res["latency_ms"],
            "timestamp":  int(decode_header(msg.headers(), "timestamp")),
        }

        payload = json.dumps(output).encode("utf-8")
        producer.produce(RESULTS_T, payload)
        producer.flush()

        print(f"{output['camera']} → detections={res['detections']} "
              f"latency={res['latency_ms']:.1f}ms")

except KeyboardInterrupt:
    pass
finally:
    consumer.close()
    print("Worker shutting down.")

