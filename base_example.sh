#!/bin/bash
set -e

# Install LXC if not present
if ! command -v lxc-create &> /dev/null; then
    sudo apt-get update
    sudo apt-get install -y lxc lxc-templates
fi

# Create base container
sudo lxc-create -n troje_base -t download -- -d alpine -r 3.18 -a amd64

# Configure container
sudo bash -c 'cat > /var/lib/lxc/troje_base/config << EOF
lxc.net.0.type = veth
lxc.net.0.link = lxcbr0
lxc.net.0.flags = up
lxc.net.0.hwaddr = 00:16:3e:xx:xx:xx

# Security
lxc.cap.drop = mac_admin mac_override sys_time sys_module sys_rawio

# Resource limits
lxc.cgroup.memory.limit_in_bytes = 512M
lxc.cgroup.cpu.shares = 512
EOF'

# Start container to configure SSH
sudo lxc-start -n troje_base
sudo lxc-attach -n troje_base -- /bin/sh -c '
    apk update
    apk add openssh
    rc-update add sshd
    echo "PermitRootLogin yes" >> /etc/ssh/sshd_config
    echo "root:changeme" | chpasswd
    /etc/init.d/sshd start
'

# Stop container for cloning
sudo lxc-stop -n troje_base
