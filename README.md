# csi-shared-lvm

A Kubernetes CSI Driver for shared storage based on LVM.

Unlike other LVM drivers, `csi-shared-lvm` is designed specifically for shared storage environments where multiple nodes
access the same physical block device concurrently. It solves the critical problem of concurrent metadata access and
locking without any external dependencies. Instead, it leverages the Kubernetes API itself as the source of truth for
distributed locking for controlled access to the LVM metadata.

> [!CAUTION]
> **This software is currently in ALPHA and is highly experimental.** It is under active development and may contain
> bugs that could result in severe **data loss**. It is intended for testing and development environments only.
>
> **DO NOT USE THIS IN PRODUCTION**. You have been warned.

## Features

- **Shared Storage Support**: Works with any block device shared across nodes.
- **No External Cluster Software**: Removes the complexity of managing a secondary cluster-forming (such as `corosync`)
  tool just for storage.
- **Kubernetes-Native Locking**: The CSI controller uses the native Kubernetes Leases / Leader Election APIs to ensure
  that only a single Pod manages the access to the LVM metadata.
- **Filesystem & Block Support**:
    - **Filesystem**: `ReadWriteOnce` (RWO) / `ReadOnlyMany` (ROX). Safely mounts ext4/xfs on a single node.
    - **Raw Block**: `ReadWriteMany` (RWX). Allows multiple nodes to attach the raw device (e.g. RAW disks for usage
      with Kube-Virt).
- **Dynamic Provisioning**: Create and delete Logical Volumes (LVs) on demand.
- **Volume Expansion**: Online resizing of both filesystems and block volumes.

## Architecture

The solution is composed of two main components:

1. **CSI Controller**: Runs as a Deployment with Leader Election. It handles the volume lifecycle and performs LVM
   metadata operations (`lvcreate`, `lvremove`, `lvextend`, etc.).
2. **CSI Node**: Runs on every node. It handles volume activation and mounting (`lvchange`, `mkfs`, `mount`,
   `resize2fs`, etc.).

## Prerequisites

- **Kubernetes Cluster**: v1.34+
- **Shared Storage**: All nodes in the cluster (or the subset of nodes you intend to use) must share a common block
  device.
- **LVM**: The Kernel driver must be enabled on all cluster nodes.
- **Volume Group**: A Volume Group (VG) must be pre-created on the shared device and visible to all nodes.

### Typical Usage Scenarios

#### 1. Local Testing (Loopback)

For single-node testing (e.g., Minikube, Kind), you can emulate a shared device using a loopback file.

```bash
truncate -s 10G /tmp/shared-disk.img
losetup -fP --show /tmp/shared-disk.img

# assuming it was mapped to /dev/loop0
pvcreate /dev/loop0
vgcreate csi-lvm-vg /dev/loop0
```

#### 2. Bare Metal (iSCSI / SAN)

Standard setup:

1. Expose a LUN from your SAN/NAS via iSCSI or Fibre Channel.
2. Configure the initiator on all nodes to log in to the target.
3. Ensure the device path (e.g., `/dev/mapper/...` for `multipath`) is visible on all nodes.
4. Create the VG on one node; it will propagate to others.

### Alternate Setups

Although untested, other setups should be possible, such as:

- **AWS EBS**: EBS Multi-Attach supports RWX attachment for Provisioned IOPS SSD volumes (`io1` and `io2`).
- **QEMU**: Create multiple VMs on the same network and bring up a Kubernetes cluster using your favorite tool (e.g.,
  RKE2). You should be able to mount a shared disk image using the `share-rw=on` device option.

## Installation

### Helm Chart

1. **Install the Chart**
   ```bash
   helm upgrade --install csi-shared-lvm ./charts/csi-shared-lvm
   ```

2. **Verify Components**
   ```bash
   kubectl get pods -n kube-system
   ```

## Usage

### StorageClass

Create a `StorageClass` pointing to your shared Volume Group.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: shared-lvm-ext4
provisioner: csi-shared-lvm.cienijr.github.com
parameters:
  volumeGroup: "csi-lvm-vg"  # should match the VG name created on the host
  fsType: "ext4"             # supports ext4 or xfs
reclaimPolicy: Delete
volumeBindingMode: Immediate
allowVolumeExpansion: true
```

### PersistentVolumeClaim (PVC)

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-shared-vol
spec:
  # ReadWriteMany supported only for Block volumes
  accessModes:
  - ReadWriteOnce
  # supported modes: Filesystem / Block
  volumeMode: Filesystem
  resources:
    requests:
      storage: 1Gi
  storageClassName: shared-lvm-ext4
```

### Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: task-pv-pod
spec:
  containers:
  - name: task-pv-container
    image: nginx
    volumeMounts:
    - mountPath: "/usr/share/nginx/html"
      name: task-pv-storage
  volumes:
  - name: task-pv-storage
    persistentVolumeClaim:
      claimName: my-shared-vol
```

### Block Mode vs. Filesystem

**Filesystem Mode (`ext4`, `xfs`)**:

* **Modes**: `ReadWriteOnce` (RWO), `ReadOnlyMany` (ROX).
* **Constraint**: Standard filesystems are **not** cluster-aware. Mounting ext4 RW on two nodes simultaneously will
  definitely destroy your data. The driver strictly enforces Single-Node-Writer behavior.

**Block Mode**:

* **Modes**: `ReadWriteOnce` (RWO), `ReadWriteMany` (RWX).
* **Usage**: The driver exposes the raw block device (e.g., `/dev/csi-lvm-vg/pvc-xxx`) inside the container.
* **Warning**: If you use RWX, **YOU** are responsible for setting up a cluster-aware filesystem or ensuring that writes
  are properly coordinated.

### Volume Expansion

Resizing is supported and online.

1. Edit the PVC: `kubectl edit pvc my-shared-vol`
2. Increase `spec.resources.requests.storage`.
3. The driver will:
    * Expand the Logical Volume (LVM) (controller plugin).
    * Resize the filesystem (resize2fs/xfs_growfs) if applicable (node plugin).

## Troubleshooting

### Volume Group Not Found

**Symptom**: Pods remain in `ContainerCreating` with events like:

```text
MountVolume.MountDevice failed for volume "pvc-xxx": rpc error: code = Internal desc = Volume group "vg-xxx" not found
```

**Cause**: The shared block device is not visible on the node where the pod was scheduled.

**Resolution**:

1. SSH into the affected node.
2. Run `vgs` to verify if the Volume Group is visible.
3. If not, check `lsblk` and your storage backend connectivity (in case of networked devices).
4. Restart the `csi-node` pod once the storage is fixed to force a retry.

**Important:** This driver is currently not topology-aware. It assumes that the backing VG is available across all
cluster nodes. If you need to restrict pods to a subset of nodes, you must configure node affinity or taints/tolerations
directly on the Pods.

## Development

### Build

```bash
# build binaries
make build

# run unit tests
make test

# build container image
make image
```

## License

MIT License. Copyright (c) 2025 José Carlos Cieni Júnior.
