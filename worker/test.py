import cv2
import torch
import torch.nn as nn
import numpy as np
from ultralytics import YOLO
from collections import defaultdict, deque
import os

class FallDetectorModel(nn.Module):
    def __init__(self):
        super().__init__()
        self.lstm = nn.LSTM(input_size=34, hidden_size=64, batch_first=True)
        self.fc = nn.Linear(64, 1)
    
    def forward(self, x):
        out, _ = self.lstm(x)
        out = out[:, -1, :]
        out = self.fc(out)
        return out.squeeze(-1)

pose_model = YOLO("yolo11n-pose.pt")
fall_model = FallDetectorModel()
checkpoint = torch.load("models/fall_classifier_v1.pt")
fall_model.load_state_dict(checkpoint['model_state_dict'])
fall_model.eval()

TEST_VIDEO = "path/to/test"
OUTPUT_DIR = "fall_detections"
os.makedirs(OUTPUT_DIR, exist_ok=True)

cap = cv2.VideoCapture(TEST_VIDEO)

keypoint_history = defaultdict(lambda: deque(maxlen=30))
frame_idx = 0

THRESHOLD = 0.7

while True:
    ret, frame = cap.read()
    if not ret:
        break
    
    h, w = frame.shape[:2]
    results = pose_model.predict(frame, conf=0.15, verbose=False)
    
    detections_this_frame = []  # (person_idx, bbox, prob)
    
    if results[0].keypoints is not None and len(results[0].keypoints.xy) > 0:
        keypoints_xy = results[0].keypoints.xy.cpu().numpy()  # (num_persons, 17, 2)
        boxes = results[0].boxes.xyxy.cpu().numpy()  # (num_persons, 4)
        
        for person_idx in range(len(keypoints_xy)):
            xy = keypoints_xy[person_idx].copy()
            xy[:, 0] /= w
            xy[:, 1] /= h
            
            keypoint_history[person_idx].append(xy)
            
            if len(keypoint_history[person_idx]) == 30:
                sequence = np.array(list(keypoint_history[person_idx])).reshape(30, -1)
                sequence_tensor = torch.FloatTensor(sequence).unsqueeze(0)
                
                with torch.no_grad():
                    prob = torch.sigmoid(fall_model(sequence_tensor)).item()
                
                bbox = boxes[person_idx]
                detections_this_frame.append((person_idx, bbox, prob))
    
    for person_idx, bbox, prob in detections_this_frame:
        x1, y1, x2, y2 = bbox.astype(int)
        
        if prob > THRESHOLD:
            color = (0, 0, 255)  # red - fall
            label = f"FALL {prob:.2f}"
            thickness = 3
        else:
            color = (0, 255, 0)  # green - ok
            label = f"OK {prob:.2f}"
            thickness = 1
        
        cv2.rectangle(frame, (x1, y1), (x2, y2), color, thickness)
        cv2.putText(frame, label, (x1, y1-10), cv2.FONT_HERSHEY_SIMPLEX, 0.6, color, 2)
    
    any_fall = any(prob > THRESHOLD for _, _, prob in detections_this_frame)
    if any_fall:
        cv2.imwrite(f"{OUTPUT_DIR}/frame_{frame_idx:04d}.jpg", frame)
        print(f"Frame {frame_idx}: FALL detected")
    
    frame_idx += 1

cap.release()
