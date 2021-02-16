package extstorage

import "context"

// Interface is the ganeti-extstorage interface according to
// https://docs.ganeti.org/docs/ganeti/3.0/html/man-ganeti-extstorage-interface.html#executable-scripts
type Interface interface {
	/*
		The create command is used for creating a new volume inside the external storage. The VOL_NAME denotes the volume’s name, which should be unique. After creation, Ganeti will refer to this volume by this name for all other actions.
		Ganeti produces this name dynamically and ensures its uniqueness inside the Ganeti context. Therefore, you should make sure not to provision manually additional volumes inside the external storage with this type of name, because this will lead to conflicts and possible loss of data.
		The VOL_SIZE variable denotes the size of the new volume to be created in mebibytes.
		If the script ends successfully, a new volume of size VOL_SIZE should exist inside the external storage. e.g:: a lun inside a NAS appliance.
		The script returns 0 on success.
	*/
	Create(context.Context, *VolumeInfo) error

	/*
		This command is used in order to make an already created volume visible to the physical node which will host the instance. This is done by mapping the already provisioned volume to a block device inside the host node.
		The VOL_NAME variable denotes the volume to be mapped.
		After successful attachment the script returns to its stdout a string, which is the full path of the block device to which the volume is mapped. e.g:: /dev/dummy1
		When attach returns, this path should be a valid block device on the host node.
		The attach script should be idempotent if the volume is already mapped. If the requested volume is already mapped, then the script should just return to its stdout the path which is already mapped to.
		In case the storage technology supports userspace access to volumes as well, e.g. the QEMU Hypervisor can see an RBD volume using its embedded driver for the RBD protocol, then the provider can return extra lines denoting the available userspace access URIs per hypervisor. The URI should be in the following format: <hypervisor>:<uri>. For example, a RADOS provider should return kvm:rbd:<pool>/<volume name> in the second line of stdout after the local block device path (e.g. /dev/rbd1).
		So, if the access disk parameter is userspace for the ext disk template, then the QEMU command will end up having file=<URI> in the -drive option.
		In case the storage technology supports only userspace access to volumes, then the first line of stdout should be an empty line, denoting that a local block device is not available. If neither a block device nor a URI is returned, then Ganeti will complain.
	*/
	Attach(context.Context, *VolumeInfo) error

	/*
		This command is used in order to unmap an already mapped volume from the host node. Detach undoes everything attach did. This is done by unmapping the requested volume from the block device it is mapped to.
		The VOL_NAME variable denotes the volume to be unmapped.
		detach doesn’t affect the volume itself. It just unmaps it from the host node. The volume continues to exist inside the external storage. It’s just not accessible by the node anymore. This script doesn’t return anything to its stdout.
		The detach script should be idempotent if the volume is already unmapped. If the volume is not mapped, the script doesn’t perform any action at all.
		The script returns 0 on success.
	*/
	Detach(context.Context, *VolumeInfo) error

	/*
		This command is used to remove an existing volume from the external storage. The volume is permanently removed from inside the external storage along with all its data.
		The VOL_NAME variable denotes the volume to be removed.
		The script returns 0 on success.
	*/
	Remove(context.Context, *VolumeInfo) error

	/*
		This command is used to grow an existing volume of the external storage.
		The VOL_NAME variable denotes the volume to grow.
		The VOL_SIZE variable denotes the current volume’s size (in mebibytes). The VOL_NEW_SIZE variable denotes the final size after the volume has been grown (in mebibytes).
		The amount of grow can be easily calculated by the script and is:
		grow_amount = VOL_NEW_SIZE - VOL_SIZE (in mebibytes)
		Ganeti ensures that: VOL_NEW_SIZE > VOL_SIZE
		If the script returns successfully, then the volume inside the external storage will have a new size of VOL_NEW_SIZE. This isn’t immediately reflected to the instance’s disk. See gnt-instance grow for more details on when the running instance becomes aware of its grown disk.
		The script returns 0 on success.
	*/
	Grow(context.Context, *VolumeInfo) error

	/*
	   This script is used to add metadata to an existing volume. It is helpful when we need to keep an external, Ganeti-independent mapping between instances and volumes; primarily for recovery reasons. This is provider specific and the author of the provider chooses whether/how to implement this. You can just exit with 0, if you do not want to implement this feature, without harming the overall functionality of the provider.
	   The VOL_METADATA variable contains the metadata of the volume.
	   Currently, Ganeti sets this value to originstname+X where X is the instance’s name.
	   The script returns 0 on success.
	*/
	Setinfo(context.Context, *VolumeInfo) error

	/*
		The verify script is used to verify consistency of the external parameters (ext-params) (see below). The command should take one or more arguments denoting what checks should be performed, and return a proper exit code depending on whether the validation failed or succeeded.
		Currently, the script is not invoked by Ganeti, but should be present for future use and consistency with gnt-os-interface’s verify script.
		The script should return 0 on success.
	*/
	Verify(context.Context, *VolumeInfo) error

	// Close closes the driver
	Close(context.Context) error
}
