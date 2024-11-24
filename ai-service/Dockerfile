# Stage 1: Builder
FROM python:3.10-slim AS builder

# Install dependencies
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

# Stage 2: Runner
FROM python:3.10-slim

# Set up working directory
WORKDIR /app

# Copy installed dependencies from builder
COPY --from=builder /usr/local/lib/python3.10/site-packages /usr/local/lib/python3.10/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# Copy application code
COPY . .

# Expose the application port
EXPOSE 7011

# Set the command to start the application with gunicorn
CMD ["gunicorn", "-w", "2", "-b", "0.0.0.0:7011", "main:app"]
