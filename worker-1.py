import os, time, json
import numpy as np
import cv2
from dotenv import load_dotenv
from confluent_kafka import Consumer, Producer, KafkaError
from ultralytics import YOLO
import supervision as sv
import torch
import torch.nn as nn
from collections import defaultdict, deque
from pytorchvideo.models.hub import i3d_r50
from supabase import create_client, Client


# ─── Load environment ──────────────────────────────────────────────────────────────
load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))
BROKER       = os.getenv("KAFKA_BROKER", "kafka:9092")
ANALYSIS_T   = os.getenv("KAFKA_ANALYSIS_TOPIC",  "video.analysis")
RESULTS_T    = os.getenv("KAFKA_RESULTS_TOPIC",   "video.results")
ALERT_NOTIFICATIONS_T = os.getenv("KAFKA_ALERTS_N_TOPIC", "video.alert_notifications")
GROUP_ID     = os.getenv("WORKER_GROUP_ID",       "worker-group")
ALERT_GROUP_ID     = os.getenv("ALERT_GROUP_ID",       "alert-worker-group")
DETECTION_MODEL = os.getenv("DETECTION_MODEL_PATH", "models/yolo26s.pt")
WEAPON_MODEL = os.getenv("WEAPON_MODEL_PATH", "models/weapons.pt")
FALL_MODEL = os.getenv("FALL_MODEL_PATH", "models/fall_classifier_v2.pt")
FIGHT_MODEL = os.getenv("FIGHT_MODEL_PATH", "models/fight_detector_v2.pt")
POSE_MODEL= os.getenv("POSE_MODEL_PATH", "/models/yolo26m-pose.pt")

# ─── Kafka setup ───────────────────────────────────────────────────────────────────
frame_consumer = Consumer({
    "bootstrap.servers": BROKER,
    "group.id":          GROUP_ID,
    "auto.offset.reset": "latest"
})
frame_consumer.subscribe([ANALYSIS_T])

producer = Producer({"bootstrap.servers": BROKER})

print(f"Listening on {ANALYSIS_T}, producing to {RESULTS_T}")

alert_consumer = Consumer({
    "bootstrap.servers": BROKER,
    "group.id": ALERT_GROUP_ID,
    "auto.offset.reset": "latest"
})

alert_consumer.subscribe([ALERT_NOTIFICATIONS_T])

device = "cuda" if torch.cuda.is_available() else "cpu"


def decode_header(headers, key):
    val = None
    if headers is not None:
        headers = dict(headers)
        val = headers.get(key)
    return val.decode("utf-8") if val else None


class FallDetectorModel(nn.Module):
    def __init__(self):
        super().__init__()
        self.lstm = nn.LSTM(input_size=51, hidden_size=128, num_layers=2, batch_first=True)
        self.fc = nn.Linear(128, 1)
    
    def forward(self, x):
        out, _ = self.lstm(x)
        out = out[:, -1, :]
        out = self.fc(out)
        return out.squeeze(-1)


class FightDetectorModel:

    def __init__(self, model_path, device=device):
        self.device = torch.device(device)
        

        model = i3d_r50(pretrained=False)
        in_features = model.blocks[-1].proj.in_features
        model.blocks[-1].proj = nn.Linear(in_features, 2)
        

        checkpoint = torch.load(model_path, map_location=self.device)
        model.load_state_dict(checkpoint['model_state_dict'])
        model.eval()
        model.to(self.device)
        
        self.model = model
        

        self.mean = torch.tensor([0.485, 0.456, 0.406]).view(3, 1, 1, 1).to(self.device)
        self.std = torch.tensor([0.229, 0.224, 0.225]).view(3, 1, 1, 1).to(self.device)
    
    def predict(self, frames_array):
        """
        frames_array: numpy array shape (N, H, W, 3), uint8
        Returns: fight probability (float)
        """
        # Sample 32 evenly-spaced frames if more than 32
        n = len(frames_array)
        if n < 32:

            indices = np.linspace(0, n-1, 32).astype(int).tolist()
        else:
            indices = np.linspace(0, n-1, 32).astype(int).tolist()
        

        selected = []
        for idx in indices:
            frame = frames_array[idx]
            if frame.shape[:2] != (224, 224):
                frame = cv2.resize(frame, (224, 224))
            selected.append(frame)
        
        clip = np.array(selected, dtype=np.uint8)  # (32, 224, 224, 3)
        

        clip = torch.from_numpy(clip).float().to(self.device) / 255.0
        clip = clip.permute(3, 0, 1, 2)  # (C, T, H, W)
        clip = (clip - self.mean) / self.std
        clip = clip.unsqueeze(0)  # batch dim
        
        with torch.no_grad():
            logits = self.model(clip)
            probs = torch.softmax(logits, dim=1)
            fight_prob = probs[0, 1].item()
        
        return fight_prob


