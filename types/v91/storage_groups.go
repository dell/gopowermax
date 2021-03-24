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

import (
	"net/http"
)

// constants of storage units
const (
	CapacityUnitTb  = "TB"
	CapacityUnitGb  = "GB"
	CapacityUnitMb  = "MB"
	CapacityUnitCyl = "CYL"
)

// VolumeAttributeType : volume attributes for 9.1
type VolumeAttributeType struct {
	NumberOfVolumes  int                   `json:"num_of_vols,omitempty"`
	VolumeIdentifier *VolumeIdentifierType `json:"volumeIdentifier,omitempty"`
	CapacityUnit     string                `json:"capacityUnit"` // CAPACITY_UNIT_{TB,GB,MB,CYL}
	VolumeSize       string                `json:"volume_size"`
}

// VolumeIdentifierType : volume identifier
type VolumeIdentifierType struct {
	VolumeIdentifierChoice string `json:"volumeIdentifierChoice,omitempty"`
	IdentifierName         string `json:"identifier_name,omitempty"`
	AppendNumber           string `json:"append_number,omitempty"`
}

// AddVolumeParam holds number volumes to add and related param for 9.1
type AddVolumeParam struct {
	Emulation             string                `json:"emulation,omitempty"`
	CreateNewVolumes      bool                  `json:"create_new_volumes,omitempty"`
	VolumeAttributes      []VolumeAttributeType `json:"volumeAttributes,omitempty"`
	EnableMobilityID      string                `json:"enable_mobility_id,omitempty"`
	RemoteSymmSGInfoParam RemoteSymmSGInfoParam `json:"remoteSymmSGInfoParam,omitempty"`
}

// RemoveTagsParam holds array of tags to be removed
type RemoveTagsParam struct {
	TagName []string `json:"tag_name,omitempty"`
}

// AddTagsParam holds array of tags to be added
type AddTagsParam struct {
	TagName []string `json:"tag_name,omitempty"`
}

// TagManagementParam holds parameters to remove or add tags
type TagManagementParam struct {
	RemoveTagsParam *RemoveTagsParam `json:"removeTagsParam,omitempty"`
	AddTagsParam    *AddTagsParam    `json:"addTagsParam,omitempty"`
}

// MergeStorageGroupParam : Payloads for updating Storage Group
type MergeStorageGroupParam struct {
	StorageGroupID string `json:"storageGroupId,omitempty"`
}

// SplitStorageGroupVolumesParam holds parameters to split
type SplitStorageGroupVolumesParam struct {
	VolumeIDs      []string `json:"volumeId,omitempty"`
	StorageGroupID string   `json:"storageGroupId,omitempty"`
	MaskingViewID  string   `json:"maskingViewId,omitempty"`
}

// SplitChildStorageGroupParam holds param to split child SG
type SplitChildStorageGroupParam struct {
	StorageGroupID string `json:"storageGroupId,omitempty"`
	MaskingViewID  string `json:"maskingViewId,omitempty"`
}

// MoveVolumeToStorageGroupParam stores parameters to move volumes to SG
type MoveVolumeToStorageGroupParam struct {
	VolumeIDs      []string `json:"volumeId,omitempty"`
	StorageGroupID string   `json:"storageGroupId,omitempty"`
	Force          bool     `json:"force,omitempty"`
}

// EditCompressionParam hold param to edit compression attribute with an SG
type EditCompressionParam struct {
	Compression bool `json:"compression,omitempty"`
}

// SetHostIOLimitsParam holds param to set host IO limit
type SetHostIOLimitsParam struct {
	HostIOLimitMBSec    string `json:"host_io_limit_mb_sec,omitempty"`
	HostIOLimitIOSec    string `json:"host_io_limit_io_sec,omitempty"`
	DynamicDistribution string `json:"dynamicDistribution,omitempty"`
}

// RemoteSymmSGInfoParam have info abput remote symmetrix Id's and storage groups
type RemoteSymmSGInfoParam struct {
	RemoteSymmetrix1ID  string   `json:"remote_symmetrix_1_id,omitempty"`
	RemoteSymmetrix1SGs []string `json:"remote_symmetrix_1_sgs,omitempty"`
	RemoteSymmetrix2ID  string   `json:"remote_symmetrix_2_id,omitempty"`
	RemoteSymmetrix2SGs []string `json:"remote_symmetrix_2_sgs,omitempty"`
	Force               bool     `json:"force,omitempty"`
}

