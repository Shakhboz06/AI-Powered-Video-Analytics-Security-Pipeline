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
WEAPON_MODEL = os.getenv("WEAPON_MODEL_PATH", "models/weapons.pt")


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
detection_model = YOLO("yolo11n.pt")
pose_model= YOLO("yolo11n-pose.pt")
weapon_model=YOLO(WEAPON_MODEL)

tracker = sv.ByteTrack()
trackers = {}

def process_frame(frame_bytes, stream_id):

    tracker = trackers[stream_id]
    arr = np.frombuffer(frame_bytes, dtype=np.uint8)
    img = cv2.imdecode(arr, cv2.IMREAD_COLOR)
    if img is None:
        return {"detections": [], "latency_ms": 0.0}

    print(f"🔍 decoded image shape: {img.shape}")

    start = time.time()
    results = detection_model.predict(img, conf=0.25)

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

    pose_results = pose_model.predict(img, conf=0.25)
    weapon_results = weapon_model.predict(img, conf=0.4, verbose=False, save=False)

    for r in weapon_results:
        print("weapon_detections: ", r.boxes)

        
    keypoints = []

    for r in pose_results:

        if r.keypoints is None:
            continue

        keypoints_xy = r.keypoints.xy.cpu().numpy()
        keypoints_conf = r.keypoints.conf.cpu().numpy()
        boxes = r.boxes.xyxy.cpu().numpy()
        
        for i in range(len(boxes)):
            pose_box = boxes[i]              
            person_xy = keypoints_xy[i]           
            person_conf = keypoints_conf[i] 

            best_iou = 0
            best_tracker_id = None
            for det in list_det:
                computed_iou = iou(pose_box, det['bound_box'])
                if computed_iou > best_iou:
                    best_iou = computed_iou
                    best_tracker_id = det['tracker_id']

            if best_tracker_id is not None and best_iou > 0.3:
                keypoints.append({
                    "tracker_id": best_tracker_id,
                    "points": {
                        "xy": person_xy.tolist(),
                        "conf": person_conf.tolist()
                    }
                })


    for w in weapon_results:
        if w.boxes is None:
            continue
    
        boxes = w.boxes.xyxy.cpu().numpy()
        classes = w.boxes.cls.cpu().numpy()
        confs = w.boxes.conf.cpu().numpy()

        for i in range(len(boxes)):
            class_id = int(classes[i])
            label = weapon_model.names[class_id]

            list_det.append({
                "class_id": class_id,
                "conf_score": float(confs[i]),
                "bound_box": boxes[i].tolist(),
                "tracker_id": -1, # weapons are not tracked
                "label": label
            })
           
    return {"detections": list_det, "keypoints": keypoints, "latency_ms": latency}

def iou(box1, box2):
    
    x1 = max(box1[0], box2[0])
    y1 = max(box1[1], box2[1])
    x2 = min(box1[2], box2[2])
    y2 = min(box1[3], box2[3])
    
    intersection = max(0, x2-x1) * max(0, y2-y1)
    
    area1 = (box1[2]-box1[0]) * (box1[3]-box1[1])
    area2 = (box2[2]-box2[0]) * (box2[3]-box2[1])
    union = area1 + area2 - intersection
    
    return intersection / union if union > 0 else 0

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

        res = process_frame(job, stream_id)
        
        output = {
            "camera":     msg.key().decode("utf-8"),
            "detections": res["detections"],
            "keypoints":  res["keypoints"],
            "latency_ms": res["latency_ms"],
            "timestamp":  int(decode_header(msg.headers(), "timestamp")),
        }

        payload = json.dumps(output).encode("utf-8")
        producer.produce(RESULTS_T, payload)
        producer.flush()
                                                                
        print(f"{output['camera']} → detections={res['detections']} →  keypoints={res['keypoints']}"
              f"latency={res['latency_ms']:.1f}ms")

except KeyboardInterrupt:
    pass
finally:
    consumer.close()
    print("Worker shutting down.")