# ─── Frame processing function ────────────────────────────────────────────────────
detection_model = YOLO(DETECTION_MODEL).to(device)
pose_model= YOLO(POSE_MODEL).to(device)
weapon_model=YOLO(WEAPON_MODEL).to(device)


fall_model = FallDetectorModel().to(device)
checkpoint = torch.load(FALL_MODEL, map_location=device)
if isinstance(checkpoint, dict) and "model_state_dict" in checkpoint:
    fall_state_dict = checkpoint["model_state_dict"]
else:
    fall_state_dict = checkpoint

fall_model.load_state_dict(fall_state_dict)
fall_model.eval()

print(f"Loaded fall detection model from {FALL_MODEL}")

fight_model = FightDetectorModel(FIGHT_MODEL, device=device)
print(f"Loaded fight detection model from {FIGHT_MODEL}")


tracker = sv.ByteTrack()
url: str = os.environ.get("SUPABASE_URL")
key: str = os.environ.get("SUPABASE_KEY")
supabase: Client = create_client(url, key)

trackers = {}

keypoint_history = defaultdict(lambda: defaultdict(lambda: deque(maxlen=30)))
frame_buffers = defaultdict(lambda: deque())
last_inference_time = defaultdict(int)
EVIDENCE_WINDOW_MS = 4000 
evidence_buffers = defaultdict(lambda: deque())



def process_frame(frame_bytes, stream_id, timestamp):

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
    
    cv2.imwrite(f"/tmp/debug_frame_{stream_id}.jpg", img)
    
    ok, buf = cv2.imencode(".jpg", img)
    if not ok:
        print("could not encode the image")
    else:
        jpg_bytes = buf.tobytes()
        evidence_buffers[stream_id].append((timestamp, jpg_bytes))

    while evidence_buffers[stream_id] and evidence_buffers[stream_id][0][0] < timestamp - EVIDENCE_WINDOW_MS:
        evidence_buffers[stream_id].popleft()

    
    

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
    norm_keypoints = []

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
                box_x_min, box_y_min, box_x_max, box_y_max = pose_box
                
                box_width = box_x_max - box_x_min
                box_height = box_y_max - box_y_min

                if box_width == 0 or box_height == 0:
                    continue

                x_norm = (person_xy[:, 0] - box_x_min) / box_width
                y_norm = (person_xy[:, 1] - box_y_min) / box_height

                norm_keypoints.append({
                    "tracker_id": best_tracker_id,
                    "x_norm": x_norm,
                    "y_norm": y_norm,
                    "conf": person_conf
                })

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
    
    for kp_entry in norm_keypoints:
        
        tracker_id = kp_entry['tracker_id']
        x_array = np.array(kp_entry['x_norm']) 
        y_array = np.array(kp_entry['y_norm']) 
        keypoint_conf = np.array(kp_entry['conf'])
        
        kyp_arr = (x_array, y_array, keypoint_conf)
        kyp_norm = np.column_stack(kyp_arr)

        keypoint_history[stream_id][tracker_id].append(kyp_norm)

        # print("keypoints:shape:", kyp_norm.shape)   

    fall_predictions = {}

    current_trackers = {kp['tracker_id'] for kp in keypoints}

    for tracker_id, history in keypoint_history[stream_id].items():

        if tracker_id not in current_trackers:
            continue

        if len(history) < 30:
            continue  
        
        sequence = np.array(list(history))  
        sequence = sequence.reshape(30, -1) 
        sequence_tensor = torch.FloatTensor(sequence).unsqueeze(0).to(device)
        
        with torch.no_grad():
            logits = fall_model(sequence_tensor)
            prob = torch.sigmoid(logits).item()
        
        fall_predictions[str(tracker_id)] = prob

    small_frame = cv2.resize(img, (224, 224))

    window_ms = 15000
    MIN_FRAMES = 16

    while frame_buffers[stream_id] and frame_buffers[stream_id][0][0] < timestamp - window_ms:
        frame_buffers[stream_id].popleft()

    frame_buffers[stream_id].append((timestamp, small_frame))
    print(f"BUFFER: {stream_id} len={len(frame_buffers[stream_id])} ts={timestamp} since_last_inf={timestamp - last_inference_time[stream_id]}")

    fight_predictions = None
    if timestamp - last_inference_time[stream_id] >= 2000 and len(frame_buffers[stream_id]) >= MIN_FRAMES:
        frames_only = [f for (t, f) in frame_buffers[stream_id]]
        buffer_array = np.array(frames_only)
        fight_predictions = fight_model.predict(buffer_array)
        last_inference_time[stream_id] = timestamp
        print(f"FIGHT INFO: {stream_id} buf={len(buffer_array)} pred={fight_predictions:.4f}")

    return {
        "detections": list_det,
        "keypoints": keypoints,
        "fall_predictions": fall_predictions,
        "fight_predictions": fight_predictions,
        "latency_ms": latency
    }

