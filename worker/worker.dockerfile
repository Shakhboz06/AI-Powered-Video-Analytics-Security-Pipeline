FROM python:3.13.5-slim
WORKDIR /app


RUN apt-get update && \
  apt-get install -y --no-install-recommends \
  libgl1-mesa-glx \
  libglib2.0-0 && \
  rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

# RUN python -c "from ultralytics import YOLO; YOLO('yolo26s.pt')"

COPY worker.py .
COPY models/ /app/models/


ENTRYPOINT ["python", "-u", "worker.py"]