// RemoveVolumeParam holds volume ids to remove from SG
type RemoveVolumeParam struct {
	VolumeIDs             []string              `json:"volumeId,omitempty"`
	RemoteSymmSGInfoParam RemoteSymmSGInfoParam `json:"remoteSymmSGInfoParam,omitempty"`
}

// AddExistingStorageGroupParam contains SG ids and compliance alert flag
type AddExistingStorageGroupParam struct {
	StorageGroupIDs        []string `json:"storageGroupId,omitempty"`
	EnableComplianceAlerts bool     `json:"enableComplianceAlerts,omitempty"`
}

// CreateStorageGroupParam : Payload for creating Storage Group
type CreateStorageGroupParam struct {
	StorageGroupID            string                      `json:"storageGroupId,omitempty"`
	CreateEmptyStorageGroup   bool                        `json:"create_empty_storage_group,omitempty"`
	SRPID                     string                      `json:"srpId,omitempty"`
	SLOBasedStorageGroupParam []SLOBasedStorageGroupParam `json:"sloBasedStorageGroupParam,omitempty"`
	Emulation                 string                      `json:"emulation,omitempty"`
	ExecutionOption           string                      `json:"executionOption,omitempty"`
}

// SLOBasedStorageGroupParam holds parameters related to an SG and SLO
type SLOBasedStorageGroupParam struct {
	CustomCascadedStorageGroupID                   string                `json:"custom_cascaded_storageGroupId,omitempty"`
	SLOID                                          string                `json:"sloId,omitempty"`
	WorkloadSelection                              string                `json:"workloadSelection,omitempty"`
	VolumeAttributes                               []VolumeAttributeType `json:"volumeAttributes,omitempty"`
	AllocateCapacityForEachVol                     bool                  `json:"allocate_capacity_for_each_vol,omitempty"`
	PersistPrealloctedCapacityThroughReclaimOrCopy bool                  `json:"persist_preallocated_capacity_through_reclaim_or_copy,omitempty"`
	NoCompression                                  bool                  `json:"noCompression,omitempty"`
	EnableMobilityID                               string                `json:"enable_mobility_id,omitempty"`
	SetHostIOLimitsParam                           *SetHostIOLimitsParam `json:"setHostIOLimitsParam,omitempty"`
}

// AddNewStorageGroupParam contains parameters required to add a
// new storage group
type AddNewStorageGroupParam struct {
	SRPID                     string                      `json:"srpId,omitempty"`
	SLOBasedStorageGroupParam []SLOBasedStorageGroupParam `json:"sloBasedStorageGroupParam,omitempty"`
	Emulation                 string                      `json:"emulation,omitempty"`
	EnableComplianceAlerts    bool                        `json:"enableComplianceAlerts,omitempty"`
}

// SpecificVolumeParam holds volume ids, volume attributes and RDF group num
type SpecificVolumeParam struct {
	VolumeIDs       []string            `json:"volumeId,omitempty"`
	VolumeAttribute VolumeAttributeType `json:"volumeAttribute,omitempty"`
	RDFGroupNumber  int                 `json:"rdfGroupNumber,omitempty"`
}

// AllVolumeParam contains volume attributes and RDF group number
type AllVolumeParam struct {
	VolumeAttribute VolumeAttributeType `json:"volumeAttribute,omitempty"`
	RDFGroupNumber  int                 `json:"rdfGroupNumber,omitempty"`
}

// ExpandVolumesParam holds parameters to expand volumes
type ExpandVolumesParam struct {
	SpecificVolumeParam []SpecificVolumeParam `json:"specificVolumeParam,omitempty"`
	AllVolumeParam      AllVolumeParam        `json:"allVolumeParam,omitempty"`
}

// AddSpecificVolumeParam holds volume ids
type AddSpecificVolumeParam struct {
	VolumeIDs             []string              `json:"volumeId,omitempty"`
	RemoteSymmSGInfoParam RemoteSymmSGInfoParam `json:"remoteSymmSGInfoParam,omitempty"`
}

