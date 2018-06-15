
$script_dir = [System.IO.Path]::GetDirectoryName($myInvocation.MyCommand.Definition)
$output_dir="$script_dir\..\output"
$linuxkit_dir="$script_dir\..\linuxkit"
$vpnkit_dir="$script_dir\..\..\vpnkit"

docker build "$vpnkit_dir\c\vpnkit-tap-vsockd" --tag cfdev/vpnkit-tap-vsockd:dev
docker build "$vpnkit_dir\c\vpnkit-9pmount-vsock" --tag cfdev/vpnkit-9pmount-vsock:dev

mkdir -p "$output_dir"

linuxkit pkg build -hash dev "$linuxkit_dir/pkg/bosh-lite-routing"
linuxkit pkg build -hash dev "$linuxkit_dir/pkg/expose-multiple-ports"
linuxkit pkg build -hash dev "$linuxkit_dir/pkg/garden-runc"
linuxkit pkg build -hash dev "$linuxkit_dir/pkg/openssl"

linuxkit build `
 -disable-content-trust `
 -name cfdev `
 -format iso-efi `
 -dir "$output_dir" `
 "$linuxkit_dir\base-windows.yml" `
 "$linuxkit_dir\garden.yml" 
