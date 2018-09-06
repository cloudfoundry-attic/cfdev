$cfdev="$PSScriptRoot\..\..\.."
$cache_dir="$HOME\.cfdev\cache"

$pkg="code.cloudfoundry.org/cfdev/config"
$cfdepsUrl="$cache_dir\cf-deps.iso"
$cfdevefiUrl="$cache_dir\cfdev-efi.iso"
$cfAnalyticsdUrl="$cache_dir\analyticsd.exe"

$date=(Get-Date -Format FileDate)

go build -ldflags `
   "-X $pkg.analyticsdUrl=$cfAnalyticsdUrl
    -X $pkg.analyticsdMd5=$((Get-FileHash $cfAnalyticsdUrl -Algorithm MD5).Hash.ToLower())
    -X $pkg.analyticsdSize=$((Get-Item $cfAnalyticsdUrl).length)

    -X $pkg.cfdepsUrl=$cfdepsUrl
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