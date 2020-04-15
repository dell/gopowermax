/*
 Copyright © 2020 Dell Inc. or its subsidiaries. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
      http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/
package types

// QueryParams is a map of key value pairs that can be
// appended to any url as query parameters.
type QueryParams map[string]interface{}

// IncludeDetails is boolean flag that can be passed as a query param to the
// volume listing endpoing for getting the extensive details about the snapshots.
const IncludeDetails = "includeDetails"

// SnapshotName can be passed as a query param to the volume listing
// endpoing for filtering the results based on snapshot name
const SnapshotName = "snapshotName"

// InSG can be passed as a query param to the volume listing
// endpoing for filtering the results based on their
// association to a storage group
const InSG = "inSG"

// IsRdf can be passed as a query param to the volume listing
// endpoing for filtering the resluts based on their RDF relationship
const IsRdf = "isRdf"

// VolumeList contains list of device names
type VolumeList struct {
	Name string `json:"name"`
}

// CreateVolumesSnapshot contains parameters to create a volume snapshot
type CreateVolumesSnapshot struct {
	SourceVolumeList []VolumeList `json:"deviceNameListSource"`
	BothSides        bool         `json:"bothSides"`
	Star             bool         `json:"star"`
	Force            bool         `json:"force"`
	TimeInHours      bool         `json:"timeInHours"`
	TimeToLive       int64        `json:"timeToLive"`
	TTL              int64        `json:"ttl,omitempty"`
	Securettl        int64        `json:"securettl,omitempty"`
	ExecutionOption  string       `json:"executionOption"`
}

// ModifyVolumeSnapshot contains input parameters to modify the snapshot
type ModifyVolumeSnapshot struct {
	VolumeNameListSource []VolumeList `json:"deviceNameListSource"`
	VolumeNameListTarget []VolumeList `json:"deviceNameListTarget"`
	Force                bool         `json:"force,omitempty"`
	Star                 bool         `json:"star,omitempty"`
	Exact                bool         `json:"exact,omitempty"`
	Copy                 bool         `json:"copy,omitempty"`
	Remote               bool         `json:"remote,omitempty"`
	Symforce             bool         `json:"symforce,omitempty"`
	NoCopy               bool         `json:"nocopy,omitempty"`
	TTL                  int64        `json:"ttl,omitempty"`
	SecureTTL            int64        `json:"securettl,omitempty"`
	NewSnapshotName      string       `json:"newsnapshotname,omitempty"`
	TimeInHours          bool         `json:"timeInHours"`
	Action               string       `json:"action"`
	Generation           int64        `json:"generation"`
	ExecutionOption      string       `json:"executionOption,omitempty"`
}

// DeleteVolumeSnapshot contains input parameters to delete the snapshot
type DeleteVolumeSnapshot struct {
	DeviceNameListSource []VolumeList `json:"deviceNameListSource"`
	Symforce             bool         `json:"symforce,omitempty"`
	Star                 bool         `json:"star,omitempty"`
	Force                bool         `json:"force,omitempty"`
	Restore              bool         `json:"restore,omitempty"`
	Generation           int64        `json:"generation"`
	ExecutionOption      string       `json:"executionOption,omitempty"`
}

// VolumeSnapshotSource holds information on volume snapshot source
type VolumeSnapshotSource struct {
	SnapshotName         string          `json:"snapshotName"`
	Generation           int64           `json:"generation"`
	TimeStamp            string          `json:"timestamp"`
	State                string          `json:"state"`
	ProtectionExpireTime int64           `json:"protectionExpireTime"`
	GCM                  bool            `json:"gcm"`
	ICDP                 bool            `json:"icdp"`
	Secured              bool            `json:"secured"`
	IsRestored           bool            `json:"isRestored"`
	TTL                  int64           `json:"ttl"`
	Expired              bool            `json:"expired"`
	LinkedVolumes        []LinkedVolumes `json:"linkedDevices"`
}

// LinkedVolumes contains information about linked volumes of the snapshot
type LinkedVolumes struct {
	TargetDevice     string `json:"targetDevice"`
	Timestamp        string `json:"timestamp"`
	State            string `json:"state"`
	TrackSize        int64  `json:"trackSize"`
	Tracks           int64  `json:"tracks"`
	PercentageCopied int64  `json:"percentageCopied"`
	Linked           bool   `json:"linked"`
	Restored         bool   `json:"restored"`
	Defined          bool   `json:"defined"`
	Copy             bool   `json:"copy"`
	Destage          bool   `json:"destage"`
	Modified         bool   `json:"modified"`
}

// VolumeSnapshotLink contains information about linked snapshots
type VolumeSnapshotLink struct {
	TargetDevice     string `json:"targetDevice"`
	Timestamp        string `json:"timestamp"`
	State            string `json:"state"`
	TrackSize        int64  `json:"trackSize"`
	Tracks           int64  `json:"tracks"`
	PercentageCopied int64  `json:"percentageCopied"`
	Linked           bool   `json:"linked"`
	Restored         bool   `json:"restored"`
	Defined          bool   `json:"defined"`
	Copy             bool   `json:"copy"`
	Destage          bool   `json:"destage"`
	Modified         bool   `json:"modified"`
}

// VolumeSnapshot contains list of volume snapshots
type VolumeSnapshot struct {
	DeviceName           string                 `json:"deviceName"`
	SnapshotName         string                 `json:"snapshotName"`
	VolumeSnapshotSource []VolumeSnapshotSource `json:"snapshotSrc"`
	VolumeSnapshotLink   []VolumeSnapshotLink   `json:"snapshotLnk,omitempty"`
}

// SnapshotVolumeGeneration contains information on all snapshots related to a volume
type SnapshotVolumeGeneration struct {
	DeviceName           string                 `json:"deviceName"`
	VolumeSnapshotSource []VolumeSnapshotSource `json:"snapshotSrcs"`
	VolumeSnapshotLink   []VolumeSnapshotLink   `json:"snapshotLnks,omitempty"`
}

// VolumeSnapshotGeneration contains information on generation of a snapshot
type VolumeSnapshotGeneration struct {
	DeviceName           string               `json:"deviceName"`
	SnapshotName         string               `json:"snapshotName"`
	Generation           int64                `json:"generation"`
	VolumeSnapshotSource VolumeSnapshotSource `json:"snapshotSrc"`
	VolumeSnapshotLink   []VolumeSnapshotLink `json:"snapshotLnk,omitempty"`
}

// VolumeSnapshotGenerations contains list of volume snapshot generations
type VolumeSnapshotGenerations struct {
	DeviceName           string                 `json:"deviceName"`
	Generation           []int64                `json:"generation"`
	SnapshotName         string                 `json:"snapshotName"`
	VolumeSnapshotSource []VolumeSnapshotSource `json:"snapshotSrc"`
	VolumeSnapshotLink   []VolumeSnapshotLink   `json:"snapshotLnk,omitempty"`
}

// SymDevice list of devices on a particular symmetrix system
type SymDevice struct {
	SymmetrixID string     `json:"symmetrixId"`
	Name        string     `json:"name"`
	Snapshot    []Snapshot `json:"snapshot"`
	RdfgNumbers []int64    `json:"rdfgNumbers"`
}

//Snapshot contains information for a snapshot
type Snapshot struct {
	Name       string `json:"name"`
	Generation int64  `json:"generation"`
	Linked     bool   `json:"linked"`
	Restored   bool   `json:"restored"`
	Timestamp  string `json:"timestamp"`
	State      string `json:"state"`
}

// SymVolumeList contains information on private volume get
type SymVolumeList struct {
	Name      []string    `json:"name"`
	SymDevice []SymDevice `json:"device"`
}

// SymmetrixCapability holds replication capabilities
type SymmetrixCapability struct {
	SymmetrixID   string `json:"symmetrixId"`
	SnapVxCapable bool   `json:"snapVxCapable"`
	RdfCapable    bool   `json:"rdfCapable"`
}

// SymReplicationCapabilities holds whether or not snapshot is licensed
type SymReplicationCapabilities struct {
	SymmetrixCapability []SymmetrixCapability `json:"symmetrixCapability"`
	Successful          bool                  `json:"successful,omitempty"`
	FailMessage         string                `json:"failMessage,omitempty"`
}

// PrivVolumeResultList : volume list resulted
type PrivVolumeResultList struct {
	PrivVolumeList []VolumeResultPrivate `json:"result"`
	From           int                   `json:"from"`
	To             int                   `json:"to"`
}

// PrivVolumeIterator : holds the iterator of resultant volume list
type PrivVolumeIterator struct {
	ResultList PrivVolumeResultList `json:"resultList"`
	ID         string               `json:"id"`
	Count      int                  `json:"count"`
	// What units is ExpirationTime in?
	ExpirationTime int64 `json:"expirationTime"`
	MaxPageSize    int   `json:"maxPageSize"`
}

// VolumeResultPrivate holds private volume information
type VolumeResultPrivate struct {
	VolumeHeader   VolumeHeader   `json:"volumeHeader"`
	TimeFinderInfo TimeFinderInfo `json:"timeFinderInfo"`
}

// VolumeHeader holds private volume header information
type VolumeHeader struct {
	VolumeID              string   `json:"volumeId"`
	NameModifier          string   `json:"nameModifier"`
	FormattedName         string   `json:"formattedName"`
	PhysicalDeviceName    string   `json:"physicalDeviceName"`
	Configuration         string   `json:"configuration"`
	SRP                   string   `json:"SRP"`
	ServiceLevel          string   `json:"serviceLevel"`
	ServiceLevelBaseName  string   `json:"serviceLevelBaseName"`
	Workload              string   `json:"workload"`
	StorageGroup          []string `json:"storageGroup"`
	FastStorageGroup      string   `json:"fastStorageGroup"`
	ServiceState          string   `json:"serviceState"`
	Status                string   `json:"status"`
	CapTB                 float64  `json:"capTB"`
	CapGB                 float64  `json:"capGB"`
	CapMB                 float64  `json:"capMB"`
	BlockSize             int64    `json:"blockSize"`
	AllocatedPercent      int64    `json:"allocatedPercent"`
	EmulationType         string   `json:"emulationType"`
	SystemResource        bool     `json:"system_resource"`
	Encapsulated          bool     `json:"encapsulated"`
	BCV                   bool     `json:"BCV"`
	SplitName             string   `json:"splitName"`
	SplitSerialNumber     string   `json:"splitSerialNumber"`
	FBA                   bool     `json:"FBA"`
	CKD                   bool     `json:"CKD"`
	Mapped                bool     `json:"mapped"`
	Private               bool     `json:"private"`
	DataDev               bool     `json:"dataDev"`
	VVol                  bool     `json:"VVol"`
	MobilityID            bool     `json:"mobilityID"`
	Meta                  bool     `json:"meta"`
	MetaHead              bool     `json:"metaHead"`
	NumSymDevMaskingViews int64    `json:"numSymDevMaskingViews"`
	NumStorageGroups      int64    `json:"numStorageGroups"`
	NumDGs                int64    `json:"numDGs"`
	NumCGs                int64    `json:"numCGs"`
	Lun                   string   `json:"lun"`
	MetaConfigNumber      int64    `json:"metaConfigNumber"`
	WWN                   string   `json:"wwn"`
	HasEffectiveWWN       bool     `json:"hasEffectiveWWN"`
	EffectiveWWN          string   `json:"effectiveWWN"`
	PersistentAllocation  string   `json:"persistentAllocation"`
	CUImageNum            string   `json:"CUImageNum"`
	CUImageStatus         string   `json:"CUImageStatus"`
	SSID                  string   `json:"SSID"`
	CUImageBaseAddress    string   `json:"CUImageBaseAddress"`
	PAVMode               string   `json:"PAVMode"`
	FEDirPorts            []string `json:"FEDirPorts"`
	CompressionEnabled    bool     `json:"compressionEnabled"`
	CompressionRatio      string   `json:"compressionRatio"`
}

// TimeFinderInfo contains snap information for a volume
type TimeFinderInfo struct {
	SnapSource    bool            `json:"snapSource"`
	SnapTarget    bool            `json:"snapTarget"`
	SnapVXSrc     bool            `json:"snapVXSrc"`
	SnapVXTgt     bool            `json:"snapVXTgt"`
	Mirror        bool            `json:"mirror"`
	CloneSrc      bool            `json:"cloneSrc"`
	CloneTarget   bool            `json:"cloneTarget"`
	SnapVXSession []SnapVXSession `json:"snapVXSession"`
	CloneSession  []CloneSession  `json:"cloneSession"`
	MirrorSession []MirrorSession `json:"MirrorSession"`
}

// SnapVXSession holds snapshot session information
type SnapVXSession struct {
	SourceSnapshotGenInfo       []SourceSnapshotGenInfo      `json:"srcSnapshotGenInfo"`
	LinkSnapshotGenInfo         []LinkSnapshotGenInfo        `json:"lnkSnapshotGenInfo"`
	TargetSourceSnapshotGenInfo *TargetSourceSnapshotGenInfo `json:"tgtSrcSnapshotGenInfo"`
}

// SourceSnapshotGenInfo contains source snapshot generation info
type SourceSnapshotGenInfo struct {
	SnapshotHeader      SnapshotHeader        `json:"snapshotHeader"`
	LinkSnapshotGenInfo []LinkSnapshotGenInfo `json:"lnkSnapshotGenInfo"`
}

// SnapshotHeader contians information for snapshot header
type SnapshotHeader struct {
	Device       string `json:"device"`
	SnapshotName string `json:"snapshotName"`
	Generation   int64  `json:"generation"`
	Secured      bool   `json:"secured"`
	Expired      bool   `json:"expired"`
	TimeToLive   int64  `json:"timeToLive"`
	Timestamp    int64  `json:"timestamp"`
}

// LinkSnapshotGenInfo contains information on snapshot generation for links
type LinkSnapshotGenInfo struct {
	TargetDevice  string `json:"targetDevice"`
	State         string `json:"state"`
	Restored      bool   `json:"restored"`
	Defined       bool   `json:"defined"`
	Destaged      bool   `json:"destaged"`
	BackgroundDef bool   `json:"backgroundDef"`
}

// TargetSourceSnapshotGenInfo contains information on target snapshot generation
type TargetSourceSnapshotGenInfo struct {
	TargetDevice string `json:"targetDevice"`
	SourceDevice string `json:"sourceDevice"`
	SnapshotName string `json:"snapshotName"`
	Generation   int64  `json:"generation"`
	Secured      bool   `json:"secured"`
	Expired      bool   `json:"expired"`
	TimeToLive   int64  `json:"timeToLive"`
	Timestamp    int64  `json:"timestamp"`
	Defined      string `json:"state"`
}

// CloneSession contains information on a clone session
type CloneSession struct {
	SourceVolume  string `json:"sourceVolume"`
	TargetVolume  string `json:"targetVolume"`
	Timestamp     int64  `json:"timestamp"`
	State         string `json:"state"`
	RemoteVolumes string `json:"remoteVolumes"`
}

// MirrorSession contains info about mirrored session
type MirrorSession struct {
	Timestamp    int64  `json:"timestamp"`
	State        string `json:"state"`
	SourceVolume string `json:"sourceVolume"`
	TargetVolume string `json:"targetVolume"`
}

//SnapTarget contains target information
type SnapTarget struct {
	Target  string
	Defined bool
	CpMode  bool
}
