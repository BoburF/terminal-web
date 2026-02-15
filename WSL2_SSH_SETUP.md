# WSL2 SSH Server External Connectivity Setup Guide

## Problem Overview

When running an SSH server inside WSL2 (Windows Subsystem for Linux), external computers cannot connect directly to it because:

1. **WSL2 runs in a virtual machine** with its own virtual network interface
2. **Windows and WSL2 are on different subnets** (Windows: `10.162.97.69`, WSL2: `172.26.176.1`)
3. **Windows doesn't automatically forward ports** to the WSL2 VM
4. **Windows Firewall blocks inbound connections** by default

This guide sets up **port forwarding** from Windows to WSL2 and configures the firewall to allow external SSH connections.

---

## Prerequisites

- Windows 10/11 with WSL2 installed
- Administrator access (PowerShell must be run as Administrator)
- SSH server running in WSL2 on port 4569
- PowerShell 5.0 or higher

---

## Quick Setup (One-Command Script)

Save this as `setup-wsl-ssh.ps1` and run as Administrator:

```powershell
#Requires -RunAsAdministrator

<#
.SYNOPSIS
    Sets up port forwarding from Windows to WSL2 for SSH server on port 4569
.DESCRIPTION
    Configures Windows to forward port 4569 to WSL2 VM and allows it through firewall
.NOTES
    Run this script as Administrator
#>

Write-Host "=== WSL2 SSH Server Port Forwarding Setup ===" -ForegroundColor Green
Write-Host ""

# Step 1: Get WSL2 IP address
Write-Host "Step 1: Getting WSL2 IP address..." -ForegroundColor Yellow
$wslIp = wsl hostname -I
$wslIp = $wslIp.Trim()

if ([string]::IsNullOrWhiteSpace($wslIp)) {
    Write-Error "Could not get WSL2 IP address. Is WSL running?"
    exit 1
}

Write-Host "  WSL2 IP: $wslIp" -ForegroundColor Cyan
Write-Host ""

# Step 2: Get Windows IP addresses
Write-Host "Step 2: Getting Windows IP addresses..." -ForegroundColor Yellow
$windowsIps = Get-NetIPAddress -AddressFamily IPv4 | Where-Object { 
    $_.IPAddress -notlike "127.*" -and 
    $_.IPAddress -notlike "169.254.*" 
} | Select-Object -ExpandProperty IPAddress

Write-Host "  Windows IPs:" -ForegroundColor Cyan
$windowsIps | ForEach-Object { Write-Host "    $_" -ForegroundColor Cyan }
Write-Host ""

# Step 3: Remove existing port forwarding rules
Write-Host "Step 3: Cleaning up existing port forwarding rules..." -ForegroundColor Yellow
netsh interface portproxy delete v4tov4 listenport=4569 listenaddress=0.0.0.0 2>$null
Write-Host "  ✓ Cleaned up existing rules" -ForegroundColor Green
Write-Host ""

# Step 4: Add port forwarding rule
Write-Host "Step 4: Adding port forwarding rule..." -ForegroundColor Yellow
$portProxyResult = netsh interface portproxy add v4tov4 listenport=4569 listenaddress=0.0.0.0 connectport=4569 connectaddress=$wslIp 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✓ Port forwarding configured" -ForegroundColor Green
    Write-Host "    Windows:0.0.0.0:4569 → WSL2:$wslIp:4569" -ForegroundColor Cyan
} else {
    Write-Error "Failed to add port forwarding: $portProxyResult"
    exit 1
}
Write-Host ""

# Step 5: Configure Windows Firewall
Write-Host "Step 5: Configuring Windows Firewall..." -ForegroundColor Yellow

# Remove existing rule if exists
netsh advfirewall firewall delete rule name="SSH Server WSL2 Port 4569" 2>$null | Out-Null

# Add new rule
$firewallResult = netsh advfirewall firewall add rule name="SSH Server WSL2 Port 4569" dir=in action=allow protocol=TCP localport=4569 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "  ✓ Firewall rule added" -ForegroundColor Green
} else {
    Write-Warning "Failed to add firewall rule: $firewallResult"
}
Write-Host ""

# Step 6: Verify configuration
Write-Host "Step 6: Verifying configuration..." -ForegroundColor Yellow
Write-Host ""
Write-Host "  Port forwarding rules:" -ForegroundColor Cyan
netsh interface portproxy show v4tov4 | Select-String "4569"
Write-Host ""

Write-Host "=== Setup Complete! ===" -ForegroundColor Green
Write-Host ""
Write-Host "Connection Information:" -ForegroundColor Yellow
Write-Host "  From other computers, connect to:" -ForegroundColor White
$windowsIps | ForEach-Object { 
    Write-Host "    ssh -p 4569 -o StrictHostKeyChecking=no $_" -ForegroundColor Cyan 
}
Write-Host ""
Write-Host "Testing:" -ForegroundColor Yellow
Write-Host "  From Windows PowerShell:" -ForegroundColor White
Write-Host "    ssh -p 4569 -o StrictHostKeyChecking=no localhost" -ForegroundColor Cyan
Write-Host ""
Write-Host "  From other computer:" -ForegroundColor White
Write-Host "    ssh -p 4569 -o StrictHostKeyChecking=no <Windows-IP>" -ForegroundColor Cyan
Write-Host ""
Write-Host "Note: The WSL2 IP ($wslIp) may change after WSL restarts." -ForegroundColor Yellow
Write-Host "      Run this script again if connections fail." -ForegroundColor Yellow
```

