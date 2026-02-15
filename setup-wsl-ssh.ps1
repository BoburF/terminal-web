#Requires -RunAsAdministrator

<#
.SYNOPSIS
    Sets up port forwarding from Windows to WSL2 for SSH server on port 4569
.DESCRIPTION
    Configures Windows to forward port 4569 to WSL2 VM and allows it through firewall
    Run this script as Administrator in PowerShell
.NOTES
    File Name      : setup-wsl-ssh.ps1
    Author         : WSL2 SSH Setup
    Prerequisite   : PowerShell 5.0 or higher, Administrator privileges
    Copyright      : Free to use
#>

param(
    [int]$Port = 4569,
    [switch]$Cleanup,
    [switch]$ShowStatus
)

function Show-Header {
    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host "  WSL2 SSH Server Port Forwarding Setup  " -ForegroundColor Cyan
    Write-Host "==========================================" -ForegroundColor Cyan
    Write-Host ""
}

function Get-WslIp {
    Write-Host "Getting WSL2 IP address..." -ForegroundColor Yellow
    try {
        $wslIp = wsl hostname -I 2>$null
        $wslIp = $wslIp.Trim()
        
        if ([string]::IsNullOrWhiteSpace($wslIp)) {
            Write-Error "Could not get WSL2 IP address. Is WSL running?"
            exit 1
        }
        
        Write-Host "  WSL2 IP: $wslIp" -ForegroundColor Green
        return $wslIp
    }
    catch {
        Write-Error "Failed to get WSL2 IP: $_"
        exit 1
    }
}

function Get-WindowsIps {
    Write-Host "Getting Windows IP addresses..." -ForegroundColor Yellow
    try {
        $windowsIps = Get-NetIPAddress -AddressFamily IPv4 | Where-Object { 
            $_.IPAddress -notlike "127.*" -and 
            $_.IPAddress -notlike "169.254.*" -and
            $_.PrefixOrigin -eq "Dhcp" -or $_.PrefixOrigin -eq "Manual"
        } | Select-Object -ExpandProperty IPAddress -Unique
        
        Write-Host "  Windows IPs:" -ForegroundColor Green
        $windowsIps | ForEach-Object { Write-Host "    $_" -ForegroundColor Cyan }
        return $windowsIps
    }
    catch {
        Write-Warning "Could not get Windows IPs: $_"
        return @()
    }
}

function Remove-PortForwarding {
    param([int]$Port)
    
    Write-Host "Cleaning up existing port forwarding rules for port $Port..." -ForegroundColor Yellow
    $result = netsh interface portproxy delete v4tov4 listenport=$Port listenaddress=0.0.0.0 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✓ Removed existing rule" -ForegroundColor Green
    }
    else {
        Write-Host "  (No existing rule to remove)" -ForegroundColor Gray
    }
}

function Add-PortForwarding {
    param(
        [int]$Port,
        [string]$WslIp
    )
    
    Write-Host "Adding port forwarding rule..." -ForegroundColor Yellow
    $result = netsh interface portproxy add v4tov4 listenport=$Port listenaddress=0.0.0.0 connectport=$Port connectaddress=$WslIp 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✓ Port forwarding configured successfully" -ForegroundColor Green
        Write-Host "    Windows:0.0.0.0:$Port → WSL2:$WslIp`:$Port" -ForegroundColor Cyan
    }
    else {
        Write-Error "Failed to add port forwarding: $result"
        exit 1
    }
}

function Remove-FirewallRule {
    param([int]$Port)
    
    Write-Host "Removing existing firewall rules for port $Port..." -ForegroundColor Yellow
    $result = netsh advfirewall firewall delete rule name="SSH Server WSL2 Port $Port" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✓ Removed existing firewall rule" -ForegroundColor Green
    }
    else {
        Write-Host "  (No existing rule to remove)" -ForegroundColor Gray
    }
}

function Add-FirewallRule {
    param([int]$Port)
    
    Write-Host "Adding Windows Firewall rule..." -ForegroundColor Yellow
    $result = netsh advfirewall firewall add rule name="SSH Server WSL2 Port $Port" dir=in action=allow protocol=TCP localport=$Port 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✓ Firewall rule added successfully" -ForegroundColor Green
    }
    else {
        Write-Warning "Failed to add firewall rule: $result"
    }
}

