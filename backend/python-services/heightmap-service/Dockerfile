FROM python:3.11-slim

RUN apt-get update && apt-get install -y \
    build-essential \
    g++ \
    gcc \
    libgl1 \
    libglib2.0-0 \
    curl \
    gdal-bin \
    libgdal-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .

ENV GDAL_VERSION=3.10.3

RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --no-cache-dir -r requirements.txt

COPY heightmap_service /app/heightmap_service

RUN mkdir -p /tmp/heightmap_processing

EXPOSE 8000

CMD ["python", "-m", "heightmap_service"]