def nearest_frame(stream_id, timestamp, tolerance_ms):

    buffer = evidence_buffers[stream_id]

    if not buffer:
        return None
    
    best_entry = None
    best_diff = float("inf")

    for (ts, jpg_bytes) in buffer:
        diff = abs(ts - timestamp)
        if diff < best_diff:
            best_diff = diff
            best_entry = (ts, jpg_bytes)

    if best_diff > tolerance_ms:
        return None
    
    return best_entry


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
        msg = frame_consumer.poll(1.0)
        if msg is None:
            pass
        elif msg.error():
            if msg.error().code() != KafkaError._PARTITION_EOF:
                print(f"Consumer error: {msg.error()}")
        else:    
            job = msg.value()
            stream_id = msg.key().decode("utf-8")

            if stream_id not in trackers:
                trackers[stream_id] = sv.ByteTrack()

            timestamp = int(decode_header(msg.headers(), "timestamp"))
            res = process_frame(job, stream_id, timestamp)

            output = {
                "camera":     msg.key().decode("utf-8"),
                "detections": res["detections"],
                "keypoints":  res["keypoints"],
                "fall_predictions": res["fall_predictions"],
                "fight_predictions": res["fight_predictions"],
                "latency_ms": res["latency_ms"],
                "timestamp":  timestamp,
            }

            payload = json.dumps(output).encode("utf-8")
            producer.produce(RESULTS_T, payload)
            producer.flush()
                             
            print(f"{output['camera']} → detections={res['detections']} →  keypoints={res['keypoints']} → fall_predictions={res['fall_predictions']} → fight_predictions={res['fight_predictions']} → "
                f"latency={res['latency_ms']:.1f}ms")
            
        alert_msg = alert_consumer.poll(0.0)
        if alert_msg is None: 
            pass
        elif alert_msg.error():
            if alert_msg.error().code() != KafkaError._PARTITION_EOF:
                print(f"Alert consumer error: {alert_msg.error()}")
        else:
            data = json.loads(alert_msg.value().decode("utf-8"))
            stream_id = data["camera"]

            result = nearest_frame(stream_id, data["timestamp"], 500)
            
            if result is None:
                print(f"no frame image found for alert{data['alert_id']}")
            else:
                # os.makedirs("frames", exist_ok=True)
                # with open(filename, "wb") as f:
                #     f.write(result[1])

                filename = f"{data['alert_id']}.jpg"

                try:
                    response = supabase.storage.from_('Alert Frames').upload(
                        f"frame/{filename}", 
                        result[1],
                        file_options={
                        "content-type": "image/jpeg",
                        "upsert": False,
                        },
                    )
                    
                except Exception as e:
                    print("Upload failed:", e)


                # print(f"alert {data['alert_id']}: requested {data['timestamp']}, matched {result[0]}, diff {abs(result[0]-data['timestamp'])}ms, saved {filename}")

except KeyboardInterrupt:
    pass
finally:
    frame_consumer.close()
    alert_consumer.close()
    print("Worker shutting down.")

