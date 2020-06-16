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

package pmax

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	types "github.com/dell/gopowermax/types/v90"
	types91 "github.com/dell/gopowermax/types/v91"
	log "github.com/sirupsen/logrus"
)

// The follow constants are for internal use within the pmax library.
const (
	SLOProvisioningX       = "sloprovisioning/"
	SymmetrixX             = "symmetrix/"
	IteratorX              = "common/Iterator/"
	XPage                  = "/page"
	XVolume                = "/volume"
	XStorageGroup          = "/storagegroup"
	XPortGroup             = "/portgroup"
	XInitiator             = "/initiator"
	XHost                  = "/host"
	XMaskingView           = "/maskingview"
	Emulation              = "FBA"
	MaxVolIdentifierLength = 64
)

var (
	// PmaxTimeout is the timeout value for pmax calls in sloprovisioning or system.
	// If Unisphere fails to answer after this period, an error will be returned.
	PmaxTimeout = 180 * time.Second
)

// GetTimeoutContext sets up a timeout of time PmaxTimeout for the returned context.
// The user caller should call the cancel function that is returned.
func GetTimeoutContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), PmaxTimeout)
	return ctx, cancel
}

//TimeSpent - Calculates and prints time spent for a caller function
func (c *Client) TimeSpent(functionName string, startTime time.Time) {
	if logResponseTimes {
		if functionName == "" {
			pc, _, _, ok := runtime.Caller(1)
			details := runtime.FuncForPC(pc)
			if ok && details != nil {
				functionName = details.Name()
			}
		}
		endTime := time.Now()
		log.Infof("pmax-time: %s took %.2f seconds to complete", functionName, endTime.Sub(startTime).Seconds())
	}
}

