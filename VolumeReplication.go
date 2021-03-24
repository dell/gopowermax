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
	"strconv"
	"time"

	types "github.com/dell/gopowermax/types/v90"
	log "github.com/sirupsen/logrus"
)

// The following constants are for internal use within the pmax library.
const (
	XRDFGroup = "/rdf_group"
	ASYNC     = "ASYNC"
)

// GetRDFGroup returns RDF group information given the RDF group number
func (c *Client) GetRDFGroup(symID, rdfGroupNo string) (*types.RDFGroup, error) {
	defer c.TimeSpent("GetRdfGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XRDFGroup + "/" + rdfGroupNo
	resp, err := c.api.DoAndGetResponseBody(context.Background(), http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetRdfGroup failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	rdfGrpInfo := new(types.RDFGroup)
	if err := json.NewDecoder(resp.Body).Decode(rdfGrpInfo); err != nil {
		return nil, err
	}
	return rdfGrpInfo, nil
}

// GetProtectedStorageGroup returns protected storage group given the storage group ID
func (c *Client) GetProtectedStorageGroup(symID, storageGroup string) (*types.RDFStorageGroup, error) {
	defer c.TimeSpent("GetProtectedStorageGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XStorageGroup + "/" + storageGroup
	resp, err := c.api.DoAndGetResponseBody(context.Background(), http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetProtectedStorageGroup failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	rdfSgInfo := new(types.RDFStorageGroup)
	if err := json.NewDecoder(resp.Body).Decode(rdfSgInfo); err != nil {
		return nil, err
	}
	return rdfSgInfo, nil
}

// ExecuteReplicationActionOnSG executes supported replication based actions on the protected SG
func (c *Client) ExecuteReplicationActionOnSG(symID, action, storageGroup, rdfGroup string, force, exemptConsistency bool) error {
	defer c.TimeSpent("ExecuteReplicationActionOnSG", time.Now())

	if _, err := c.IsAllowedArray(symID); err != nil {
		return err
	}

	modifyParam := &types.ModifySGRDFGroup{}

	switch action {
	case "Suspend":
		actionParam := &types.Suspend{
			Force:      force,
			SymForce:   false,
			Star:       false,
			Hop2:       false,
			Immediate:  false,
			ConsExempt: exemptConsistency,
		}
		modifyParam = &types.ModifySGRDFGroup{
			Suspend:         actionParam,
			Action:          action,
			ExecutionOption: types.ExecutionOptionSynchronous,
		}
	case "Resume":
		actionParam := &types.Resume{
			Force:        force,
			SymForce:     false,
			Star:         false,
			Hop2:         false,
			Remote:       false,
			RecoverPoint: false,
		}
		modifyParam = &types.ModifySGRDFGroup{
			Resume:          actionParam,
			Action:          action,
			ExecutionOption: types.ExecutionOptionSynchronous,
		}
	default:
		return fmt.Errorf("not a supported action on a protected storage group")
	}
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XStorageGroup + "/" + storageGroup + XRDFGroup + "/" + rdfGroup
	fields := map[string]interface{}{
		http.MethodPut: URL,
	}
	ctx, cancel := GetTimeoutContext()
	defer cancel()
	err := c.api.Put(
		ctx, URL, c.getDefaultHeaders(), modifyParam, nil)
	if err != nil {
		log.WithFields(fields).Error("Error in ExecuteReplicationActionOnSG: " + err.Error())
		return err
	}
	log.Info(fmt.Sprintf("Action (%s) on protected StorageGroup (%s) with RDF group (%s) is successful", action, storageGroup, rdfGroup))
	return nil
}

// GetCreateSGReplicaPayload returns a payload to create a storage group on remote array from local array and protect it with rdfgNo
func (c *Client) GetCreateSGReplicaPayload(remoteSymID string, rdfMode string, rdfgNo int, remoteSGName string, remoteServiceLevel string, establish bool) *types.CreateSGSRDF {

	var payload *types.CreateSGSRDF
	if rdfMode == ASYNC {
		payload = &types.CreateSGSRDF{
			ReplicationMode:        "Asynchronous",
			RemoteSLO:              remoteServiceLevel,
			RemoteSymmID:           remoteSymID,
			RdfgNumber:             rdfgNo,
			RemoteStorageGroupName: remoteSGName,
			Establish:              establish,
			ExecutionOption:        types.ExecutionOptionSynchronous,
		}
	}
	return payload
}

// CreateSGReplica creates a storage group on remote array and protect them with given RDF Mode and a given source storage group
func (c *Client) CreateSGReplica(symID, remoteSymID, rdfMode, rdfGroupNo, sourceSG, remoteSGName, remoteServiceLevel string) (*types.SGRDFInfo, error) {
	defer c.TimeSpent("CreateSGReplica", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	rdfgNo, _ := strconv.Atoi(rdfGroupNo)
	createSGReplicaPayload := c.GetCreateSGReplicaPayload(remoteSymID, rdfMode, rdfgNo, remoteSGName, remoteServiceLevel, true)
	Debug = true
	ifDebugLogPayload(createSGReplicaPayload)
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XStorageGroup + "/" + sourceSG + XRDFGroup

	ctx, cancel := GetTimeoutContext()
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodPost, URL, c.getDefaultHeaders(), createSGReplicaPayload)
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rdfSG := &types.SGRDFInfo{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(rdfSG); err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully created SG replica for %s", sourceSG))
	return rdfSG, nil
}

// GetCreateRDFPairPayload returns payload for adding a replication pair based on replication mode
func (c *Client) GetCreateRDFPairPayload(devList types.LocalDeviceListCriteria, rdfMode, rdfType string, establish, exemptConsistency bool) *types.CreateRDFPair {

	var payload *types.CreateRDFPair
	if rdfMode == ASYNC {
		payload = &types.CreateRDFPair{
			RdfMode:                 "Asynchronous",
			RdfType:                 rdfType,
			Establish:               establish,
			Exempt:                  exemptConsistency,
			LocalDeviceListCriteria: &devList,
			ExecutionOption:         types.ExecutionOptionSynchronous,
		}
	}
	return payload
}

// CreateRDFPair creates an RDF device pair in the given RDF group
func (c *Client) CreateRDFPair(symID, rdfGroupNo, deviceID, rdfMode, rdfType string, establish, exemptConsistency bool) (*types.RDFDevicePairList, error) {
	defer c.TimeSpent("CreateRDFPair", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	var deviceList []string
	deviceList = append(deviceList, deviceID)
	devList := types.LocalDeviceListCriteria{
		LocalDeviceList: deviceList,
	}
	createPairPayload := c.GetCreateRDFPairPayload(devList, rdfMode, rdfType, establish, exemptConsistency)
	Debug = true
	ifDebugLogPayload(createPairPayload)
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XRDFGroup + "/" + rdfGroupNo + XVolume + "/" + deviceID

	ctx, cancel := GetTimeoutContext()
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodPost, URL, c.getDefaultHeaders(), createPairPayload)
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rdfPairList := &types.RDFDevicePairList{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(rdfPairList); err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully created volume replica for %s", deviceID))
	return rdfPairList, nil
}

// GetRDFDevicePairInfo returns RDF volume information
func (c *Client) GetRDFDevicePairInfo(symID, rdfGroup, volumeID string) (*types.RDFDevicePair, error) {
	defer c.TimeSpent("GetRDFDevicePairInfo", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XRDFGroup + "/" + rdfGroup + XVolume + "/" + volumeID
	resp, err := c.api.DoAndGetResponseBody(context.Background(), http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetRDFDevicePairInfo failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	rdfDevPairInfo := new(types.RDFDevicePair)
	if err := json.NewDecoder(resp.Body).Decode(rdfDevPairInfo); err != nil {
		return nil, err
	}
	return rdfDevPairInfo, nil
}

// GetStorageGroupRDFInfo returns the of RDF info of protected storage group
func (c *Client) GetStorageGroupRDFInfo(symID, sgName, rdfGroupNo string) (*types.StorageGroupRDFG, error) {
	defer c.TimeSpent("GetStorageGroupRDFInfo", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XStorageGroup + "/" + sgName + XRDFGroup + "/" + rdfGroupNo
	resp, err := c.api.DoAndGetResponseBody(context.Background(), http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetStorageGroupRDFInfo failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	sgRdfInfo := new(types.StorageGroupRDFG)
	if err := json.NewDecoder(resp.Body).Decode(sgRdfInfo); err != nil {
		return nil, err
	}
	return sgRdfInfo, nil
}
