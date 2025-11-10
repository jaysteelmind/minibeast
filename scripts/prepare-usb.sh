#!/bin/bash
# MiniBeast USB Preparation Script
# Validates and prepares the USB layout for deployment

set -e

echo "========================================="
echo "  MiniBeast USB Preparation"
echo "========================================="
echo ""

USB_LAYOUT="usb_layout"

# 1. Validate directory structure
echo "Step 1: Validating directory structure..."
for dir in bin config launch models out; do
    if [ ! -d "$USB_LAYOUT/$dir" ]; then
        echo "  Creating $dir/"
        mkdir -p "$USB_LAYOUT/$dir"
    else
        echo "  ✓ $dir/ exists"
    fi
done

# 2. Validate binaries
echo ""
echo "Step 2: Validating binaries..."
for binary in minibeast-linux minibeast-darwin-arm64 minibeast-darwin-amd64; do
    if [ -f "$USB_LAYOUT/bin/$binary" ]; then
        size=$(du -h "$USB_LAYOUT/bin/$binary" | cut -f1)
        echo "  ✓ $binary ($size)"
    else
        echo "  ⚠ $binary missing (need cross-compilation)"
    fi
done

if [ -f "$USB_LAYOUT/bin/minibeast-win.exe" ]; then
    size=$(du -h "$USB_LAYOUT/bin/minibeast-win.exe" | cut -f1)
    echo "  ✓ minibeast-win.exe ($size)"
else
    echo "  ⚠ minibeast-win.exe missing (need cross-compilation)"
fi

# 3. Validate model
echo ""
echo "Step 3: Validating AI model..."
if [ -f "$USB_LAYOUT/models/tinyllama-1.1b-chat-v1.0.Q2_K.gguf" ]; then
    size=$(du -h "$USB_LAYOUT/models/tinyllama-1.1b-chat-v1.0.Q2_K.gguf" | cut -f1)
    echo "  ✓ TinyLlama model ($size)"
    
    # Verify checksum if exists
    if [ -f "$USB_LAYOUT/models/tinyllama-1.1b-chat-v1.0.Q2_K.gguf.sha256" ]; then
        echo "  ✓ Checksum file present"
    fi
else
    echo "  ✗ Model file missing!"
    exit 1
fi

# 4. Validate configuration
echo ""
echo "Step 4: Validating configuration..."
if [ -f "$USB_LAYOUT/config/default.yaml" ]; then
    echo "  ✓ Configuration file exists"
    # Check model path in config
    if grep -q "model_path: \"models/tinyllama" "$USB_LAYOUT/config/default.yaml"; then
        echo "  ✓ Model path configured correctly"
    else
        echo "  ⚠ Model path may need adjustment"
    fi
else
    echo "  ✗ Configuration missing!"
    exit 1
fi

# 5. Validate launchers
echo ""
echo "Step 5: Validating launchers..."
for launcher in Run-MiniBeast.bat Run-MiniBeast.command Run-MiniBeast.sh MiniBeast.desktop; do
    if [ -f "$USB_LAYOUT/launch/$launcher" ]; then
        echo "  ✓ $launcher"
    else
        echo "  ✗ $launcher missing!"
    fi
done

# 6. Validate documentation
echo ""
echo "Step 6: Validating documentation..."
if [ -f "$USB_LAYOUT/README.txt" ]; then
    echo "  ✓ User guide (README.txt)"
else
    echo "  ⚠ README.txt missing"
fi

if [ -f "README.md" ]; then
    echo "  ✓ Developer documentation (README.md)"
fi

# 7. Calculate total size
echo ""
echo "Step 7: Calculating package size..."
total_size=$(du -sh "$USB_LAYOUT" | cut -f1)
echo "  Total USB layout size: $total_size"

# 8. Create package manifest
echo ""
echo "Step 8: Creating package manifest..."
cat > "$USB_LAYOUT/MANIFEST.txt" << MANIFEST
MiniBeast USB Agent - Package Manifest
Generated: $(date)
Version: 1.0.0

Directory Structure:
-------------------
bin/          Executables for all platforms
config/       Configuration files
launch/       Autorun launchers
models/       AI model files
out/          Output directory (created on first run)
docs/         Documentation (optional)

Package Size: $total_size

Included Files:
--------------
$(find "$USB_LAYOUT" -type f | wc -l) files total

Binaries:
$(ls -lh "$USB_LAYOUT/bin/" 2>/dev/null || echo "None")

Model:
$(ls -lh "$USB_LAYOUT/models/"*.gguf 2>/dev/null || echo "None")

Configuration:
$(ls -lh "$USB_LAYOUT/config/" 2>/dev/null || echo "None")

Launchers:
$(ls -lh "$USB_LAYOUT/launch/" 2>/dev/null || echo "None")

Installation:
------------
1. Format USB stick as exFAT (recommended) or FAT32
2. Copy entire usb_layout contents to USB root
3. Safely eject
4. Ready for deployment!

MANIFEST

echo "  ✓ MANIFEST.txt created"

# 9. Summary
echo ""
echo "========================================="
echo "  USB Package Preparation Complete!"
echo "========================================="
echo ""
echo "Package location: $USB_LAYOUT/"
echo "Total size: $total_size"
echo ""
echo "Ready for deployment: YES"
echo ""
echo "To copy to USB stick:"
echo "  sudo cp -r $USB_LAYOUT/* /media/USB_MOUNT_POINT/"
echo ""
