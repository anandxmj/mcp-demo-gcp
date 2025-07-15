# Multi-stage build for Flight Ticket Tools MCP Server using uv
# Stage 1: Build stage with uv
FROM python:3.13-slim as builder

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /usr/local/bin/

# Set environment variables
ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

# Set working directory
WORKDIR /app

# Copy uv configuration files
COPY pyproject.toml uv.lock ./

# Install dependencies into the system using uv
RUN uv sync --frozen --no-install-project --no-dev

# Copy application code
COPY main.py ./

# Stage 2: Runtime stage with Python slim (not distroless to have uv available)
FROM python:3.13-slim

# Set environment variables for Cloud Run
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    ENVIRONMENT=cloudrun \
    PORT=8080

# Install uv in the runtime stage
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /usr/local/bin/

# Copy the project files and virtual environment from builder
COPY --from=builder --chown=nonroot:nonroot /app /app

# Set working directory
WORKDIR /app

# Create a non-root user for security
RUN groupadd --gid 65532 nonroot && \
    useradd --uid 65532 --gid 65532 --shell /bin/bash --create-home nonroot

# Change ownership to nonroot user
RUN chown -R nonroot:nonroot /app

# Switch to non-root user
USER nonroot

# Expose the port that Cloud Run expects
EXPOSE 8080

# Use uv to run the application without installing as a project
ENTRYPOINT ["uv", "run", "--no-project", "main.py"]
