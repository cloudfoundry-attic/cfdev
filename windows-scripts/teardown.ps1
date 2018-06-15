stop-vm -name cfdev -turnoff
remove-vm -name cfdev -force
stop-job -Name "vpnkit"
remove-vmswitch -Name cfdev -force

$ethServiceGuid="7207f451-2ca3-4b88-8d01-820a21d78293"
Remove-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices/$ethServiceGuid"

$portServiceGuid="cc2a519a-fb40-4e45-a9f1-c7f04c5ad7fa"
Remove-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices/$portServiceGuid"

$forwarderServiceGuid="e3ae8f06-8c25-47fb-b6ed-c20702bcef5e"
Remove-Item -Path "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Virtualization\GuestCommunicationServices/$forwarderServiceGuid"

