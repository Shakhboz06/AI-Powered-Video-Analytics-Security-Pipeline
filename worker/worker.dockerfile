FROM python:3.13.5-slim
WORKDIR /app


RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      libgl1-mesa-glx \
      libglib2.0-0 \
      wget && \
    rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt



COPY models ./models

RUN mkdir -p models && \
    wget -O models/MobileNetSSD_deploy.prototxt \
      https://raw.githubusercontent.com/chuanqi305/MobileNet-SSD/master/deploy.prototxt && \
    wget -O models/MobileNetSSD_deploy.caffemodel \
      https://raw.githubusercontent.com/chuanqi305/MobileNet-SSD/master/mobilenet_iter_73000.caffemodel && \
    test -s models/MobileNetSSD_deploy.prototxt && \
    test -s models/MobileNetSSD_deploy.caffemodel

COPY worker.py .


ENTRYPOINT ["python", "-u", "worker.py"]