---

## Manual Step-by-Step Setup

If you prefer to run commands individually, follow these steps:

### Step 1: Open PowerShell as Administrator

1. Press `Win + X`
2. Click "Windows PowerShell (Admin)" or "Terminal (Admin)"
3. Confirm the UAC prompt

**What this does:** Administrative privileges are required to modify network configuration and firewall rules.

---

### Step 2: Get WSL2 IP Address

```powershell
wsl hostname -I
```

**Example output:**
```
172.26.176.1
```

**What this does:** 
- Retrieves the IP address of the WSL2 virtual machine
- WSL2 creates a virtual network adapter with its own IP in the `172.x.x.x` range
- This IP is different from your Windows IP and changes when WSL2 restarts

**Note:** Save this IP address - you'll need it for the port forwarding command.

---

### Step 3: Clean Up Existing Port Forwarding Rules

```powershell
netsh interface portproxy delete v4tov4 listenport=4569 listenaddress=0.0.0.0
```

**What this does:**
- Removes any existing port forwarding rules for port 4569
- Prevents conflicts with duplicate rules
- The command will show "Element not found" if no rule exists - this is normal

**Parameters explained:**
- `delete v4tov4` - Delete IPv4 to IPv4 port proxy rule
- `listenport=4569` - The port Windows will listen on
- `listenaddress=0.0.0.0` - Listen on all network interfaces

---

### Step 4: Add Port Forwarding Rule

```powershell
netsh interface portproxy add v4tov4 listenport=4569 listenaddress=0.0.0.0 connectport=4569 connectaddress=172.26.176.1
```

**Replace `172.26.176.1` with your actual WSL2 IP from Step 2.**

**What this does:**
- Creates a bridge between Windows network and WSL2 VM
- Forwards all traffic on Windows port 4569 to WSL2 port 4569
- Allows external computers to reach the SSH server inside WSL2

**Parameters explained:**
- `add v4tov4` - Add IPv4 to IPv4 port proxy rule
- `listenport=4569` - Port Windows listens on
- `listenaddress=0.0.0.0` - Accept connections on all interfaces (0.0.0.0 = all IPs)
- `connectport=4569` - Port on WSL2 to forward to
- `connectaddress=172.26.176.1` - WSL2's IP address (the target)

