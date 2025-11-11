#!/bin/bash
# Install Talos Linux on existing Debian VM via kexec
# This script should be run on each GCP VM running Debian 12

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TALOS_VERSION="v1.11.5"
PLATFORM="metal"
ARCH="amd64"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Talos Linux Installation via kexec${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root${NC}"
    echo "Run: sudo $0"
    exit 1
fi

# Check if running on Debian
if [ ! -f /etc/debian_version ]; then
    echo -e "${RED}Error: This script is designed for Debian systems${NC}"
    exit 1
fi

echo -e "${GREEN}Detected Debian version:${NC} $(cat /etc/debian_version)"
echo -e "${GREEN}Talos version:${NC} $TALOS_VERSION"
echo ""

# Step 1: Capture current network configuration
echo -e "${YELLOW}Step 1: Capturing network configuration...${NC}"

PRIMARY_IF=$(ip route | grep default | awk '{print $5}' | head -n1)
IP_ADDR=$(ip addr show $PRIMARY_IF | grep 'inet ' | awk '{print $2}' | cut -d'/' -f1)
NETMASK=$(ip addr show $PRIMARY_IF | grep 'inet ' | awk '{print $2}' | cut -d'/' -f2)
GATEWAY=$(ip route | grep default | awk '{print $3}')
DNS=$(grep nameserver /etc/resolv.conf | head -n1 | awk '{print $2}')

echo "  Interface: $PRIMARY_IF"
echo "  IP Address: $IP_ADDR"
echo "  Netmask: $NETMASK"
echo "  Gateway: $GATEWAY"
echo "  DNS: $DNS"
echo ""

# Confirm configuration
read -p "Is this network configuration correct? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${RED}Installation cancelled${NC}"
    exit 1
fi

# Step 2: Download Talos kernel and initramfs
echo -e "${YELLOW}Step 2: Downloading Talos kernel and initramfs...${NC}"

DOWNLOAD_DIR="/tmp/talos-install"
mkdir -p $DOWNLOAD_DIR
cd $DOWNLOAD_DIR

KERNEL_URL="https://github.com/siderolabs/talos/releases/download/${TALOS_VERSION}/vmlinuz-${ARCH}"
INITRD_URL="https://github.com/siderolabs/talos/releases/download/${TALOS_VERSION}/initramfs-${ARCH}.xz"

echo "  Downloading vmlinuz..."
wget -q --show-progress $KERNEL_URL -O vmlinuz-${ARCH}

echo "  Downloading initramfs..."
wget -q --show-progress $INITRD_URL -O initramfs-${ARCH}.xz

echo -e "${GREEN}  ✓ Download complete${NC}"
echo ""

# Step 3: Install kexec-tools if not present
echo -e "${YELLOW}Step 3: Installing kexec-tools...${NC}"

if ! command -v kexec &> /dev/null; then
    apt-get update -qq
    apt-get install -y -qq kexec-tools
    echo -e "${GREEN}  ✓ kexec-tools installed${NC}"
else
    echo -e "${GREEN}  ✓ kexec-tools already installed${NC}"
fi

echo ""

# Step 4: Prepare kexec command
echo -e "${YELLOW}Step 4: Preparing kexec command...${NC}"

# Build kernel command line
# GCP metadata requires DHCP, so we'll use DHCP instead of static IP
CMDLINE="talos.platform=${PLATFORM} console=tty0 console=ttyS0 init_on_alloc=1 slab_nomerge pti=on consoleblank=0 nvme_core.io_timeout=4294967295 printk.devkmsg=on ima_template=ima-ng ima_appraise=fix ima_hash=sha512"

echo "  Kernel cmdline: $CMDLINE"
echo ""

# Step 5: Load Talos kernel with kexec
echo -e "${YELLOW}Step 5: Loading Talos kernel with kexec...${NC}"

kexec -l vmlinuz-${ARCH} \
    --initrd=initramfs-${ARCH}.xz \
    --command-line="$CMDLINE"

echo -e "${GREEN}  ✓ Talos kernel loaded${NC}"
echo ""

# Final confirmation
echo -e "${RED}========================================${NC}"
echo -e "${RED}WARNING: The system will now reboot into Talos Linux${NC}"
echo -e "${RED}========================================${NC}"
echo ""
echo "After reboot:"
echo "  1. Talos will run entirely in RAM"
echo "  2. The system will be accessible via the same IP"
echo "  3. SSH access will no longer work (use talosctl instead)"
echo "  4. Terraform will persist Talos to disk in the next step"
echo ""
echo "Network configuration:"
echo "  - Talos will use DHCP to obtain IP (GCP standard)"
echo "  - GCP metadata server access will be preserved"
echo "  - Terraform will apply machine configuration"
echo ""

read -p "Proceed with kexec reboot? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Installation cancelled${NC}"
    echo "To restart, run this script again"
    exit 1
fi

# Execute kexec
echo -e "${GREEN}Executing kexec...${NC}"
sleep 2

kexec -e

# This line will never be reached if kexec succeeds
echo -e "${RED}Error: kexec failed to execute${NC}"
exit 1