// GetVolumeIDsIterator returns a VolumeIDs Iterator. It generally fetches the first page in the result as part of the operation.
func (c *Client) GetVolumeIDsIterator(symID string, volumeIdentifierMatch string, like bool) (*types.VolumeIterator, error) {
	defer c.TimeSpent("GetVolumeIDsIterator", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	var query string
	if volumeIdentifierMatch != "" {
		if like {
			query = fmt.Sprintf("?volume_identifier=<like>%s", volumeIdentifierMatch)
		} else {
			query = fmt.Sprintf("?volume_identifier=%s", volumeIdentifierMatch)
		}
	}

	return c.getVolumeIDsIteratorBase(symID, query)
}

// GetVolumesInStorageGroupIterator returns a iterator of a list of volumes associated with a StorageGroup.
func (c *Client) GetVolumesInStorageGroupIterator(symID string, storageGroupID string) (*types.VolumeIterator, error) {
	var query string
	if storageGroupID == "" {
		return nil, fmt.Errorf("storageGroupID is empty")
	}

	query = fmt.Sprintf("?storageGroupId=%s", storageGroupID)
	return c.getVolumeIDsIteratorBase(symID, query)
}

// GetVolumeIDsIterator returns a VolumeIDs Iterator. It generally fetches the first page in the result as part of the operation.
func (c *Client) getVolumeIDsIteratorBase(symID string, query string) (*types.VolumeIterator, error) {
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XVolume
	if query != "" {
		URL = URL + query
	}

	ctx, cancel := GetTimeoutContext()
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetVolumeIDList failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	iter := &types.VolumeIterator{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(iter); err != nil {
		return nil, err
	}
	return iter, nil
}

// GetVolumeIDsIteratorPage fetches the next page of the iterator's result. From is the starting point. To can be left as 0, or can be set to the last element desired.
func (c *Client) GetVolumeIDsIteratorPage(iter *types.VolumeIterator, from, to int) ([]string, error) {
	defer c.TimeSpent("GetVolumeIDsIteratorPage", time.Now())
	if to == 0 || to-from+1 > iter.MaxPageSize {
		to = from + iter.MaxPageSize - 1
	}
	if to > iter.Count {
		to = iter.Count
	}
	queryParams := fmt.Sprintf("?from=%d&to=%d", from, to)
	URL := RESTPrefix + IteratorX + iter.ID + XPage + queryParams

	ctx, cancel := GetTimeoutContext()
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetVolumeIDsIteratorPage failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	result := &types.VolumeResultList{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(result); err != nil {
		return nil, err
	}

	volumeIDList := make([]string, to-from+1)
	for i := range result.VolumeList {
		volumeIDList[i] = result.VolumeList[i].VolumeIDs
	}
	return volumeIDList, nil
}

// DeleteVolumeIDsIterator deletes a volume iterator.
func (c *Client) DeleteVolumeIDsIterator(iter *types.VolumeIterator) error {
	defer c.TimeSpent("DeleteVolumeIDsIterator", time.Now())
	URL := RESTPrefix + IteratorX + iter.ID
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Delete(ctx, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		return err
	}
	return nil
}

// GetVolumeIDList gets a list of matching volume ids. If volumeIdentifierMatch is the empty string,
// all volumes are returned. Otherwise the volumes are filtered to volumes whose VolumeIdentifier
// exactly matches the volumeIdentfierMatch argument (when like is false), or whose VolumeIdentifier
// contains the volumeIdentifierMatch argument (when like is true).
func (c *Client) GetVolumeIDList(symID string, volumeIdentifierMatch string, like bool) ([]string, error) {
	defer c.TimeSpent("GetVolumeIDList", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	iter, err := c.GetVolumeIDsIterator(symID, volumeIdentifierMatch, like)
	if err != nil {
		return nil, err
	}
	return c.volumeIteratorToVolIDList(iter)
}

// GetVolumeIDListInStorageGroup - Gets a list of volume in a SG
func (c *Client) GetVolumeIDListInStorageGroup(symID string, storageGroupID string) ([]string, error) {
	iter, err := c.GetVolumesInStorageGroupIterator(symID, storageGroupID)
	if err != nil {
		return nil, err
	}
	return c.volumeIteratorToVolIDList(iter)
}

func (c *Client) volumeIteratorToVolIDList(iter *types.VolumeIterator) ([]string, error) {
	if iter.MaxPageSize < iter.Count {
		// The iterator only needs to be deleted if there are more entries than MaxPageSize?
		defer c.DeleteVolumeIDsIterator(iter)
	}

	// Get the initial results
	result := iter.ResultList
	volumeIDList := make([]string, len(result.VolumeList))
	for i := range result.VolumeList {
		volumeIDList[i] = result.VolumeList[i].VolumeIDs
	}

	// Iterate through addiional pages
	for from := result.To + 1; from <= iter.Count; {
		idlist, err := c.GetVolumeIDsIteratorPage(iter, from, 0)
		if err != nil {
			return nil, err
		}
		volumeIDList = append(volumeIDList, idlist...)
		from = from + len(idlist)
	}
	if len(volumeIDList) != iter.Count {
		return nil, fmt.Errorf("Expected %d ids but got %d ids", iter.Count, len(volumeIDList))
	}
	return volumeIDList, nil
}

// GetVolumeByID returns a Volume structure given the symmetrix and volume ID (volume ID is 5-digit hex field)
func (c *Client) GetVolumeByID(symID string, volumeID string) (*types.Volume, error) {
	defer c.TimeSpent("GetVolumeByID", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XVolume + "/" + volumeID
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetVolumeByID failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	volume := &types.Volume{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(volume); err != nil {
		return nil, err
	}
	return volume, nil
}

// GetStorageGroupIDList returns a list of StorageGroupIds in a StorageGroupIDList type.
func (c *Client) GetStorageGroupIDList(symID string) (*types.StorageGroupIDList, error) {
	defer c.TimeSpent("GetStorageGroupIDList", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XStorageGroup

	ctx, cancel := GetTimeoutContext()
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetStorageGroupIDList failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	sgIDList := &types.StorageGroupIDList{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(sgIDList); err != nil {
		return nil, err
	}
	return sgIDList, nil
}

//GetCreateStorageGroupPayload returns U4P payload for creating storage group
func (c *Client) GetCreateStorageGroupPayload(storageGroupID, srpID, serviceLevel string, thickVolumes bool) (payload interface{}) {
	workload := "None"
	if c.version == "90" {
		sloParams := []types.SLOBasedStorageGroupParam{}
		if srpID != "None" {
			sloParams = []types.SLOBasedStorageGroupParam{
				{
					SLOID:             serviceLevel,
					WorkloadSelection: workload,
					NumberOfVolumes:   0,
					VolumeAttribute: types.VolumeAttributeType{
						VolumeSize:   "0",
						CapacityUnit: "CYL",
					},
					AllocateCapacityForEachVol: thickVolumes,
					// compression not allowed with thick volumes
					NoCompression: thickVolumes,
				},
			}
		}
		createStorageGroupParam := &types.CreateStorageGroupParam{
			StorageGroupID:            storageGroupID,
			SRPID:                     srpID,
			Emulation:                 Emulation,
			ExecutionOption:           types.ExecutionOptionSynchronous,
			SLOBasedStorageGroupParam: sloParams,
		}
		return createStorageGroupParam
	}
	sloParams := []types91.SLOBasedStorageGroupParam{}
	if srpID != "None" {
		sloParams = []types91.SLOBasedStorageGroupParam{
			{
				SLOID:             serviceLevel,
				WorkloadSelection: workload,
				VolumeAttributes: []types91.VolumeAttributeType{
					{
						VolumeSize:      "0",
						CapacityUnit:    "CYL",
						NumberOfVolumes: 0,
					},
				},
				AllocateCapacityForEachVol: thickVolumes,
				// compression not allowed with thick volumes
				NoCompression: thickVolumes,
			},
		}
	}
	createStorageGroupParam := &types91.CreateStorageGroupParam{
		StorageGroupID:            storageGroupID,
		SRPID:                     srpID,
		Emulation:                 Emulation,
		ExecutionOption:           types91.ExecutionOptionSynchronous,
		SLOBasedStorageGroupParam: sloParams,
	}
	return createStorageGroupParam
}

// CreateStorageGroup creates a Storage Group given the storageGroupID (name), srpID (storage resource pool), service level, and boolean for thick volumes.
// If srpID is "None" then serviceLevel and thickVolumes settings are ignored
func (c *Client) CreateStorageGroup(symID, storageGroupID, srpID, serviceLevel string, thickVolumes bool) (*types.StorageGroup, error) {
	defer c.TimeSpent("CreateStorageGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XStorageGroup
	payload := c.GetCreateStorageGroupPayload(storageGroupID, srpID, serviceLevel, thickVolumes)
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodPost, URL, c.getDefaultHeaders(), payload)
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	storageGroup := &types.StorageGroup{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(storageGroup); err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully created SG: %s", storageGroupID))
	return storageGroup, nil
}

//DeleteStorageGroup deletes a storage group
func (c *Client) DeleteStorageGroup(symID string, storageGroupID string) error {
	defer c.TimeSpent("DeleteStorageGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XStorageGroup + "/" + storageGroupID
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Delete(ctx, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("DeleteStorageGroup failed: " + err.Error())
		return err
	}
	log.Info(fmt.Sprintf("Successfully deleted SG: %s", storageGroupID))
	return nil
}

//DeleteMaskingView deletes a storage group
func (c *Client) DeleteMaskingView(symID string, maskingViewID string) error {
	defer c.TimeSpent("DeleteMaskingView", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XMaskingView + "/" + maskingViewID
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Delete(ctx, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("DeleteMaskingView failed: " + err.Error())
		return err
	}
	log.Info(fmt.Sprintf("Successfully deleted Masking View: %s", maskingViewID))
	return nil
}

// GetStorageGroup returns a StorageGroup given the Symmetrix ID and Storage Group ID (which is really a name).
func (c *Client) GetStorageGroup(symID string, storageGroupID string) (*types.StorageGroup, error) {
	defer c.TimeSpent("GetStorageGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XStorageGroup + "/" + storageGroupID
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetStorageGroup failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	storageGroup := &types.StorageGroup{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(storageGroup); err != nil {
		return nil, err
	}
	return storageGroup, nil
}

// GetStoragePool returns a StoragePool given the Symmetrix ID and Storage Pool ID
func (c *Client) GetStoragePool(symID string, storagePoolID string) (*types.StoragePool, error) {
	defer c.TimeSpent("GetStoragePool", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + "/" + StorageResourcePool + "/" + storagePoolID
	storagePool := &types.StoragePool{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), storagePool)
	if err != nil {
		log.Error("GetStoragePool failed: " + err.Error())
		return nil, err
	}
	return storagePool, nil
}

// UpdateStorageGroup is a general method to update a StorageGroup (PUT operation) using a UpdateStorageGroupPayload.
func (c *Client) UpdateStorageGroup(symID string, storageGroupID string, payload interface{}) (*types.Job, error) {
	defer c.TimeSpent("UpdateStorageGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XStorageGroup + "/" + storageGroupID
	job := &types.Job{}
	fields := map[string]interface{}{
		http.MethodPut: URL,
	}

	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Put(
		ctx, URL, c.getDefaultHeaders(), payload, job)
	if err != nil {
		log.WithFields(fields).Error("Error in UpdateStorageGroup: " + err.Error())
		return nil, err
	}
	return job, nil
}

func ifDebugLogPayload(payload interface{}) {
	if Debug == false {
		return
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Error("could not Marshal json payload: " + err.Error())
	} else {
		log.Info("payload: " + string(payloadBytes))
	}
}

// CreateVolumeInStorageGroup creates a volume in the specified Storage Group with a given volumeName
// and the size of the volume in cylinders.
func (c *Client) CreateVolumeInStorageGroup(
	symID string, storageGroupID string, volumeName string, sizeInCylinders int) (*types.Volume, error) {
	defer c.TimeSpent("CreateVolumeInStorageGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}

	if len(volumeName) > MaxVolIdentifierLength {
		return nil, fmt.Errorf("Length of volumeName exceeds max limit")
	}

	job := &types.Job{}
	var err error
	payload := c.GetCreateVolInSGPayload(sizeInCylinders, volumeName)
	job, err = c.UpdateStorageGroup(symID, storageGroupID, payload)
	if err != nil || job == nil {
		return nil, fmt.Errorf("A job was not returned from UpdateStorageGroup")
	}
	job, err = c.WaitOnJobCompletion(symID, job.JobID)
	if err != nil {
		return nil, err
	}

	switch job.Status {
	case types.JobStatusFailed:
		return nil, fmt.Errorf("The UpdateStorageGroup job failed: " + c.JobToString(job))
	}

	// Look up the volume by the identifier.
	volIDList, err := c.GetVolumeIDList(symID, volumeName, false)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get Volume ID List: " + err.Error())
	}
	if len(volIDList) > 1 {
		log.Warning("Found multiple volumes matching the identifier " + volumeName)
	}
	for _, volumeID := range volIDList {
		vol, err := c.GetVolumeByID(symID, volumeID)
		if err == nil {
			for _, sgID := range vol.StorageGroupIDList {
				if sgID == storageGroupID && vol.CapacityCYL == sizeInCylinders {
					// Return the first match
					return vol, nil
				}
			}
		}
	}
	errormsg := fmt.Sprintf("Failed to find newly created volume with name: %s in SG: %s", volumeName, storageGroupID)
	log.Error(errormsg)
	return nil, fmt.Errorf(errormsg)
}

// ExpandVolume expands an existing volume to a new (larger) size in GB
func (c *Client) ExpandVolume(symID string, volumeID string, newSizeGB int) (*types.Volume, error) {
	payload := &types.EditVolumeParam{
		EditVolumeActionParam: types.EditVolumeActionParam{
			ExpandVolumeParam: &types.ExpandVolumeParam{
				VolumeAttribute: types.VolumeAttributeType{
					VolumeSize:   fmt.Sprintf("%d", newSizeGB),
					CapacityUnit: "GB",
				},
			},
		},
	}

	payload.ExecutionOption = types.ExecutionOptionSynchronous
	ifDebugLogPayload(payload)

	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XVolume + "/" + volumeID
	err := c.api.Put(context.Background(), URL, c.getDefaultHeaders(), payload, nil)

	var vol *types.Volume
	if err == nil {
		vol, err = c.GetVolumeByID(symID, volumeID)
	}

	return vol, err
}

// AddVolumesToStorageGroup adds one or more volumes (given by their volumeIDs) to a StorageGroup.
func (c *Client) AddVolumesToStorageGroup(symID string, storageGroupID string, volumeIDs ...string) error {
	defer c.TimeSpent("AddVolumesToStorageGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return err
	}
	// Check if the volume id list is not empty
	if len(volumeIDs) == 0 {
		return fmt.Errorf("At least one volume id has to be specified")
	}
	payload := c.GetAddVolumeToSGPayload(volumeIDs...)
	job, err := c.UpdateStorageGroup(symID, storageGroupID, payload)
	if err != nil || job == nil {
		return fmt.Errorf("A job was not returned from UpdateStorageGroup")
	}

	job, err = c.WaitOnJobCompletion(symID, job.JobID)
	if err != nil {
		return err
	}

	switch job.Status {
	case types.JobStatusFailed:
		return fmt.Errorf("The UpdateStorageGroup job failed: " + c.JobToString(job))
	}
	return nil
}

// RemoveVolumesFromStorageGroup removes one or more volumes (given by their volumeIDs) from a StorageGroup.
func (c *Client) RemoveVolumesFromStorageGroup(symID string, storageGroupID string, volumeIDs ...string) (*types.StorageGroup, error) {
	defer c.TimeSpent("RemoveVolumesFromStorageGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	payload := c.GetRemoveVolumeFromSGPayload(volumeIDs...)
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XStorageGroup + "/" + storageGroupID
	fields := map[string]interface{}{
		http.MethodPut: URL,
	}

	updatedStorageGroup := &types.StorageGroup{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Put(
		ctx, URL, c.getDefaultHeaders(), payload, updatedStorageGroup)
	if err != nil {
		log.WithFields(fields).Error("Error in RemoveVolumesFromStorageGroup: " + err.Error())
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully removed volumes: [%s] from SG: %s", strings.Join(volumeIDs, " "), storageGroupID))
	return updatedStorageGroup, nil
}

// GetCreateVolInSGPayload returns payload for adding volume/s to SG.
func (c *Client) GetCreateVolInSGPayload(sizeInCylinders int, volumeName string) (payload interface{}) {
	size := strconv.Itoa(sizeInCylinders)
	if c.version == "90" {
		addVolumeParam := &types.AddVolumeParam{
			NumberOfVols: 1,
			VolumeAttribute: types.VolumeAttributeType{
				VolumeSize:   size,
				CapacityUnit: "CYL",
			},
			// CreateNewVolumes: true,
			Emulation: "FBA",
			VolumeIdentifier: types.VolumeIdentifierType{
				VolumeIdentifierChoice: "identifier_name",
				IdentifierName:         volumeName,
			},
		}

		payload = &types.UpdateStorageGroupPayload{
			EditStorageGroupActionParam: types.EditStorageGroupActionParam{
				ExpandStorageGroupParam: &types.ExpandStorageGroupParam{
					AddVolumeParam: addVolumeParam,
				},
			},
			ExecutionOption: types.ExecutionOptionAsynchronous,
		}

	} else {
		addVolumeParam := &types91.AddVolumeParam{
			CreateNewVolumes: true,
			Emulation:        "FBA",
			VolumeAttributes: []types91.VolumeAttributeType{
				{
					NumberOfVolumes: 1,
					VolumeIdentifier: &types91.VolumeIdentifierType{
						VolumeIdentifierChoice: "identifier_name",
						IdentifierName:         volumeName,
					},
					CapacityUnit: "CYL",
					VolumeSize:   size,
				},
			},
		}

		payload = &types91.UpdateStorageGroupPayload{
			EditStorageGroupActionParam: types91.EditStorageGroupActionParam{
				ExpandStorageGroupParam: &types91.ExpandStorageGroupParam{
					AddVolumeParam: addVolumeParam,
				},
			},
			ExecutionOption: types91.ExecutionOptionAsynchronous,
		}
	}
	if payload != nil {
		ifDebugLogPayload(payload)
	}
	return payload
}

// GetAddVolumeToSGPayload returns payload for adding specific volume/s to SG.
func (c *Client) GetAddVolumeToSGPayload(volumeIDs ...string) (payload interface{}) {
	if c.version == "90" {
		addSpecificVolumeParam := &types.AddSpecificVolumeParam{
			VolumeIDs: volumeIDs,
		}
		payload = &types.UpdateStorageGroupPayload{
			EditStorageGroupActionParam: types.EditStorageGroupActionParam{
				ExpandStorageGroupParam: &types.ExpandStorageGroupParam{
					AddSpecificVolumeParam: addSpecificVolumeParam,
				},
			},
			ExecutionOption: types.ExecutionOptionAsynchronous,
		}
	} else {
		addSpecificVolumeParam := &types91.AddSpecificVolumeParam{
			VolumeIDs: volumeIDs,
		}
		payload = &types91.UpdateStorageGroupPayload{
			EditStorageGroupActionParam: types91.EditStorageGroupActionParam{
				ExpandStorageGroupParam: &types91.ExpandStorageGroupParam{
					AddSpecificVolumeParam: addSpecificVolumeParam,
				},
			},
			ExecutionOption: types91.ExecutionOptionAsynchronous,
		}
	}
	if payload != nil {
		ifDebugLogPayload(payload)
	}
	return payload
}

// GetRemoveVolumeFromSGPayload returns payload for removing volume/s from SG.
func (c *Client) GetRemoveVolumeFromSGPayload(volumeIDs ...string) (payload interface{}) {
	if c.version == "90" {
		removeVolumeParam := &types.RemoveVolumeParam{
			VolumeIDs: volumeIDs,
		}
		payload = &types.UpdateStorageGroupPayload{
			EditStorageGroupActionParam: types.EditStorageGroupActionParam{
				RemoveVolumeParam: removeVolumeParam,
			},
			ExecutionOption: types.ExecutionOptionSynchronous,
		}
	} else {
		removeVolumeParam := &types91.RemoveVolumeParam{
			VolumeIDs: volumeIDs,
		}
		payload = &types91.UpdateStorageGroupPayload{
			EditStorageGroupActionParam: types91.EditStorageGroupActionParam{
				RemoveVolumeParam: removeVolumeParam,
			},
			ExecutionOption: types91.ExecutionOptionSynchronous,
		}
	}
	if payload != nil {
		ifDebugLogPayload(payload)
	}
	return payload
}

// GetStoragePoolList returns a StoragePoolList object, which contains a list of all the Storage Pool names.
func (c *Client) GetStoragePoolList(symid string) (*types.StoragePoolList, error) {
	defer c.TimeSpent("GetStoragePoolList", time.Now())
	if _, err := c.IsAllowedArray(symid); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symid + "/" + StorageResourcePool
	spList := &types.StoragePoolList{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), spList)
	if err != nil {
		log.Error("GetStoragePoolList failed: " + err.Error())
		return nil, err
	}
	return spList, nil
}

// RenameVolume renames a volume.
func (c *Client) RenameVolume(symID string, volumeID string, newName string) (*types.Volume, error) {
	defer c.TimeSpent("RenameVolume", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	modifyVolumeIdentifierParam := &types.ModifyVolumeIdentifierParam{
		VolumeIdentifier: types.VolumeIdentifierType{
			VolumeIdentifierChoice: "identifier_name",
			IdentifierName:         newName,
		},
	}

	payload := &types.EditVolumeParam{
		EditVolumeActionParam: types.EditVolumeActionParam{
			ModifyVolumeIdentifierParam: modifyVolumeIdentifierParam,
		},
		ExecutionOption: types.ExecutionOptionSynchronous,
	}
	ifDebugLogPayload(payload)
	volume := &types.Volume{}

	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XVolume + "/" + volumeID
	fields := map[string]interface{}{
		http.MethodPut: URL,
		"VolumeID":     volumeID,
		"NewName":      newName,
	}
	log.WithFields(fields).Info("Renaming volume")
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Put(
		ctx, URL, c.getDefaultHeaders(), payload, volume)
	if err != nil {
		log.WithFields(fields).Error("Error in RenameVolume: " + err.Error())
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully renamed volume: %s", volumeID))
	return volume, nil
}

// DeleteVolume deletes a volume given the symmetrix ID and volume ID.
// Any storage tracks for the volume must have been previously deallocated using InitiateDeallocationOfTracksFromVolume,
// and the volume must not be a member of any Storage Group.
func (c *Client) DeleteVolume(symID string, volumeID string) error {
	defer c.TimeSpent("DeleteVolume", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XVolume + "/" + volumeID
	fields := map[string]interface{}{
		http.MethodPut: URL,
		"VolumeID":     volumeID,
	}
	log.WithFields(fields).Info("Deleting volume")
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Delete(ctx, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.WithFields(fields).Error("Error in DeleteVolume: " + err.Error())
	} else {
		log.Info(fmt.Sprintf("Successfully deleted volume: %s", volumeID))
	}
	return err
}

// InitiateDeallocationOfTracksFromVolume is an asynchrnous operation (that returns a job) to remove tracks from a volume.
func (c *Client) InitiateDeallocationOfTracksFromVolume(symID string, volumeID string) (*types.Job, error) {
	defer c.TimeSpent("InitiateDeallocationOfTracksFromVolume", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	freeVolumeParam := &types.FreeVolumeParam{
		FreeVolume: true,
	}
	payload := &types.EditVolumeParam{
		EditVolumeActionParam: types.EditVolumeActionParam{
			FreeVolumeParam: freeVolumeParam,
		},
		ExecutionOption: types.ExecutionOptionAsynchronous,
	}
	ifDebugLogPayload(payload)
	job := &types.Job{}

	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XVolume + "/" + volumeID
	fields := map[string]interface{}{
		http.MethodPut: URL,
		"VolumeID":     volumeID,
	}
	log.WithFields(fields).Info("Initiating track deletion...")
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Put(ctx, URL, c.getDefaultHeaders(), payload, job)
	if err != nil {
		log.WithFields(fields).Error("Error in InitiateDellocationOfTracksFromVolume: " + err.Error())
		return nil, err
	}
	return job, nil
}

// GetPortGroupList returns a PortGroupList object, which contains a list of the Port Groups
// which can be optionally filtered based on type
func (c *Client) GetPortGroupList(symID string, portGroupType string) (*types.PortGroupList, error) {
	defer c.TimeSpent("GetPortGroupList", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	filter := "?"
	if strings.EqualFold(portGroupType, "fibre") {
		filter += "fibre=true"
	} else if strings.EqualFold(portGroupType, "iscsi") {
		filter += "iscsi=true"
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XPortGroup
	if len(filter) > 1 {
		URL += filter
	}
	pgList := &types.PortGroupList{}

	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), pgList)
	if err != nil {
		log.Error("GetPortGrouplList failed: " + err.Error())
		return nil, err
	}
	return pgList, nil
}

// GetPortGroupByID returns a PortGroup given the Symmetrix ID and Port Group ID.
func (c *Client) GetPortGroupByID(symID string, portGroupID string) (*types.PortGroup, error) {
	defer c.TimeSpent("GetPortGroupByID", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XPortGroup + "/" + portGroupID
	portGroup := &types.PortGroup{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), portGroup)
	if err != nil {
		log.Error("GetPortGroupByID failed: " + err.Error())
		return nil, err
	}
	return portGroup, nil
}

// GetInitiatorList returns an InitiatorList object, which contains a list of all the Initiators.
// initiatorHBA, isISCSI, inHost are optional arguments which act as filters for the initiator list
func (c *Client) GetInitiatorList(symID string, initiatorHBA string, isISCSI bool, inHost bool) (*types.InitiatorList, error) {
	defer c.TimeSpent("GetInitiatorList", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	filter := "?"
	if inHost {
		if len(filter) > 1 {
			filter += "&"
		}
		filter += "in_a_host=true"
	}
	if initiatorHBA != "" {
		if len(filter) > 1 {
			filter += "&"
		}
		filter = filter + "initiator_hba=" + initiatorHBA
	}
	if isISCSI {
		if len(filter) > 1 {
			filter += "&"
		}
		filter += "iscsi=true"
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XInitiator
	if len(filter) > 1 {
		URL += filter
	}
	initList := &types.InitiatorList{}

	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), initList)
	if err != nil {
		log.Error("GetInitiatorList failed: " + err.Error())
		return nil, err
	}
	return initList, nil
}

// GetInitiatorByID returns an Initiator given the Symmetrix ID and Initiator ID.
func (c *Client) GetInitiatorByID(symID string, initID string) (*types.Initiator, error) {
	defer c.TimeSpent("GetInitiatorByID", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XInitiator + "/" + initID
	initiator := &types.Initiator{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), initiator)
	if err != nil {
		log.Error("GetInitiatorByID failed: " + err.Error())
		return nil, err
	}
	return initiator, nil
}

// GetHostList returns an HostList object, which contains a list of all the Hosts.
func (c *Client) GetHostList(symID string) (*types.HostList, error) {
	defer c.TimeSpent("GetHostList", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XHost
	hostList := &types.HostList{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), hostList)
	if err != nil {
		log.Error("GetHostList failed: " + err.Error())
		return nil, err
	}
	return hostList, nil
}

// GetHostByID returns a Host given the Symmetrix ID and Host ID.
func (c *Client) GetHostByID(symID string, hostID string) (*types.Host, error) {
	defer c.TimeSpent("GetHostByID", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XHost + "/" + hostID
	host := &types.Host{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), host)
	if err != nil {
		log.Error("GetHostByID failed: " + err.Error())
		return nil, err
	}
	return host, nil
}

// CreateHost creates a host from a list of InitiatorIDs (and optional HostFlags) return returns a types.Host.
// Initiator IDs do not contain the storage port designations, just the IQN string or FC WWN.
// Initiator IDs cannot be a member of more than one host.
func (c *Client) CreateHost(symID string, hostID string, initiatorIDs []string, hostFlags *types.HostFlags) (*types.Host, error) {
	defer c.TimeSpent("CreateHost", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	hostParam := &types.CreateHostParam{
		HostID:          hostID,
		InitiatorIDs:    initiatorIDs,
		HostFlags:       hostFlags,
		ExecutionOption: types.ExecutionOptionSynchronous,
	}
	host := &types.Host{}
	Debug = true
	ifDebugLogPayload(hostParam)
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XHost
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Post(ctx, URL, c.getDefaultHeaders(), hostParam, host)
	if err != nil {
		log.Error("CreateHost failed: " + err.Error())
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully created Host: %s", hostID))
	return host, nil
}

// UpdateHostInitiators updates a host from a list of InitiatorIDs and returns a types.Host.
func (c *Client) UpdateHostInitiators(symID string, host *types.Host, initiatorIDs []string) (*types.Host, error) {
	defer c.TimeSpent("UpdateHostInitiators", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	if host == nil {
		return nil, fmt.Errorf("Host can't be nil")
	}
	initRemove := []string{}
	initAdd := []string{}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XHost + "/" + host.HostID
	updatedHost := &types.Host{}

	// figure out which initiators are being added
	for _, init := range initiatorIDs {
		// if this initiator is not in the list of current initiators, add it
		if !stringInSlice(init, host.Initiators) {
			initAdd = append(initAdd, init)
		}
	}
	// check for initiators to be removed
	for _, init := range host.Initiators {
		if !stringInSlice(init, initiatorIDs) {
			initRemove = append(initRemove, init)
		}
	}

	ctx, cancel := GetTimeoutContext()
	defer cancel()
	// add initiators if needed
	if len(initAdd) > 0 {
		hostParam := &types.UpdateHostAddInitiatorsParam{}
		hostParam.EditHostAction = &types.AddHostInitiators{}
		hostParam.EditHostAction.AddInitiator = &types.ChangeInitiatorParam{}
		hostParam.EditHostAction.AddInitiator.Initiators = initAdd
		hostParam.ExecutionOption = types.ExecutionOptionSynchronous

		ifDebugLogPayload(hostParam)
		err := c.api.Put(ctx, URL, c.getDefaultHeaders(), hostParam, updatedHost)
		if err != nil {
			log.Error("UpdateHostInitiators failed: " + err.Error())
			return nil, err
		}
	}
	// remove initiators if needed
	if len(initRemove) > 0 {
		hostParam := &types.UpdateHostRemoveInititorsParam{}
		hostParam.EditHostAction = &types.RemoveHostInitiators{}
		hostParam.EditHostAction.RemoveInitiator = &types.ChangeInitiatorParam{}
		hostParam.EditHostAction.RemoveInitiator.Initiators = initRemove
		hostParam.ExecutionOption = types.ExecutionOptionSynchronous

		ifDebugLogPayload(hostParam)
		err := c.api.Put(ctx, URL, c.getDefaultHeaders(), hostParam, updatedHost)
		if err != nil {
			log.Error("UpdateHostInitiators failed: " + err.Error())
			return nil, err
		}
	}

	return updatedHost, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// DeleteHost deletes a host entry.
func (c *Client) DeleteHost(symID string, hostID string) error {
	defer c.TimeSpent("DeleteHost", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XHost + "/" + hostID
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Delete(ctx, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("DeleteHost failed: " + err.Error())
		return err
	}
	log.Info(fmt.Sprintf("Successfully deleted Host: %s", hostID))
	return nil
}

// GetMaskingViewList  returns a list of the MaskingView names.
func (c *Client) GetMaskingViewList(symID string) (*types.MaskingViewList, error) {
	defer c.TimeSpent("GetMaskingViewList", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XMaskingView
	mvList := &types.MaskingViewList{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), mvList)
	if err != nil {
		log.Error("GetMaskingViewList failed: " + err.Error())
		return nil, err
	}
	return mvList, nil
}

// GetMaskingViewByID returns a masking view given it's identifier (which is the name)
func (c *Client) GetMaskingViewByID(symID string, maskingViewID string) (*types.MaskingView, error) {
	defer c.TimeSpent("GetMaskingViewByID", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XMaskingView + "/" + maskingViewID
	mv := &types.MaskingView{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), mv)
	if err != nil {
		log.Error("GetMaskingViewByID failed: " + err.Error())
		return nil, err
	}
	return mv, nil
}

// GetMaskingViewConnections returns the connections of a masking view (optionally for a specific volume id.)
// Here volume id is the 5 digit volume ID.
func (c *Client) GetMaskingViewConnections(symID string, maskingViewID string, volumeID string) ([]*types.MaskingViewConnection, error) {
	defer c.TimeSpent("GetMaskingViewConnections", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XMaskingView + "/" + maskingViewID + "/connections"
	if volumeID != "" {
		URL = URL + "?volume_id=" + volumeID
	}
	cn := &types.MaskingViewConnectionsResult{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), cn)
	if err != nil {
		log.Error("GetMaskingViewConnections failed: " + err.Error())
		return nil, err
	}
	return cn.MaskingViewConnections, nil
}

// CreatePortGroup - Creates a Port Group
func (c *Client) CreatePortGroup(symID string, portGroupID string, dirPorts []types.PortKey) (*types.PortGroup, error) {
	defer c.TimeSpent("CreatePortGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XPortGroup
	createPortGroupParams := &types.CreatePortGroupParams{
		PortGroupID:      portGroupID,
		SymmetrixPortKey: dirPorts,
		ExecutionOption:  types.ExecutionOptionSynchronous,
	}
	ifDebugLogPayload(createPortGroupParams)
	portGroup := &types.PortGroup{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Post(ctx, URL, c.getDefaultHeaders(), createPortGroupParams, portGroup)
	if err != nil {
		log.Error("CreatePortGroup failed: " + err.Error())
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully created Port Group: %s", portGroupID))
	return portGroup, nil
}

// CreateMaskingView creates a masking view and returns the masking view object
func (c *Client) CreateMaskingView(symID string, maskingViewID string, storageGroupID string, hostOrhostGroupID string, isHost bool, portGroupID string) (*types.MaskingView, error) {
	defer c.TimeSpent("CreateMaskingView", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XMaskingView
	useExistingStorageGroupParam := &types.UseExistingStorageGroupParam{
		StorageGroupID: storageGroupID,
	}
	useExistingPortGroupParam := &types.UseExistingPortGroupParam{
		PortGroupID: portGroupID,
	}
	hostOrHostGroupSelection := &types.HostOrHostGroupSelection{}
	if isHost {
		hostOrHostGroupSelection.UseExistingHostParam = &types.UseExistingHostParam{
			HostID: hostOrhostGroupID,
		}
	} else {
		hostOrHostGroupSelection.UseExistingHostGroupParam = &types.UseExistingHostGroupParam{
			HostGroupID: hostOrhostGroupID,
		}
	}
	createMaskingViewParam := &types.MaskingViewCreateParam{
		MaskingViewID:            maskingViewID,
		HostOrHostGroupSelection: hostOrHostGroupSelection,
		PortGroupSelection: &types.PortGroupSelection{
			UseExistingPortGroupParam: useExistingPortGroupParam,
		},
		StorageGroupSelection: &types.StorageGroupSelection{
			UseExistingStorageGroupParam: useExistingStorageGroupParam,
		},
	}
	ifDebugLogPayload(createMaskingViewParam)
	maskingView := &types.MaskingView{}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Post(ctx, URL, c.getDefaultHeaders(), createMaskingViewParam, maskingView)
	if err != nil {
		log.Error("CreateMaskingView failed: " + err.Error())
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully created Masking View: %s", maskingViewID))
	return maskingView, nil
}

// DeletePortGroup - Deletes a PG
func (c *Client) DeletePortGroup(symID string, portGroupID string) error {
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XPortGroup + "/" + portGroupID

	err := c.api.Delete(context.Background(), URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("DeletePortGroup failed: " + err.Error())
		return err
	}
	return nil
}

// UpdatePortGroup - Update the PortGroup based on the 'ports' slice. The slice represents the intended
// configuration of the PortGroup after _successful_ completion of the request.
// NB: based on the passed in 'ports' the implementation will determine how to update
// the PortGroup and make appropriate REST calls sequentially. Take this into
// consideration when making parallel calls.
func (c *Client) UpdatePortGroup(symID string, portGroupID string, ports []types.PortKey) (*types.PortGroup, error) {
	URL := c.urlPrefix() + SLOProvisioningX + SymmetrixX + symID + XPortGroup + "/" + portGroupID
	fmt.Println(URL)

	// Create map of string "<DIRECTOR ID>/<PORT ID>" to a SymmetrixPortKeyType object based on the passed in 'ports'
	inPorts := make(map[string]*types.SymmetrixPortKeyType)
	for _, port := range ports {
		director := strings.ToUpper(port.DirectorID)
		port := strings.ToLower(port.PortID)
		key := fmt.Sprintf("%s/%s", director, port)
		if inPorts[key] == nil {
			inPorts[key] = &types.SymmetrixPortKeyType{
				DirectorID: director,
				PortID:     port,
			}
		}
	}

	pg, err := c.GetPortGroupByID(symID, portGroupID)
	if err != nil {
		log.Error("Could not get portGroup: " + err.Error())
		return nil, err
	}

	portIDRegex, _ := regexp.Compile("\\w+:(\\d+)")

	// Create map of string "<DIRECTOR ID>/<PORT ID>" to a SymmetrixPortKeyType object based on what's found
	// in the PortGroup
	pgPorts := make(map[string]*types.SymmetrixPortKeyType)
	for _, p := range pg.SymmetrixPortKey {
		director := strings.ToUpper(p.DirectorID)
		// PortID string may come as a combination of directory + port_number
		// Extract just the port_number part
		port := strings.ToLower(p.PortID)
		submatch := portIDRegex.FindAllStringSubmatch(port, -1)
		if len(submatch) > 0 {
			port = submatch[0][1]
		}
		key := fmt.Sprintf("%s/%s", director, port)
		pgPorts[key] = &types.SymmetrixPortKeyType{
			DirectorID: director,
			PortID:     port,
		}
	}

	// Diff ports in request with ones in PortGroup --> ports to add
	var added []types.SymmetrixPortKeyType
	for k, v := range inPorts {
		if pgPorts[k] == nil {
			added = append(added, *v)
		}
	}

	// Diff ports in PortGroup with ones in request --> ports to remove
	var removed []types.SymmetrixPortKeyType
	for k, v := range pgPorts {
		if inPorts[k] == nil {
			removed = append(removed, *v)
		}
	}

	if len(added) > 0 {
		log.Info(fmt.Sprintf("Adding ports %v", added))
		edit := &types.EditPortGroupActionParam{
			AddPortParam: &types.AddPortParam{
				Ports: added,
			},
		}
		add := types.EditPortGroup{
			EditPortGroupActionParam: edit,
		}
		err := c.api.Put(context.Background(), URL, c.getDefaultHeaders(), add, &pg)
		if err != nil {
			log.Error("UpdatePortGroup failed when trying to add ports: " + err.Error())
			return nil, err
		}
	}

	if len(removed) > 0 {
		log.Info(fmt.Sprintf("Removing ports %v", removed))
		edit := &types.EditPortGroupActionParam{
			RemovePortParam: &types.RemovePortParam{
				Ports: removed,
			},
		}
		remove := types.EditPortGroup{
			EditPortGroupActionParam: edit,
		}
		err := c.api.Put(context.Background(), URL, c.getDefaultHeaders(), remove, &pg)
		if err != nil {
			log.Error("UpdatePortGroup failed when trying to remove ports: " + err.Error())
			return nil, err
		}
	}

	return pg, nil
}
