# MiniBeast Deployment Checklist

## Pre-Deployment

- [x] Phase 1 tested and validated
- [x] Phase 2 tested and validated (Linux)
- [x] Phase 3 integration complete (Linux)
- [x] All binaries compiled
- [x] USB package structure created
- [x] Documentation complete
- [x] Launchers created and tested

## USB Preparation

- [ ] Format USB stick (512MB minimum, 1GB recommended)
  - Recommended: exFAT (cross-platform)
  - Alternative: FAT32 (wider compatibility)
- [ ] Copy usb_layout/* to USB root
- [ ] Verify all files copied correctly (481MB total)
- [ ] Test launchers on USB stick
- [ ] Safely eject USB

## Testing

### Linux Testing
- [ ] Plug USB into Linux system
- [ ] Run: `./launch/Run-MiniBeast.sh`
- [ ] Verify Phase 1 output (JSON + signature)
- [ ] Verify Phase 2 output (report.txt)
- [ ] Check execution time (should be <200ms)
- [ ] Verify cryptographic keys created

### Windows Testing (Phase 1)
- [ ] Plug USB into Windows system
- [ ] Run: `launch\Run-MiniBeast.bat`
- [ ] Verify Phase 1 output (JSON + signature)
- [ ] Note: No AI report (mock template only)
- [ ] Check execution time (should be <50ms)

### macOS Testing (Phase 1)
- [ ] Plug USB into macOS system
- [ ] Run: `./launch/Run-MiniBeast.command`
- [ ] Allow execution if prompted by Gatekeeper
- [ ] Verify Phase 1 output (JSON + signature)
- [ ] Note: No AI report (mock template only)
- [ ] Check execution time (should be <50ms)

## Deployment

### Single System Deployment
- [ ] Insert USB stick
- [ ] Run appropriate launcher
- [ ] Wait for completion
- [ ] Retrieve results from `out/` folder
- [ ] Safely eject USB

### Batch Deployment
- [ ] Create USB master image
- [ ] Replicate to multiple USB sticks
- [ ] Label each USB stick
- [ ] Distribute to target systems
- [ ] Collect results centrally

## Post-Deployment

### Verification
- [ ] Check all systems generated output
- [ ] Verify signatures are valid
- [ ] Review AI reports (Linux systems)
- [ ] Compile results centrally

### Reporting
- [ ] Aggregate JSON data
- [ ] Review AI-generated reports
- [ ] Identify common issues/risks
- [ ] Generate executive summary

### Maintenance
- [ ] Store USB sticks securely
- [ ] Back up cryptographic keys
- [ ] Document any issues encountered
- [ ] Plan next deployment cycle

## Troubleshooting

### Common Issues

**Issue:** Permission denied
- **Solution:** Run with sudo/admin privileges
- **Or:** Accept "unknown" values for restricted data

**Issue:** Model not found (Linux)
- **Solution:** Verify models/*.gguf exists on USB
- **Check:** File size should be 461MB

**Issue:** Very slow execution
- **Cause:** USB 2.0 vs USB 3.0 speed difference
- **Solution:** Use USB 3.0 port if available

**Issue:** Windows Defender blocks execution
- **Solution:** Add exclusion or sign binary
- **Alternative:** Run from trusted location

**Issue:** macOS Gatekeeper blocks
- **Solution:** Right-click → Open → Open Anyway
- **Or:** Disable Gatekeeper temporarily

## Success Criteria

- [ ] All target systems scanned
- [ ] All reports generated successfully
- [ ] No data loss or corruption
- [ ] Execution time within expectations
- [ ] No security incidents
- [ ] Results collected centrally

## Completion

Date Deployed: _______________
Systems Scanned: _____________
Success Rate: ________________
Issues Encountered: __________
Next Deployment: _____________

**Deployment Status:** [ ] Complete

---

*MiniBeast USB Agent v1.0.0*  
*Deployment Checklist*
