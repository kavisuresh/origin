package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

// ClusterNetwork describes a cluster network
type ClusterNetwork struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty"`

	// Network is a CIDR string to specify the global overlay network's L3 space
	Network string `json:"network" description:"CIDR string to specify the global overlay network's L3 space"`
	// HostSubnetLength is the number of bits to allocate to each host's subnet e.g. 8 would mean a /24 network on the host
	HostSubnetLength int `json:"hostsubnetlength" description:"number of bits to allocate to each host's subnet e.g. 8 would mean a /24 network on the host"`
	// ServiceNetwork is the CIDR string to specify the service network
	ServiceNetwork string `json:"serviceNetwork" description:"CIDR string to specify the service network"`
}

// ClusterNetworkList is a collection of ClusterNetworks
type ClusterNetworkList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	unversioned.ListMeta `json:"metadata,omitempty"`
	// Items is the list of cluster networks
	Items []ClusterNetwork `json:"items" description:"list of cluster networks"`
}

// HostSubnet encapsulates the inputs needed to define the container subnet network on a node
type HostSubnet struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty"`

	// Host is the name of the host that is registered at the master. May just be an IP address, resolvable hostname or a complete DNS.
	// A lease will be sought after this name.
	Host string `json:"host" description:"Name of the host that is registered at the master. A lease will be sought after this name."`
	// HostIP is the IP address to be used as vtep by other hosts in the overlay network
	HostIP string `json:"hostIP" description:"IP address to be used as vtep by other hosts in the overlay network"`
	// Subnet is the actual subnet CIDR lease assigned to the host
	Subnet string `json:"subnet" description:"Actual subnet CIDR lease assigned to the host"`
}

// HostSubnetList is a collection of HostSubnets
type HostSubnetList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	unversioned.ListMeta `json:"metadata,omitempty"`
	// Items is the list of host subnets
	Items []HostSubnet `json:"items" description:"list of host subnets"`
}

// NetNamespace encapsulates the inputs needed to define a unique network namespace on the cluster
type NetNamespace struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty"`

	// NetName is the name of the network namespace
	NetName string `json:"netname" description:"Name of the network namespace."`
	// NetID is the network identifier of the network namespace assigned to each overlay network packet
	NetID uint `json:"netid" description:"NetID of the network namespace assigned to each overlay network packet."`
}

// NetNamespaceList is a collection of NetNamespaces
type NetNamespaceList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	unversioned.ListMeta `json:"metadata,omitempty"`
	// Items is the list of net namespaces
	Items []NetNamespace `json:"items" description:"list of net namespaces"`
}
