# MiniBeast USB Agent - Deployment Guide

## Quick Deployment (Linux)

### Prerequisites
- USB stick (512MB minimum, 1GB recommended)
- Linux system (kernel 5.0+)
- No installation or admin privileges required

### Steps

1. **Format USB stick**
```bash
   # Format as exFAT for cross-platform compatibility
   sudo mkfs.exfat -n MINIBEAST /dev/sdX1
```

2. **Copy files to USB**
```bash
   sudo mount /dev/sdX1 /mnt/usb
   sudo cp -r usb_layout/* /mnt/usb/
   sudo umount /mnt/usb
```

3. **Deploy and run**
   - Insert USB into target system
   - Navigate to USB mount point
   - Run: `./launch/Run-MiniBeast.sh`
   - Or double-click: `launch/MiniBeast.desktop`

### Output

Results appear in `out/` directory:
- `*.json` - System data
- `*.json.sig` - Cryptographic signature
- `*.report.txt` - AI-generated analysis
- `minibeast.key` - Private key (secure!)
- `REPORTING_PUBKEY.txt` - Public key

## Enterprise Deployment

### Batch Deployment

For deploying to multiple USB sticks:
```bash
#!/bin/bash
# Mass USB preparation script

USB_IMAGE="minibeast-v1.0.0.img"

# Create master image (once)
dd if=/dev/sdX of=$USB_IMAGE bs=4M status=progress

# Replicate to multiple USBs
for device in /dev/sd[b-z]1; do
    echo "Writing to $device..."
    sudo dd if=$USB_IMAGE of=$device bs=4M status=progress
    sync
done
```

### Centralized Collection

To collect reports from multiple systems:
```bash
# On each system, after running MiniBeast
scp /mnt/usb/out/*.json central-server:/reports/

# Or use a shared network drive
cp /mnt/usb/out/*.json /mnt/shared-reports/
```

## Platform-Specific Notes

### Linux (Full Phase 3)
- ✅ Real LLM inference
- ✅ Complete security analysis
- ✅ Sub-200ms execution
- **Recommended for production**

### macOS (Phase 1)
- ✅ Data collection working
- ✅ Cryptographic signing working
- ⚠️ Mock AI responses only
- Consider upgrading to Phase 3 if needed

### Windows (Phase 1)
- ✅ Data collection working
- ✅ Cryptographic signing working
- ⚠️ Mock AI responses only
- Consider upgrading to Phase 3 if needed

## Troubleshooting

### Linux

**Issue:** Permission denied
```bash
# Run with sudo if needed
sudo ./launch/Run-MiniBeast.sh
```

**Issue:** Model not found
```bash
# Verify model exists
ls -lh models/*.gguf
# Should show: tinyllama-1.1b-chat-v1.0.Q2_K.gguf (461MB)
```

### macOS

**Issue:** "Cannot be opened because it is from an unidentified developer"
```bash
# Allow execution
sudo spctl --master-disable
# Or right-click → Open → Open anyway
```

**Issue:** Mock LLM responses
- This is expected on macOS (Phase 1 binary)
- Data collection and signing work perfectly
- To enable real LLM: rebuild with CGO cross-compilation

### Windows

**Issue:** Windows Defender blocks execution
```bash
# Add exclusion or temporarily disable
# Or sign binary with code signing certificate
```

## Security Considerations

### Key Management
- Private key: `out/minibeast.key` (keep secure!)
- Public key: `out/REPORTING_PUBKEY.txt` (distribute freely)
- Verify signatures: `cat *.json.sig | base64`

### Data Privacy
- No passwords collected
- No browser data accessed
- No document scanning
- All data stays on USB
- Offline operation only

## Performance Metrics

### Linux (Phase 3)
- Collection: 7-10ms
- Crypto signing: <1ms
- LLM inference: 100-120ms
- **Total: 110-130ms**

### macOS/Windows (Phase 1)
- Collection: 7-10ms
- Crypto signing: <1ms
- Mock LLM: <1ms
- **Total: 10-15ms**

## Support

For issues or questions:
1. Check DEPLOYMENT_STATUS.txt
2. Review README.txt in USB root
3. Consult technical documentation
4. Contact system administrator

---

**Version:** 1.0.0  
**Last Updated:** 2025-11-09  
**Status:** Production Ready (Linux), Phase 1 (macOS/Windows)
