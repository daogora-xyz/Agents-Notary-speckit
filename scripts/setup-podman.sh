#!/usr/bin/env bash
# Initialize Podman for rootless operation
# This script sets up Podman to work without a daemon

set -e

echo "🐳 Setting up Podman (daemonless container runtime)..."
echo ""

# Check if Podman is installed
if ! command -v podman &> /dev/null; then
    echo "❌ Error: Podman not found"
    echo "Run this inside 'nix develop' to get Podman"
    exit 1
fi

echo "✅ Podman found: $(podman --version)"

# Initialize Podman machine (if on macOS)
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo ""
    echo "📱 Detected macOS - initializing Podman machine..."

    if ! podman machine list | grep -q "running"; then
        echo "Creating and starting Podman machine..."
        podman machine init 2>/dev/null || echo "Machine already exists"
        podman machine start || echo "Machine already running"
    else
        echo "✅ Podman machine already running"
    fi
fi

# Create Podman socket directory for rootless operation
SOCKET_DIR="/run/user/$(id -u)/podman"
if [ ! -d "$SOCKET_DIR" ]; then
    echo ""
    echo "Creating Podman socket directory: $SOCKET_DIR"
    mkdir -p "$SOCKET_DIR"
fi

# Start Podman socket service (rootless)
echo ""
echo "Starting Podman socket service (rootless)..."
if ! podman info --format '{{.Host.RemoteSocket.Path}}' &>/dev/null; then
    echo "Starting podman system service..."
    podman system service --time=0 unix://$SOCKET_DIR/podman.sock &
    sleep 2
    echo "✅ Podman socket service started"
else
    echo "✅ Podman socket already active"
fi

# Test Podman
echo ""
echo "Testing Podman..."
if podman info >/dev/null 2>&1; then
    echo "✅ Podman is working correctly"

    # Show Podman info
    echo ""
    echo "Podman Configuration:"
    echo "  Version: $(podman --version)"
    echo "  Storage Driver: $(podman info --format '{{.Store.GraphDriverName}}')"
    echo "  Root: $(podman info --format '{{.Store.GraphRoot}}')"
    echo "  Runroot: $(podman info --format '{{.Store.RunRoot}}')"
    echo "  Socket: unix://$SOCKET_DIR/podman.sock"
else
    echo "⚠️  Podman test failed"
    echo "Try running: podman info"
    exit 1
fi

# Configure Docker compatibility
echo ""
echo "Setting up Docker compatibility..."
export DOCKER_HOST="unix://$SOCKET_DIR/podman.sock"
echo "✅ Set DOCKER_HOST=$DOCKER_HOST"
echo ""
echo "💡 Docker commands will now use Podman backend"
echo "   Example: docker ps → uses Podman"

echo ""
echo "🎉 Podman setup complete!"
echo ""
echo "Next steps:"
echo "  1. Use 'podman-compose up -d' to start services"
echo "  2. Or use 'make up' (automatically detects Podman)"
echo "  3. Podman works without root and without a daemon!"
echo ""
