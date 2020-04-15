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

// Following structures are to in/out cast the Unisphere rest payload

// VolumeIDList : list of volume ids
type VolumeIDList struct {
	VolumeIDs string `json:"volumeId"`
}

// VolumeResultList : volume list resulted
type VolumeResultList struct {
	VolumeList []VolumeIDList `json:"result"`
	From       int            `json:"from"`
	To         int            `json:"to"`
}

// VolumeIterator : holds the iterator of resultant volume list
type VolumeIterator struct {
	ResultList VolumeResultList `json:"resultList"`
	ID         string           `json:"id"`
	Count      int              `json:"count"`
	// What units is ExpirationTime in?
	ExpirationTime int64 `json:"expirationTime"`
	MaxPageSize    int   `json:"maxPageSize"`
}

// Volume : information about a volume
type Volume struct {
	VolumeID              string   `json:"volumeID"`
	Type                  string   `json:"type"`
	Emulation             string   `json:"emulation"`
	SSID                  string   `json:"ssid"`
	AllocatedPercent      int      `json:"allocated_percent"`
	CapacityGB            float64  `json:"cap_gb"`
	FloatCapacityMB       float64  `json:"cap_mb"`
	CapacityCYL           int      `json:"cap_cyl"`
	Status                string   `json:"status"`
	Reserved              bool     `json:"reserved"`
	Pinned                bool     `json:"pinned"`
	PhysicalName          string   `json:"pysical_name"`
	VolumeIdentifier      string   `json:"volume_identifier"`
	WWN                   string   `json:"wwn"`
	Encapsulated          bool     `json:"encapsulated"`
	NumberOfStorageGroups int      `json:"num_of_storage_groups"`
	NumberOfFrontEndPaths int      `json:"num_of_front_end_paths"`
	StorageGroupIDList    []string `json:"storageGroupId"`
	// Don't know how to handle symmetrixPortKey for sure
	SymmetrixPortKey []SymmetrixPortKeyType `json:"symmetrixPortKey"`
	Success          bool                   `json:"success"`
	Message          string                 `json:"message"`
	SnapSource       bool                   `json:"snapvx_source"`
	SnapTarget       bool                   `json:"snapvx_target"`
}

// FreeVolumeParam : boolean value representing data to be freed
type FreeVolumeParam struct {
	FreeVolume bool `json:"free_volume"`
}

// ExpandVolumeParam : attributes to expand a volume
type ExpandVolumeParam struct {
	VolumeAttribute VolumeAttributeType `json:"volumeAttribute"`
	RDFGroupNumber  int                 `json:"rdfGroupNumber,omitempty"`
}

// ModifyVolumeIdentifierParam : volume identifier to modify the volume information
type ModifyVolumeIdentifierParam struct {
	VolumeIdentifier VolumeIdentifierType `json:"volumeIdentifier"`
}

// EditVolumeActionParam : action information to edit volume
type EditVolumeActionParam struct {
	FreeVolumeParam             *FreeVolumeParam             `json:"freeVolumeParam,omitempty"`
	ExpandVolumeParam           *ExpandVolumeParam           `json:"expandVolumeParam,omitempty"`
	ModifyVolumeIdentifierParam *ModifyVolumeIdentifierParam `json:"modifyVolumeIdentifierParam,omitempty"`
}

// EditVolumeParam : parameters required to edit volume information
type EditVolumeParam struct {
	EditVolumeActionParam EditVolumeActionParam `json:"editVolumeActionParam"`
	ExecutionOption       string                `json:"executionOption"`
}
