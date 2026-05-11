#!/usr/bin/env bash
set -euo pipefail

# BangmodMonitor Agent — Linux installer
# Usage: curl -fsSL https://your-domain.com/install.sh | bash -s -- --token=<TOKEN> [--server=URL] [--region=th] [--interval=30]

TOKEN=""
SERVER="https://api.bangmodmonitor.com"
REGION="default"
INTERVAL=30

for arg in "$@"; do
  case $arg in
    --token=*)   TOKEN="${arg#*=}" ;;
    --server=*)  SERVER="${arg#*=}" ;;
    --region=*)  REGION="${arg#*=}" ;;
    --interval=*) INTERVAL="${arg#*=}" ;;
    *) echo "Unknown argument: $arg"; exit 1 ;;
  esac
done

if [ -z "$TOKEN" ]; then
  echo "Error: --token is required"
  echo "Usage: curl -fsSL $SERVER/install.sh | bash -s -- --token=<YOUR_TOKEN>"
  exit 1
fi

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)         ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

INSTALL_DIR="/usr/local/bin"
BINARY_NAME="bangmod-agent"
DOWNLOAD_URL="$SERVER/downloads/bangmod-agent-$OS-$ARCH"

echo "==> Installing BangmodMonitor Agent"
echo "    Server : $SERVER"
echo "    Region : $REGION"
echo "    Arch   : $OS/$ARCH"

# Download agent binary
echo "==> Downloading agent..."
curl -fsSL "$DOWNLOAD_URL" -o "/tmp/$BINARY_NAME"
chmod +x "/tmp/$BINARY_NAME"
mv "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

# Detect init system and install service
if command -v systemctl &>/dev/null; then
  cat > /etc/systemd/system/bangmod-agent.service <<EOF
[Unit]
Description=BangmodMonitor Agent
After=network-online.target
Wants=network-online.target

[Service]
Environment="AGENT_TOKEN=$TOKEN"
Environment="API_URL=$SERVER"
Environment="AGENT_REGION=$REGION"
Environment="AGENT_INTERVAL=$INTERVAL"
ExecStart=$INSTALL_DIR/$BINARY_NAME
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
  systemctl enable bangmod-agent
  systemctl restart bangmod-agent
  echo "==> Agent installed as systemd service"
  echo "    Status : systemctl status bangmod-agent"
  echo "    Logs   : journalctl -u bangmod-agent -f"

elif command -v rc-update &>/dev/null; then
  # Alpine/OpenRC
  cat > /etc/init.d/bangmod-agent <<EOF
#!/sbin/openrc-run
name="bangmod-agent"
command="$INSTALL_DIR/$BINARY_NAME"
command_background=true
pidfile="/run/bangmod-agent.pid"
export AGENT_TOKEN="$TOKEN"
export API_URL="$SERVER"
export AGENT_REGION="$REGION"
export AGENT_INTERVAL="$INTERVAL"
EOF
  chmod +x /etc/init.d/bangmod-agent
  rc-update add bangmod-agent default
  rc-service bangmod-agent start
  echo "==> Agent installed as OpenRC service"

else
  # Fallback: write env file and run in background
  cat > /etc/bangmod-agent.env <<EOF
AGENT_TOKEN=$TOKEN
API_URL=$SERVER
AGENT_REGION=$REGION
AGENT_INTERVAL=$INTERVAL
EOF
  echo "==> Binary installed to $INSTALL_DIR/$BINARY_NAME"
  echo "    Run manually: env \$(cat /etc/bangmod-agent.env | xargs) $INSTALL_DIR/$BINARY_NAME"
fi

echo ""
echo "==> BangmodMonitor Agent installed successfully!"
