FROM python:3.13.5-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      libgl1-mesa-glx libglib2.0-0 && \
    rm -rf /var/lib/apt/lists/*


COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY ingestor.py upload_ingestor.py ./

ENTRYPOINT ["python", "-u", "ingestor.py"]
