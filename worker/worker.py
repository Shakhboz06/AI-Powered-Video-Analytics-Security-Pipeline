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

# ─── Load environment ──────────────────────────────────────────────────────────────
load_dotenv(os.path.join(os.path.dirname(__file__), "../.env"))
BROKER       = os.getenv("KAFKA_BROKER", "kafka:9092")
ANALYSIS_T   = os.getenv("KAFKA_ANALYSIS_TOPIC",  "video.analysis")
RESULTS_T    = os.getenv("KAFKA_RESULTS_TOPIC",   "video.results")
GROUP_ID     = os.getenv("WORKER_GROUP_ID",       "worker-group")
DETECTION_MODEL = os.getenv("DETECTION_MODEL_PATH", "models/yolo26s.pt")
WEAPON_MODEL = os.getenv("WEAPON_MODEL_PATH", "models/weapons.pt")
FALL_MODEL = os.getenv("FALL_MODEL_PATH", "models/fall_classifier_v1.pt")
FIGHT_MODEL = os.getenv("FIGHT_MODEL_PATH", "models/fight_detector_v2.pt")
POSE_MODEL= os.getenv("POSE_MODEL_PATH", "/models/yolo11n-pose.pt")
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

class FallDetectorModel(nn.Module):
    def __init__(self):
        super().__init__()
        self.lstm = nn.LSTM(input_size=51, hidden_size=64, batch_first=True)
        self.fc = nn.Linear(64, 1)
    
    def forward(self, x):
        out, _ = self.lstm(x)
        out = out[:, -1, :]
        out = self.fc(out)
        return out.squeeze(-1)


class FightDetectorModel:

    def __init__(self, model_path, device='cpu'):
        self.device = torch.device(device)
        
        # Build architecture
        model = i3d_r50(pretrained=False)
        in_features = model.blocks[-1].proj.in_features
        model.blocks[-1].proj = nn.Linear(in_features, 2)
        
        # Load trained weights
        checkpoint = torch.load(model_path, map_location=self.device)
        model.load_state_dict(checkpoint['model_state_dict'])
        model.eval()
        model.to(self.device)
        
        self.model = model
        
        # Normalization tensors (precomputed)
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
            # Pad with last frame
            indices = np.linspace(0, n-1, 32).astype(int).tolist()
        else:
            indices = np.linspace(0, n-1, 32).astype(int).tolist()
        
        # Take selected frames and resize
        selected = []
        for idx in indices:
            frame = frames_array[idx]
            if frame.shape[:2] != (224, 224):
                frame = cv2.resize(frame, (224, 224))
            selected.append(frame)
        
        clip = np.array(selected, dtype=np.uint8)  # (32, 224, 224, 3)
        
        # To tensor
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
detection_model = YOLO(DETECTION_MODEL)
pose_model= YOLO(POSE_MODEL)
weapon_model=YOLO(WEAPON_MODEL)

# Fall Classifier
fall_model = FallDetectorModel()
checkpoint = torch.load(FALL_MODEL)
fall_model.load_state_dict(checkpoint['model_state_dict'])
fall_model.eval()
print(f"Loaded fall detection model from {FALL_MODEL}")

fight_model = FightDetectorModel(FIGHT_MODEL, device='cpu')
print(f"Loaded fight detection model from {FIGHT_MODEL}")


tracker = sv.ByteTrack()
trackers = {}

keypoint_history = defaultdict(lambda: defaultdict(lambda: deque(maxlen=30)))
frame_buffers = defaultdict(lambda: deque())
last_inference_time = defaultdict(int)

# ─── Per-stream state sweep ───────────────────────────────────────────────
# Upload jobs create a stream per job_id; once a job finishes its state
# would linger forever (same leak as long-gone live cameras). Age-based
# sweep drops everything for streams silent longer than STATE_TTL_S.
last_seen = {}
STATE_TTL_S = 600
SWEEP_INTERVAL_S = 60
last_sweep = time.time()


def sweep_stale_streams():
    now = time.time()
    stale = [sid for sid, seen in last_seen.items() if now - seen > STATE_TTL_S]
    for sid in stale:
        last_seen.pop(sid, None)
        trackers.pop(sid, None)
        keypoint_history.pop(sid, None)
        frame_buffers.pop(sid, None)
        last_inference_time.pop(sid, None)
        print(f"🧹 swept per-stream state for idle stream {sid}")

def process_frame(frame_bytes, stream_id, timestamp):

    tracker = trackers[stream_id]
    arr = np.frombuffer(frame_bytes, dtype=np.uint8)
    img = cv2.imdecode(arr, cv2.IMREAD_COLOR)
    if img is None:
        return {"detections": [], "latency_ms": 0.0}

    print(f"🔍 decoded image shape: {img.shape}")

    start = time.time()
    results = detection_model.predict(img, conf=0.25)
    # results = detection_model.predict(img, conf=0.10)

    detections = sv.Detections.from_ultralytics(results[0])
    detections = tracker.update_with_detections(detections)
    
    cv2.imwrite(f"/tmp/debug_frame_{stream_id}.jpg", img)

    print(f"Shape: {img.shape}, dtype: {img.dtype}")
    print(f"Min/max: {img.min()}/{img.max()}")
    print(f"Mean per channel: {img.mean(axis=(0,1))}")

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

        print("keypoints:shape:", kyp_norm.shape)   
        

    fall_predictions = {}

    current_trackers = {kp['tracker_id'] for kp in keypoints}

    for tracker_id, history in keypoint_history[stream_id].items():

        if tracker_id not in current_trackers:
            continue

        if len(history) < 30:
            continue  
        
        
        sequence = np.array(list(history))  
        print("sequence:", sequence)
        sequence = sequence.reshape(30, -1) 
        sequence_tensor = torch.FloatTensor(sequence).unsqueeze(0)
        
        
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


    # if frame_counters[stream_id] % FIGHT_INFERENCE_INTERVAL == 0 and len(frame_buffers[stream_id]) >= 32:

    #     buffer_array = np.array(list(frame_buffers[stream_id]))

    #     os.makedirs(f"debug/{stream_id}", exist_ok=True)
    #     n_to_save = 5  # save 5 frames spread across the buffer
    #     indices = [0, len(buffer_array)//4, len(buffer_array)//2, 3*len(buffer_array)//4, len(buffer_array)-1]
    #     for i, idx in enumerate(indices):
    #         cv2.imwrite(f"debug/{stream_id}/inf{frame_counters[stream_id]}_frame{i}.jpg", buffer_array[idx])

    #     fight_predictions = fight_model.predict(buffer_array)
    #     print(f"INFERENCE: {stream_id} buffer_size={len(buffer_array)} prediction={fight_predictions}")

    return {"detections": list_det, "keypoints": keypoints, "fall_predictions": fall_predictions, "fight_predictions": fight_predictions, "latency_ms": latency}

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

        last_seen[stream_id] = time.time()

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

        if time.time() - last_sweep >= SWEEP_INTERVAL_S:
            sweep_stale_streams()
            last_sweep = time.time()

except KeyboardInterrupt:
    pass
finally:
    consumer.close()
    print("Worker shutting down.")
