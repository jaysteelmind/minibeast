#!/bin/bash
# MiniBeast USB Agent - macOS Launcher
# Double-click this file to run MiniBeast

echo "========================================"
echo "   MiniBeast USB Agent - Starting"
echo "========================================"
echo ""

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
USB_ROOT="$SCRIPT_DIR/.."

cd "$USB_ROOT"

# Run MiniBeast (choose correct architecture)
echo "Running system analysis..."
if [ "$(uname -m)" = "arm64" ]; then
    ./bin/minibeast-darwin-arm64 --root "$USB_ROOT" --config "$USB_ROOT/config/default.yaml"
else
    ./bin/minibeast-darwin-amd64 --root "$USB_ROOT" --config "$USB_ROOT/config/default.yaml"
fi

# Show results
echo ""
echo "========================================"
echo "   Analysis Complete!"
echo "========================================"
echo ""
echo "Results saved to: $USB_ROOT/out/"
echo ""

# Open output directory
open "$USB_ROOT/out"

echo ""
echo "Press Enter to exit..."
read
