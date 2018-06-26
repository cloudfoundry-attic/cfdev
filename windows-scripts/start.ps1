param (
  [Parameter(Mandatory=$true)]
  $i,
  [Parameter(Mandatory=$true)]
  $f
)
if ([System.IO.Path]::IsPathRooted($i)) {
  $cfdev_efi_iso = $i
} else {
  $cfdev_efi_iso = Join-Path $PWD $i
}

if ([System.IO.Path]::IsPathRooted($f)) {
  $cf_deps_iso = $f
} else {
  $cf_deps_iso = Join-Path $PWD $f
}


$script_dir = [System.IO.Path]::GetDirectoryName($myInvocation.MyCommand.Definition)
$cf_dev_home="$HOME\.cfdev"
$vmName="cfdev"

function Generate-DNSFiles {
    $file_path = Join-Path $script_dir generate-dns-files.go
    go run $file_path
}

function Register-ServiceGuids {
  $ethServiceGuid="7207f451-2ca3-4b88-8d01-820a21d78293"
  $ethService = New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" -Name $ethServiceGuid
  $ethService.SetValue("ElementName", "CF Dev VPNkit Ethernet Service")

  $portServiceGuid="cc2a519a-fb40-4e45-a9f1-c7f04c5ad7fa"
  $portService = New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" -Name $portServiceGuid
  $portService.SetValue("ElementName", "CF Dev VPNkit Port Service")

  $forwarderServiceGuid="e3ae8f06-8c25-47fb-b6ed-c20702bcef5e"
  $forwarderService = New-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices" -Name $forwarderServiceGuid
  $forwarderService.SetValue("ElementName", "CF Dev VPNkit Forwarder Service")
}

function Add-IPAliases {
  $switchName="cfdev"
  New-VMSwitch -Name $switchName `
    -SwitchType Internal `
    -Notes 'Switch for CF Dev Networking'

  netsh interface ip add address  "vEthernet ($switchName)" 10.245.0.2 255.255.255.255
  netsh interface ip add address  "vEthernet ($switchName)" 10.144.0.34 255.255.255.255
}

function Create-VM {
  New-VM -Name $vmName -Generation 2 -NoVHD
  Set-VM -Name $vmName `
    -AutomaticStartAction Nothing `
    -AutomaticStopAction ShutDown `
    -CheckpointType Disabled `
    -MemoryStartupBytes 5GB `
    -StaticMemory `
    -ProcessorCount 4

  $id=(Get-VM -Name $vmName).Id
  echo "VM GUID: " $id
  mkdir $cf_dev_home -force
  rm $cf_dev_home\vpnkit.log

  Add-VMDvdDrive -VMName $vmName `
  -Path $cfdev_efi_iso

  Add-VMDvdDrive -VMName $vmName `
    -Path $cf_deps_iso

  Remove-VMNetworkAdapter -VMName cfdev -Name  "Network Adapter"
  $emptyVHD="$cf_dev_home\cfdev.vhd"
  rm $emptyVHD

  New-VHD -Path $emptyVHD -SizeBytes '200000000000' -Dynamic

  Add-VMHardDiskDrive -VMName $vmName -Path $emptyVHD

  Set-VMFirmware -VMName $vmName -EnableSecureBoot Off -FirstBootDevice $cdrom

  Set-VMComPort -VMName $vmName -number 1 -Path \\.\pipe\cfdev-com
}

function Start-VPNKit {
  $env:dns_path = Join-Path $script_dir resolv.conf
  $env:dhcp_path = Join-Path $script_dir dhcp.json 

  start-job -Name "vpnkit"  `
      -InitializationScript { $id=(Get-VM -Name cfdev).Id } `
      -ScriptBlock { C:\Users\pivotal\vpnkit\vpnkit.exe `
      --ethernet hyperv-connect://$id/"7207f451-2ca3-4b88-8d01-820a21d78293" `
      --port hyperv-connect://$id/"cc2a519a-fb40-4e45-a9f1-c7f04c5ad7fa" `
      --port hyperv-connect://$id/"e3ae8f06-8c25-47fb-b6ed-c20702bcef5e" `
      --dns $env:dns_path `
      --dhcp $env:dhcp_path `
      --diagnostics "\\.\pipe\cfdevVpnKitDiagnostics" `
      --listen-backlog 32 `
      --lowest-ip 169.254.82.3 `
      --highest-ip 169.254.82.255 `
      --host-ip 169.254.82.2 `
      --gateway-ip 169.254.82.1 2>&1 > $HOME\.cfdev\vpnkit.log }
}

function Main {
  Generate-DNSFiles
  Add-IPAliases
  Register-ServiceGuids
  Create-VM
  Start-VPNKit
  Start-VM -Name $vmName
}

Main