#!/bin/bash

# Configuration
USER_NAME="ubuntu"
GROUP_NAME="ubuntu"
APP_NAME="fullstackapp"
APP_DIR="/home/${USER_NAME}/${APP_NAME}"
EXEC_PATH="${APP_DIR}/server"
ENV_FILE="${APP_DIR}/.env"

# Create the service file content
SERVICE_CONTENT="[Unit]
Description=FullStackApp Go Backend
After=network.target

[Service]
User=${USER_NAME}
Group=${GROUP_NAME}
WorkingDirectory=${APP_DIR}
ExecStart=${EXEC_PATH}
Restart=always
EnvironmentFile=${ENV_FILE}

[Install]
WantedBy=multi-user.target
"

# Path to the systemd service file
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"

# Write the service file
echo "Creating systemd service file at ${SERVICE_FILE}..."
sudo bash -c "echo '${SERVICE_CONTENT}' > ${SERVICE_FILE}"

# Reload systemd daemon
echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

# Enable the service
echo "Enabling ${APP_NAME} service..."
sudo systemctl enable ${APP_NAME}

# Start the service
echo "Starting ${APP_NAME} service..."
sudo systemctl start ${APP_NAME}

# Show status
echo "Service status:"
sudo systemctl status ${APP_NAME} --no-pager
