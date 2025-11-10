#!/bin/bash
# MiniBeast USB Agent - Linux Launcher
# Run this file to execute MiniBeast

echo "========================================"
echo "   MiniBeast USB Agent - Starting"
echo "========================================"
echo ""

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
USB_ROOT="$SCRIPT_DIR/.."

cd "$USB_ROOT"

# Run MiniBeast
echo "Running system analysis..."
./bin/minibeast-linux --root "$USB_ROOT" --config "$USB_ROOT/config/default.yaml"

# Show results
echo ""
echo "========================================"
echo "   Analysis Complete!"
echo "========================================"
echo ""
echo "Results saved to: $USB_ROOT/out/"
echo ""

# Try to open file manager to output directory
if command -v xdg-open > /dev/null; then
    xdg-open "$USB_ROOT/out"
elif command -v nautilus > /dev/null; then
    nautilus "$USB_ROOT/out"
fi

echo ""
echo "Press Enter to exit..."
read