**Visual representation:**
```
External Computer                    Windows Host                     WSL2 VM
     |                                   |                               |
     |  ssh -p 4569 10.162.97.69        |                               |
     |---------------------------------->|                               |
     |                                   |                               |
     |                                   |  Forward to 172.26.176.1:4569 |
     |                                   |------------------------------>|
     |                                   |                               |
     |                                   |<------------------------------|
     |<----------------------------------|  Response                     |
     |                                   |                               |
```

---

### Step 5: Verify Port Forwarding

```powershell
netsh interface portproxy show v4tov4
```

**Example output:**
```
Listen on ipv4:             Connect to ipv4:

Address         Port        Address         Port
--------------- ----------  --------------- ----------
0.0.0.0         4569        172.26.176.1    4569
```

**What this shows:**
- Confirms the rule was created successfully
- Displays the mapping: `0.0.0.0:4569 → 172.26.176.1:4569`

---

### Step 6: Configure Windows Firewall

```powershell
netsh advfirewall firewall add rule name="SSH Server WSL2 Port 4569" dir=in action=allow protocol=TCP localport=4569
```

**What this does:**
- Creates a firewall rule allowing inbound TCP connections on port 4569
- Without this, Windows blocks external connections even with port forwarding

**Parameters explained:**
- `advfirewall firewall add rule` - Add new firewall rule
- `name="SSH Server WSL2 Port 4569"` - Descriptive name for the rule
- `dir=in` - Direction: inbound (external → Windows)
- `action=allow` - Allow the traffic (not block)
- `protocol=TCP` - TCP protocol (SSH uses TCP)
- `localport=4569` - Apply to port 4569 only

---

### Step 7: Verify Firewall Rule

```powershell
netsh advfirewall firewall show rule name="SSH Server WSL2 Port 4569"
```

**Example output:**
```
Rule Name:                            SSH Server WSL2 Port 4569
----------------------------------------------------------------------
Enabled:                              Yes
Direction:                            In
Profiles:                             Domain,Private,Public
Grouping:                             
LocalIP:                              Any
RemoteIP:                             Any
Protocol:                             TCP
LocalPort:                            4569
RemotePort:                           Any
Edge traversal:                       No
Action:                               Allow
```

**What this shows:**
- Confirms the firewall rule is active
- Shows it's allowing inbound TCP traffic on port 4569
- Applies to all network profiles (Domain, Private, Public)

---

## Testing the Setup

### Test 1: Verify Port Forwarding from Windows

