import cv2
import torch
import torch.nn as nn
import numpy as np
import os
from pytorchvideo.models.hub import i3d_r50


def create_fight_model():
    model = i3d_r50(pretrained=False)
    in_features = model.blocks[-1].proj.in_features
    model.blocks[-1].proj = nn.Linear(in_features, 2)
    return model


def extract_clip_evenly_spaced(video_path, num_samples=32, size=224):
    cap = cv2.VideoCapture(video_path)
    total = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))

    if total <= num_samples:
        indices = list(range(total)) + [total - 1] * (num_samples - total)
    else:
        indices = np.linspace(0, total - 1, num_samples).astype(int).tolist()

    target_dict = {}
    for pos, idx in enumerate(indices):
        target_dict.setdefault(idx, []).append(pos)

    output = [None] * num_samples
    frame_idx = 0
    while True:
        ret, frame = cap.read()
        if not ret:
            break
        if frame_idx in target_dict:
            resized = cv2.resize(frame, (size, size))
            for pos in target_dict[frame_idx]:
                output[pos] = resized
        frame_idx += 1
        if frame_idx > max(indices):
            break

    cap.release()
    return np.array(output, dtype=np.uint8)


MODEL_PATH = "models/fight_detector_v2.pt"
TEST_VIDEO = "C:/Users/shakh/OneDrive/Desktop/video-analytics-pipeline/dev/videos/yt_bankmt-Fight.mp4"

THRESHOLD = 0.6

# Load model
fight_model = create_fight_model()
checkpoint = torch.load(MODEL_PATH, map_location='cpu')
fight_model.load_state_dict(checkpoint['model_state_dict'])
fight_model.eval()
print(f"Loaded fight model (val_acc: {checkpoint['val_acc']:.4f})")

# Extract clip same way as training
clip_array = extract_clip_evenly_spaced(TEST_VIDEO)
print(f"Clip shape: {clip_array.shape}")

# Preprocess same as Dataset class
clip = torch.from_numpy(clip_array).float() / 255.0
clip = clip.permute(3, 0, 1, 2)
mean = torch.tensor([0.485, 0.456, 0.406]).view(3, 1, 1, 1)
std = torch.tensor([0.229, 0.224, 0.225]).view(3, 1, 1, 1)
clip = (clip - mean) / std
clip = clip.unsqueeze(0)

# Inference
with torch.no_grad():
    output = fight_model(clip)
    probs = torch.softmax(output, dim=1)

fight_prob = probs[0, 1].item()
nonfight_prob = probs[0, 0].item()

print(f"\nResults for: {TEST_VIDEO}")
print(f"NonFight probability: {nonfight_prob:.4f}")
print(f"Fight probability:    {fight_prob:.4f}")

if fight_prob > THRESHOLD:
    print(f"🚨 FIGHT DETECTED (confidence: {fight_prob:.2f})")
else:
    print(f"✓ No fight detected")