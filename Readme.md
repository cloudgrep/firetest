# Test JAILER enclosed MicroVM

Steps required to run:

1. Download all the binaries 

Getting all the rootfs and kernel.
```
ARCH="$(uname -m)"

latest=$(wget "http://spec.ccfc.min.s3.amazonaws.com/?prefix=firecracker-ci/v1.10/x86_64/vmlinux-5.10&list-type=2" -O - 2>/dev/null | grep "(?<=<Key>)(firecracker-ci/v1.10/x86_64/vmlinux-5\.10\.[0-9]{3})(?=</Key>)" -o -P)

# Download a linux kernel binary
wget "https://s3.amazonaws.com/spec.ccfc.min/${latest}"

# Download a rootfs
wget "https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.10/${ARCH}/ubuntu-22.04.ext4"

# Download the ssh key for the rootfs
wget "https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.10/${ARCH}/ubuntu-22.04.id_rsa"

# Set user read permission on the ssh key
chmod 400 ./ubuntu-22.04.id_rsa
```

Getting the firecracker and jailer binay.

Note : For jailer binary you might need to build the firecracker from scratch with docker.

```
ARCH="$(uname -m)"
release_url="https://github.com/firecracker-microvm/firecracker/releases"
latest=$(basename $(curl -fsSLI -o /dev/null -w  %{url_effective} ${release_url}/latest))
curl -L ${release_url}/download/${latest}/firecracker-${latest}-${ARCH}.tgz \
| tar -xz

# Rename the binary to "firecracker"
mv release-${latest}-$(uname -m)/firecracker-${latest}-${ARCH} firecracker

```

To instead build firecracker from source, you will need to have docker installed:

```
ARCH="$(uname -m)"

# Clone the firecracker repository
git clone https://github.com/firecracker-microvm/firecracker firecracker_src

# Start docker
sudo systemctl start docker

# Build firecracker
#
# It is possible to build for gnu, by passing the arguments '-l gnu'.
#
# This will produce the firecracker and jailer binaries under
# `./firecracker/build/cargo_target/${toolchain}/debug`.
#
sudo ./firecracker_src/tools/devtool build

# Rename the binary to "firecracker"
sudo cp ./firecracker_src/build/cargo_target/${ARCH}-unknown-linux-musl/debug/firecracker firecracker
```

Place all the binaries in the root directory with named as following:

ubuntu                                  --> fc_rfs
kernel                                  --> fc_kernel
firecracker_release                     --> firecracker
jailer_release                          --> jailer

Parameters set in static configuration file for firecracker.
```
UID := 123
GID := 100

const id = "4580"
const socketPath = "api.socket"
const kernelImagePath = "./fc_kernel"
const rfsPath = "./fc_rfs"
const firecrackerPath = "./firecracker"
const jailerPath = "./jailer"
const ChrootBaseDir = "/srv/jailer"
    
```
Note: 
Ip address of vm set to "172.16.0.2" !!!
CHroot Base Dir set to "/srv/jailer/" !!!

Sandbox Environment Details:

Docker network name - testJailer
ContainerId - randomly generated
Ip address - "172.16.0.2"


## To start the vm 
```
make run 
```

## Environment Details

Terminal stdio, stdout, stderr will be connected to microvm shell.