// ExpandStorageGroupParam holds params related to expanding size of an SG
type ExpandStorageGroupParam struct {
	AddExistingStorageGroupParam *AddExistingStorageGroupParam `json:"addExistingStorageGroupParam,omitempty"`
	AddNewStorageGroupParam      *AddNewStorageGroupParam      `json:"addNewStorageGroupParam,omitempty"`
	ExpandVolumesPar1Gam         *ExpandVolumesParam           `json:"expandVolumesParam,omitempty"`
	AddSpecificVolumeParam       *AddSpecificVolumeParam       `json:"addSpecificVolumeParam,omitempty"`
	AddVolumeParam               *AddVolumeParam               `json:"addVolumeParam,omitempty"`
}

// EditStorageGroupWorkloadParam holds selected work load
type EditStorageGroupWorkloadParam struct {
	WorkloadSelection string `json:"workloadSelection,omitempty,omitempty"`
}

// EditStorageGroupSLOParam hold param to change SLOs
type EditStorageGroupSLOParam struct {
	SLOID string `json:"sloId,omitempty"`
}

// EditStorageGroupSRPParam holds param to change SRPs
type EditStorageGroupSRPParam struct {
	SRPID string `json:"srpId,omitempty"`
}

// RemoveStorageGroupParam holds parameters to remove an SG
type RemoveStorageGroupParam struct {
	StorageGroupIDs []string `json:"storageGroupId,omitempty"`
	Force           bool     `json:"force,omitempty"`
}

// RenameStorageGroupParam holds new name of a storage group
type RenameStorageGroupParam struct {
	NewStorageGroupName string `json:"new_storage_Group_name,omitempty"`
}

// EditStorageGroupActionParam holds parameters to modify an SG
type EditStorageGroupActionParam struct {
	MergeStorageGroupParam        *MergeStorageGroupParam        `json:"mergeStorageGroupParam,omitempty"`
	SplitStorageGroupVolumesParam *SplitStorageGroupVolumesParam `json:"splitStorageGroupVolumesParam,omitempty"`
	SplitChildStorageGroupParam   *SplitChildStorageGroupParam   `json:"splitChildStorageGroupParam,omitempty"`
	MoveVolumeToStorageGroupParam *MoveVolumeToStorageGroupParam `json:"moveVolumeToStorageGroupParam,omitempty"`
	EditCompressionParam          *EditCompressionParam          `json:"editCompressionParam,omitempty"`
	SetHostIOLimitsParam          *SetHostIOLimitsParam          `json:"setHostIOLimitsParam,omitempty"`
	RemoveVolumeParam             *RemoveVolumeParam             `json:"removeVolumeParam,omitempty"`
	ExpandStorageGroupParam       *ExpandStorageGroupParam       `json:"expandStorageGroupParam,omitempty"`
	EditStorageGroupWorkloadParam *EditStorageGroupWorkloadParam `json:"editStorageGroupWorkloadParam,omitempty"`
	EditStorageGroupSLOParam      *EditStorageGroupSLOParam      `json:"editStorageGroupSLOParam,omitempty"`
	EditStorageGroupSRPParam      *EditStorageGroupSRPParam      `json:"editStorageGroupSRPParam,omitempty"`
	RemoveStorageGroupParam       *RemoveStorageGroupParam       `json:"removeStorageGroupParam,omitempty"`
	RenameStorageGroupParam       *RenameStorageGroupParam       `json:"renameStorageGroupParam,omitempty"`
}

// ExecutionOptionSynchronous : execute tasks synchronously
const ExecutionOptionSynchronous = "SYNCHRONOUS"

// ExecutionOptionAsynchronous : execute tasks asynchronously
const ExecutionOptionAsynchronous = "ASYNCHRONOUS"

// UpdateStorageGroupPayload : updates SG rest paylod
type UpdateStorageGroupPayload struct {
	EditStorageGroupActionParam EditStorageGroupActionParam `json:"editStorageGroupActionParam"`
	// ExecutionOption "SYNCHRONOUS" or "ASYNCHRONOUS"
	ExecutionOption string `json:"executionOption"`
	metadata        http.Header
}

// MetaData returns the metadata headers.
func (vp *UpdateStorageGroupPayload) MetaData() http.Header {
	if vp.metadata == nil {
		return make(http.Header)
	}
	return vp.metadata
}

// SetMetaData sets the metadata headers.
func (vp *UpdateStorageGroupPayload) SetMetaData(metadata http.Header) {
	vp.metadata = metadata
}