function Show-PortProxyStatus {
    param([int]$Port)
    
    Write-Host ""
    Write-Host "Current Port Forwarding Rules:" -ForegroundColor Yellow
    Write-Host "----------------------------------------" -ForegroundColor Gray
    $rules = netsh interface portproxy show v4tov4 2>&1
    
    $found = $false
    foreach ($line in $rules) {
        if ($line -match $Port) {
            Write-Host $line -ForegroundColor Green
            $found = $true
        }
    }
    
    if (-not $found) {
        Write-Host "  No rules found for port $Port" -ForegroundColor Red
    }
    Write-Host "----------------------------------------" -ForegroundColor Gray
}

function Show-FirewallStatus {
    param([int]$Port)
    
    Write-Host ""
    Write-Host "Firewall Rule Status:" -ForegroundColor Yellow
    Write-Host "----------------------------------------" -ForegroundColor Gray
    $result = netsh advfirewall firewall show rule name="SSH Server WSL2 Port $Port" 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host $result -ForegroundColor Green
    }
    else {
        Write-Host "  No firewall rule found for port $Port" -ForegroundColor Red
    }
    Write-Host "----------------------------------------" -ForegroundColor Gray
}

function Show-Summary {
    param(
        [int]$Port,
        [array]$WindowsIps,
        [string]$WslIp
    )
    
    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host "         Setup Complete!                 " -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Connection Information:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  From Windows PowerShell:" -ForegroundColor White
    Write-Host "    ssh -p $Port -o StrictHostKeyChecking=no localhost" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  From other computers:" -ForegroundColor White
    if ($WindowsIps.Count -gt 0) {
        $WindowsIps | ForEach-Object { 
            Write-Host "    ssh -p $Port -o StrictHostKeyChecking=no $_" -ForegroundColor Cyan 
        }
    }
    else {
        Write-Host "    Run 'ipconfig' to find your Windows IP" -ForegroundColor Yellow
    }
    Write-Host ""
    Write-Host "Important Notes:" -ForegroundColor Yellow
    Write-Host "  • WSL2 IP ($WslIp) may change after WSL restarts" -ForegroundColor Gray
    Write-Host "  • Run this script again if connections fail" -ForegroundColor Gray
    Write-Host "  • To remove setup, run with -Cleanup flag" -ForegroundColor Gray
    Write-Host ""
}

function Remove-Setup {
    param([int]$Port)
    
    Write-Host ""
    Write-Host "Removing WSL2 SSH Port Forwarding Setup..." -ForegroundColor Yellow
    Write-Host ""
    
    Remove-PortForwarding -Port $Port
    Remove-FirewallRule -Port $Port
    
    Write-Host ""
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host "         Cleanup Complete!               " -ForegroundColor Green
    Write-Host "==========================================" -ForegroundColor Green
    Write-Host ""
}

# Main execution
Show-Header

if ($Cleanup) {
    Remove-Setup -Port $Port
    exit 0
}

if ($ShowStatus) {
    Show-PortProxyStatus -Port $Port
    Show-FirewallStatus -Port $Port
    exit 0
}

# Normal setup flow
Write-Host "This script will configure port forwarding and firewall rules" -ForegroundColor White
Write-Host "to allow external SSH connections to your WSL2 server." -ForegroundColor White
Write-Host ""
Write-Host "Press any key to continue or Ctrl+C to cancel..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
Write-Host ""

# Get IPs
$wslIp = Get-WslIp
$windowsIps = Get-WindowsIps

Write-Host ""

# Clean up existing rules
Remove-PortForwarding -Port $Port

# Add new port forwarding
Add-PortForwarding -Port $Port -WslIp $wslIp

Write-Host ""

# Configure firewall
Remove-FirewallRule -Port $Port
Add-FirewallRule -Port $Port

Write-Host ""

# Show status
Show-PortProxyStatus -Port $Port

# Show summary
Show-Summary -Port $Port -WindowsIps $windowsIps -WslIp $wslIp
