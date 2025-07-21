# import os, time, json
# import numpy as np
# import cv2
# from dotenv import load_dotenv
# from confluent_kafka import Consumer, Producer, KafkaError

# # ─── Load environment ──────────────────────────────────────────────────────────────
# load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))
# BROKER       = os.getenv("KAFKA_BROKER", "kafka:9092")
# ANALYSIS_T   = os.getenv("KAFKA_ANALYSIS_TOPIC",  "video.analysis")
# RESULTS_T    = os.getenv("KAFKA_RESULTS_TOPIC",   "video.results")
# GROUP_ID     = os.getenv("WORKER_GROUP_ID",       "worker-group")

# # ─── Kafka setup ───────────────────────────────────────────────────────────────────
# consumer = Consumer({
#     "bootstrap.servers": BROKER,
#     "group.id":          GROUP_ID,
#     "auto.offset.reset": "earliest"
# })
# consumer.subscribe([ANALYSIS_T])

# producer = Producer({"bootstrap.servers": BROKER})

# print(f"Listening on {ANALYSIS_T}, producing to {RESULTS_T}")

# # ─── Load the DNN model once at startup ────────────────────────────────────────────
# MODEL_PROTO  = "models/MobileNetSSD_deploy.prototxt"
# MODEL_WEIGHTS= "models/MobileNetSSD_deploy.caffemodel"

# net = cv2.dnn.readNetFromCaffe(MODEL_PROTO, MODEL_WEIGHTS)

# # these are the 21 classes that MobileNet‑SSD knows about
# CLASSES = [
#     "background","aeroplane","bicycle","bird","boat",
#     "bottle","bus","car","cat","chair","cow","diningtable",
#     "dog","horse","motorbike","person","pottedplant",
#     "sheep","sofa","train","tvmonitor"
# ]

# # ─── Frame processing function ────────────────────────────────────────────────────
# def process_frame(frame_bytes):
#     """
#     1) Decode the JPEG bytes into an OpenCV image
#     2) Build a blob and run it through the DNN
#     3) Count how many 'person' and 'car' detections exceed 0.5 confidence
#     4) Return counts and latency
#     """
#     # 1) decode
#     arr = np.frombuffer(frame_bytes, dtype=np.uint8)
#     img = cv2.imdecode(arr, cv2.IMREAD_COLOR)
#     if img is None:
#         return {"persons": 0, "cars": 0, "latency_ms": 0.0}
    
#     print(f"🔍 decoded image shape: {img.shape}") 

#     # 2) blob creation
#     blob = cv2.dnn.blobFromImage(img, 0.007843, (300,300), 127.5)

#     # 3) forward pass
#     start = time.time()
#     net.setInput(blob)
#     detections = net.forward()
#     latency = (time.time() - start) * 1000.0  # ms

#     # 4) count classes of interest
#     (h, w) = img.shape[:2]
#     persons = cars = 0
#     for i in range(detections.shape[2]):
#         confidence = detections[0,0,i,2]
#         if confidence < 0.5:
#             continue
#         idx = int(detections[0,0,i,1])
#         if CLASSES[idx] == "person":
#             persons += 1
#         elif CLASSES[idx] == "car":
#             cars += 1

#     return {"persons": persons, "cars": cars, "latency_ms": latency}

# # ─── Main loop ─────────────────────────────────────────────────────────────────────
# try:
#     while True:
#         msg = consumer.poll(1.0)
#         if msg is None:
#             continue
#         if msg.error():
#             if msg.error().code() != KafkaError._PARTITION_EOF:
#                 print(f"Consumer error: {msg.error()}")
#             continue

#         job = json.loads(msg.value())
#         frame_bytes = bytes(job.get("frame_bytes", ""), "latin1")

#         # run detection
#         res = process_frame(frame_bytes)

#         # build result with timestamp
#         output = {
#             "camera":     job.get("camera", "unknown"),
#             "persons":    res["persons"],
#             "cars":       res["cars"],
#             "latency_ms": res["latency_ms"],
#             "timestamp":  int(time.time() * 1000)
#         }

#         payload = json.dumps(output).encode("utf-8")
#         producer.produce(RESULTS_T, payload)
#         producer.flush()

#         print(f"{output['camera']} → persons={res['persons']} cars={res['cars']} "
#               f"latency={res['latency_ms']:.1f}ms")

# except KeyboardInterrupt:
#     pass
# finally:
#     consumer.close()
#     print("Worker shutting down.")



import os, time, json
import base64                                # 1) import base64
import numpy as np
import cv2
from dotenv import load_dotenv
from confluent_kafka import Consumer, Producer, KafkaError

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
    "auto.offset.reset": "earliest"
})
consumer.subscribe([ANALYSIS_T])

producer = Producer({"bootstrap.servers": BROKER})

print(f"Listening on {ANALYSIS_T}, producing to {RESULTS_T}")

# ─── Load the DNN model once at startup ────────────────────────────────────────────
MODEL_PROTO   = "models/MobileNetSSD_deploy.prototxt"
MODEL_WEIGHTS = "models/MobileNetSSD_deploy.caffemodel"
net = cv2.dnn.readNetFromCaffe(MODEL_PROTO, MODEL_WEIGHTS)

CLASSES = [
    "background","aeroplane","bicycle","bird","boat",
    "bottle","bus","car","cat","chair","cow","diningtable",
    "dog","horse","motorbike","person","pottedplant",
    "sheep","sofa","train","tvmonitor"
]

# ─── Frame processing function ────────────────────────────────────────────────────
def process_frame(frame_bytes):
    arr = np.frombuffer(frame_bytes, dtype=np.uint8)
    img = cv2.imdecode(arr, cv2.IMREAD_COLOR)
    if img is None:
        return {"persons": 0, "cars": 0, "latency_ms": 0.0}

    print(f"🔍 decoded image shape: {img.shape}")  # debug

    blob = cv2.dnn.blobFromImage(img, 0.007843, (300,300), 127.5)
    start = time.time()
    net.setInput(blob)
    detections = net.forward()
    latency = (time.time() - start) * 1000.0

    persons = cars = 0
    for i in range(detections.shape[2]):
        conf = detections[0,0,i,2]
        if conf < 0.5:
            continue
        idx = int(detections[0,0,i,1])
        if CLASSES[idx] == "person":
            persons += 1
        elif CLASSES[idx] == "car":
            cars += 1

    return {"persons": persons, "cars": cars, "latency_ms": latency}

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

        job = json.loads(msg.value())

        # 2) Base64-decode the incoming frame bytes
        b64 = job.get("frame_bytes", "")
        try:
            frame_bytes = base64.b64decode(b64)     # <-- proper decode
        except Exception as e:
            print(f"⚠️ invalid base64 frame: {e}")   # 3) debug bad payload
            continue

        # run detection
        res = process_frame(frame_bytes)

        output = {
            "camera":     job.get("camera", "unknown"),
            "persons":    res["persons"],
            "cars":       res["cars"],
            "latency_ms": res["latency_ms"],
            "timestamp":  int(time.time() * 1000)
        }

        payload = json.dumps(output).encode("utf-8")
        producer.produce(RESULTS_T, payload)
        producer.flush()

        print(f"{output['camera']} → persons={res['persons']} cars={res['cars']} "
              f"latency={res['latency_ms']:.1f}ms")

except KeyboardInterrupt:
    pass
finally:
    consumer.close()
    print("Worker shutting down.")
