FROM python:3.14-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt \
    && groupadd --system appuser \
    && useradd --system --gid appuser --home-dir /app --shell /usr/sbin/nologin appuser \
    && chown -R appuser:appuser /app
COPY --chown=appuser:appuser server.py .
USER appuser:appuser
EXPOSE 3000
CMD ["python", "server.py"]
