$cfdev="$PSScriptRoot\..\..\.."
$cache_dir="$HOME\.cfdev\cache"

$pkg="code.cloudfoundry.org/cfdev/config"

$cfdepsUrl="$cfdev\output\cf-deps.iso"
If (-NOT (Test-Path $cfdepsUrl)){
    $cfdepsUrl="$cache_dir\cf-deps.iso"
}

$cfdevefiUrl="$cfdev\output\cfdev-efi.iso"
If (-NOT (Test-Path $cfdevefiUrl)) {
      $cfdevefiUrl="$cache_dir\cfdev-efi.iso"
}

$date=(Get-Date -Format FileDate)

go build `
  -ldflags `
    "-X $pkg.cfdepsUrl=$cfdepsUrl`
     -X $pkg.cfdepsMd5=$((Get-FileHash $cfdepsUrl -Algorithm MD5).Hash.ToLower())
     -X $pkg.cfdepsSize=$((Get-Item $cfdepsUrl).length)

     -X $pkg.cfdevefiUrl=$cfdevefiUrl
     -X $pkg.cfdevefiMd5=$((Get-FileHash $cfdevefiUrl -Algorithm MD5).Hash.ToLower())
     -X $pkg.cfdevefiSize=$((Get-Item $cfdevefiUrl).length)

     -X $pkg.vpnkitUrl=$cache_dir\vpnkit.exe
     -X $pkg.vpnkitMd5=$((Get-FileHash $cache_dir\vpnkit.exe -Algorithm MD5).Hash.ToLower())
     -X $pkg.vpnkitSize=$((Get-Item $cache_dir\vpnkit.exe).length)

     -X $pkg.winswUrl=$cache_dir\winsw.exe
     -X $pkg.winswMd5=$((Get-FileHash $cache_dir\winsw.exe -Algorithm MD5).Hash.ToLower())
     -X $pkg.winswSize=$((Get-Item $cache_dir\winsw.exe).length)

     -X $pkg.cliVersion=0.0.$date
     -X $pkg.analyticsKey=WFz4dVFXZUxN2Y6MzfUHJNWtlgXuOYV2" `
     code.cloudfoundry.org/cfdev