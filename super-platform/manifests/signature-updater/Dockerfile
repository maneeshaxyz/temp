FROM python:3.10.19-slim-trixie

# Install dependencies
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install cvdupdate
RUN pip3 install --no-cache-dir cvdupdate, curl

# Create database directory
RUN mkdir -p /clamav

# Configure cvd
RUN cvd config set --dbdir /clamav

# Copy update script
COPY update.sh /update.sh
RUN chmod +x /update.sh

# Run initial update on startup, then loop
CMD ["/update.sh"]