From Windows PowerShell (doesn't need to be Administrator):

```powershell
ssh -p 4569 -o StrictHostKeyChecking=no localhost
```

**Expected result:** Connection succeeds and you see the SSH server's TUI/resume

**What this tests:**
- Port forwarding is working correctly
- Windows can reach WSL2 through the forwarded port
- SSH server is accepting connections

---

### Test 2: Get Windows IP Address

From Windows PowerShell:

```powershell
ipconfig | findstr "IPv4"
```

Look for your Wi-Fi or Ethernet adapter IP (not the WSL virtual adapter).

**Example:**
```
IPv4 Address. . . . . . . . . . . : 10.162.97.69
```

---

### Test 3: Connect from Another Computer

From another computer on the same network:

```bash
ssh -p 4569 -o StrictHostKeyChecking=no 10.162.97.69
```

**Replace `10.162.97.69` with your actual Windows IP.**

**Expected result:** Connection succeeds and you see the SSH server's TUI/resume

**What this tests:**
- Full network path works (other computer → Windows → WSL2)
- Firewall allows external connections
- Port forwarding works for external traffic

---

## Connection Commands Reference

| From | Command | Notes |
|------|---------|-------|
| WSL2 itself | `ssh -p 4569 localhost` | Direct connection |
| Windows | `ssh -p 4569 localhost` | Through port forwarding |
| Same computer | `ssh -p 4569 127.0.0.1` | Localhost works too |
| Other computer | `ssh -p 4569 <Windows-IP>` | External connection |
| Other computer | `ssh -p 4569 10.162.97.69` | Example with actual IP |

---

## Important Notes

### WSL2 IP Changes

**The WSL2 IP address changes every time WSL2 restarts.**

When this happens:
1. Port forwarding will break
2. You need to re-run the setup script or update the port proxy rule

**To update the rule:**
```powershell
# 1. Get new WSL2 IP
$wslIp = wsl hostname -I

# 2. Delete old rule
netsh interface portproxy delete v4tov4 listenport=4569 listenaddress=0.0.0.0

# 3. Add new rule with updated IP
netsh interface portproxy add v4tov4 listenport=4569 listenaddress=0.0.0.0 connectport=4569 connectaddress=$wslIp
```

---

### Persistent IP (Advanced)

To make WSL2 use a static IP, you can configure it in `.wslconfig`:

```powershell
notepad $env:USERPROFILE\.wslconfig
```

Add:
```ini
[wsl2]
networkingMode=mirrored
```

Then restart WSL:
```powershell
wsl --shutdown
```

This enables mirrored networking mode where WSL2 shares the Windows network stack, eliminating the need for port forwarding.

---

## Troubleshooting

### Issue: "Connection refused" from Windows

**Cause:** Port forwarding not configured or WSL2 IP changed

**Solution:**
```powershell
# Check if rule exists
netsh interface portproxy show v4tov4

# If rule exists but wrong IP, delete and recreate
netsh interface portproxy delete v4tov4 listenport=4569 listenaddress=0.0.0.0
netsh interface portproxy add v4tov4 listenport=4569 listenaddress=0.0.0.0 connectport=4569 connectaddress=<NEW_WSL2_IP>
```

---

### Issue: "Connection timed out" from other computers

**Cause:** Firewall blocking or wrong IP address

**Solution:**
```powershell
# Verify firewall rule
netsh advfirewall firewall show rule name="SSH Server WSL2 Port 4569"

# If not found, recreate it
netsh advfirewall firewall add rule name="SSH Server WSL2 Port 4569" dir=in action=allow protocol=TCP localport=4569

# Check Windows IP
ipconfig
```

---

### Issue: "No route to host" from other computers

**Cause:** Computers on different networks or AP isolation enabled

**Solution:**
- Ensure both computers are on the same network (same Wi-Fi/router)
- If using mobile hotspot, some phones have "AP isolation" that prevents device-to-device communication
- Try connecting both computers to the same Wi-Fi network instead

---

### Issue: Setup script fails with "Access denied"

**Cause:** PowerShell not running as Administrator

**Solution:**
1. Close PowerShell
2. Right-click on PowerShell → "Run as Administrator"
3. Try the script again

---

### Issue: Port already in use

**Cause:** Another application is using port 4569

**Solution:**
```powershell
# Find what's using port 4569
netstat -ano | findstr :4569

# Stop that process (replace <PID> with the number from above)
taskkill /PID <PID> /F
```

Or use a different port:
```powershell
# Use port 7022 instead
netsh interface portproxy add v4tov4 listenport=7022 listenaddress=0.0.0.0 connectport=7022 connectaddress=172.26.176.1
netsh advfirewall firewall add rule name="SSH Server WSL2 Port 7022" dir=in action=allow protocol=TCP localport=7022
```

---

## Cleanup / Removal

If you want to remove the port forwarding and firewall rules:

```powershell
# Remove port forwarding
netsh interface portproxy delete v4tov4 listenport=4569 listenaddress=0.0.0.0

# Remove firewall rule
netsh advfirewall firewall delete rule name="SSH Server WSL2 Port 4569"
```

---

## Summary

**What we accomplished:**
1. ✅ Identified WSL2's virtual IP address
2. ✅ Created port forwarding from Windows (0.0.0.0:4569) to WSL2 (172.26.176.1:4569)
3. ✅ Configured Windows Firewall to allow inbound connections on port 4569
4. ✅ Verified the setup works from Windows and other computers

**Key concept:** WSL2 runs in an isolated VM. Port forwarding acts as a bridge, allowing external computers to reach services running inside WSL2 by connecting to the Windows host, which then forwards traffic to WSL2.

**Remember:** The WSL2 IP changes on restart, so you may need to re-run the setup after rebooting WSL2.
