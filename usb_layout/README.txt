========================================
    MINIBEAST USB AGENT v1.0.0
    System Inventory & Analysis Tool
========================================

QUICK START:
-----------
1. Double-click the launcher for your operating system:
   - Windows: launch\Run-MiniBeast.bat
   - macOS:   launch/Run-MiniBeast.command
   - Linux:   launch/Run-MiniBeast.sh

2. Wait for analysis to complete (usually <1 second)

3. View results in the 'out' folder:
   - *.json = System data (technical details)
   - *.report.txt = Human-readable analysis
   - *.sig = Cryptographic signature

WHAT IT DOES:
------------
MiniBeast collects essential system information:
- Operating system details
- Hardware identifiers
- Network configuration
- User accounts (local only)

Then generates an AI-powered security analysis report.

PRIVACY & SECURITY:
------------------
✓ No passwords collected
✓ No browser data accessed
✓ No documents scanned
✓ All data stays on USB stick
✓ Cryptographically signed for integrity
✓ Works completely offline

OUTPUT FILES:
------------
out/<hostname>_<uuid>_<timestamp>.json
  - Complete system facts in JSON format
  
out/<hostname>_<uuid>_<timestamp>.json.sig
  - Digital signature (Ed25519, 64 bytes)
  
out/<hostname>_<uuid>_<timestamp>.report.txt
  - AI-generated analysis report

out/minibeast.key
  - Private signing key (keep secure!)
  
out/REPORTING_PUBKEY.txt
  - Public key for signature verification

SYSTEM REQUIREMENTS:
-------------------
- Windows 10 or later
- macOS 12 (Monterey) or later
- Linux with kernel 5.0+
- 4GB RAM minimum
- No installation required
- No administrator privileges required

PERFORMANCE:
-----------
- Typical execution: 100-200ms
- Maximum execution: 5 seconds
- Model size: 461MB (included on USB)
- Output size: ~10KB per run

TROUBLESHOOTING:
---------------
Problem: "Permission denied" errors
Solution: Some system data requires admin/root access.
          Run with elevated privileges or accept "unknown" values.

Problem: Launcher doesn't run
Solution: Right-click → Properties → Unblock (Windows)
          Right-click → Open (macOS first run)
          chmod +x launch/*.sh (Linux)

Problem: Model not found
Solution: Ensure 'models' folder contains tinyllama*.gguf file

Problem: Very slow on first run
Solution: First run loads AI model (~1 second), subsequent runs are faster

SUPPORT:
-------
For questions or issues, refer to the technical documentation
in the 'docs' folder or contact your IT administrator.

========================================
Version: 1.0.0
Build Date: 2025-11-09
License: Proprietary
========================================
