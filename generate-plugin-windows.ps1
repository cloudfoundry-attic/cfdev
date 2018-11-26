$cache_dir="$HOME\.cfdev\cache"

$pkg="code.cloudfoundry.org/cfdev/config"
# $cfdepsUrl="$cache_dir\cf-deps.iso"
$cfdepsUrl="C:\Users\pivotal\Downloads\cfdev-deps-windows-0.0.19.tgz"
$cfAnalyticsdUrl="$PWD\analytix.exe"

$date=(Get-Date -Format FileDate)

go build -ldflags `
  "-X main.version=0.0.$date
   -X main.testAnalyticsKey=WFz4dVFXZUxN2Y6MzfUHJNWtlgXuOYV2" `
   -o $cfAnalyticsdUrl `
   code.cloudfoundry.org/cfdev/analyticsd

go build -ldflags `
   "-X $pkg.analyticsdUrl=$cfAnalyticsdUrl
    -X $pkg.analyticsdMd5=$((Get-FileHash $cfAnalyticsdUrl -Algorithm MD5).Hash.ToLower())
    -X $pkg.analyticsdSize=$((Get-Item $cfAnalyticsdUrl).length)

    -X $pkg.cfdepsUrl=$cfdepsUrl
    -X $pkg.cfdepsMd5=$((Get-FileHash $cfdepsUrl -Algorithm MD5).Hash.ToLower())
    -X $pkg.cfdepsSize=$((Get-Item $cfdepsUrl).length)

    -X $pkg.cliVersion=0.0.$date
    -X $pkg.testAnalyticsKey=WFz4dVFXZUxN2Y6MzfUHJNWtlgXuOYV2" `
    code.cloudfoundry.org/cfdev
