FROM pytorch/pytorch:2.12.0-cuda12.6-cudnn9-runtime

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \ 
    python3-venv \
    libgl1 \
    libglib2.0-0 \
    ffmpeg && \
    rm -rf /var/lib/apt/lists/*

RUN python -m venv /venv
ENV PATH="/venv/bin:$PATH"

COPY requirement.txt .

RUN pip install --no-cache-dir -r requirement.txt

COPY worker-1.py .
COPY worker/models/ /app/models/

ENTRYPOINT ["python", "-u", "worker-1.py"]