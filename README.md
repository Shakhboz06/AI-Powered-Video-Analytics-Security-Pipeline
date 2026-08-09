# Video Analytics Security Pipeline

Real-time, AI-powered security intelligence that watches camera feeds and flags the moments that actually matter — a fall, a weapon, a fight, someone where they shouldn't be — and captures forensic evidence for each one.

Built solo as a working prototype, with a privacy-first architecture.

> **Live demo:** [securityanalytics.duckdns.org](https://securityanalytics.duckdns.org)
> Live analysis runs Mon–Fri, 10:00–16:00 (Berlin time). Outside these hours the pipeline returns no results.

---

## What it does

The pipeline analyses video — either a live camera feed or an uploaded clip — and detects:

- **Weapon detection** — pistols, rifles, knives (custom-trained model)
- **Fall detection** — a person collapsing (ML sequence model on body keypoints)
- **Fighting** — physical altercation between people (3D action-recognition model)
- **Zone intrusion** — a person entering an operator-defined restricted area
- **Running** — abnormal movement speed
- **Loitering** — a person lingering too long
- **Abandoned object** — an item left behind unattended

Every alert captures a clean evidence frame with the detection drawn on it, stored privately and served through short-lived signed URLs — so there's a forensic record without raw video ever leaving secure storage.

---

## Architecture

The system is built as independent services communicating through a message queue, so any single part can fail or restart without taking the rest down.

```
                       ┌─────────────────────────────────────────────┐
   Camera (RTSP)  ─────┤                                             │
   via edge bridge     │   Ingestor  →  Kafka (video.analysis)        │
                       │                     │                       │
   Video upload   ─────┤                     ▼                       │
                       │              GPU Worker (Python)             │
                       │        YOLO detection + tracking,            │
                       │        pose, weapon, fall, fight models      │
                       │                     │                       │
                       │        Kafka (video.results)                 │
                       │                     ▼                       │
                       │            Aggregator (Go)                   │
                       │     alert logic + zone/behaviour rules       │
                       │              │            │                  │
                       │              ▼            ▼                  │
                       │      TimescaleDB       Redis                 │
                       │              │            │                  │
                       │              ▼            ▼                  │
                       │      Dashboard API (Go/Gin)  →  Nuxt frontend│
                       └─────────────────────────────────────────────┘

   Evidence frames  →  Supabase private storage  →  signed URLs
   GPU worker is provisioned on-demand (Vast.ai) by a scheduler,
   Mon–Fri 10:00–16:00 Berlin time, then torn down to save cost.
```

**Key components**

- **GPU worker (Python)** — runs the full model stack on every frame: object detection + tracking, pose estimation, weapon detection, fall classification, and fight recognition.
- **Aggregator (Go)** — consumes detections, applies zone/behaviour rules, produces alerts, and emits alert notifications.
- **Dashboard API (Go/Gin)** — serves detections, alerts, zones, and signed evidence-frame URLs to the frontend.
- **Frontend (Nuxt 4)** — live dashboard, alert workflow, video-upload analysis, and evidence display with bounding-box overlays.
- **Message queue (Kafka)** — decouples ingestion, inference, and aggregation.
- **Storage** — TimescaleDB (time-series detections/alerts), Redis (caching/live streaming), Supabase (private evidence frames).
  
---

## Tech stack

- **ML / inference:** Python, PyTorch, Ultralytics YOLO (detection + ByteTrack), pose estimation, custom weapon model, LSTM fall classifier, I3D fight recognition
- **Services:** Go / Gin
- **Streaming & data:** Apache Kafka, TimescaleDB, Redis
- **Storage:** Supabase (private buckets, signed URLs)
- **Frontend:** Nuxt 4 / Vue
** Infra:** Docker & Docker Compose, OVH VPS, Kubernutes, GPU, APScheduler

---

## Engineering decisions worth noting

- **Privacy-first evidence handling.** Raw frames stay in private storage; only detection metadata and short-lived signed URLs are exposed. Clean, unannotated frames are stored for forensic integrity — annotation is presentation-only.
- **Message-queue architecture.** Ingestion, inference, and aggregation are separate services on Kafka, so the system degrades gracefully instead of failing as a whole.
- **Edge-first ingestion.** Cameras behind a network are read by a local bridge that pushes outward — no inbound ports opened on the customer network, which is both safer and how real deployments work.

---

## Status & limitations

This is a **working prototype**, not a production product.

- Live analysis is available **Mon–Fri, 10:00–16:00 (Berlin time)** — the GPU is scheduled to control cost. Outside these hours, uploads return no results.
- Detector accuracy is at prototype stage: models are trained largely on public datasets, so real-world false positives (and misses) still occur and are an ongoing tuning effort. This is the normal data-and-tuning long game of the field, not a limit of the architecture.
- Infrastructure is proven; accuracy is the roadmap.

---

## The story behind it

<!-- [Link your narrative post here once written — the "why I built this" story for non-technical readers] -->
Read the story of how and why I built this → [coming soon]
