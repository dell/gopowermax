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

	v100 "github.com/dell/gopowermax/v2/types/v100"

	log "github.com/sirupsen/logrus"
)

// The following constants are for internal use within the pmax library.
const (
	XRDFGroup = "/rdf_group"
	ASYNC     = "ASYNC"
	METRO     = "METRO"
	SYNC      = "SYNC"
)

// GetRDFGroup returns RDF group information given the RDF group number
func (c *Client) GetRDFGroup(ctx context.Context, symID, rdfGroupNo string) (*v100.RDFGroup, error) {
	defer c.TimeSpent("GetRdfGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XRDFGroup + "/" + rdfGroupNo
	resp, err := c.api.DoAndGetResponseBody(ctx, http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetRdfGroup failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	rdfGrpInfo := new(v100.RDFGroup)
	if err := json.NewDecoder(resp.Body).Decode(rdfGrpInfo); err != nil {
		return nil, err
	}
	return rdfGrpInfo, nil
}

// GetProtectedStorageGroup returns protected storage group given the storage group ID
func (c *Client) GetProtectedStorageGroup(ctx context.Context, symID, storageGroup string) (*v100.RDFStorageGroup, error) {
	defer c.TimeSpent("GetProtectedStorageGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XStorageGroup + "/" + storageGroup
	resp, err := c.api.DoAndGetResponseBody(ctx, http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetProtectedStorageGroup failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	rdfSgInfo := new(v100.RDFStorageGroup)
	if err := json.NewDecoder(resp.Body).Decode(rdfSgInfo); err != nil {
		return nil, err
	}
	return rdfSgInfo, nil
}

// ExecuteReplicationActionOnSG executes supported replication based actions on the protected SG
func (c *Client) ExecuteReplicationActionOnSG(ctx context.Context, symID, action, storageGroup, rdfGroup string, force, exemptConsistency, bias bool) error {
	defer c.TimeSpent("ExecuteReplicationActionOnSG", time.Now())

	if _, err := c.IsAllowedArray(symID); err != nil {
		return err
	}

	modifyParam := &v100.ModifySGRDFGroup{}

	switch action {
	case "Establish":
		actionParam := &v100.Establish{
			Force:    force,
			SymForce: false,
			Star:     false,
			Hop2:     false,
		}
		if bias {
			actionParam.MetroBias = true
		}
		modifyParam = &v100.ModifySGRDFGroup{
			Establish:       actionParam,
			Action:          action,
			ExecutionOption: v100.ExecutionOptionSynchronous,
		}
	case "Suspend":
		actionParam := &v100.Suspend{
			Force:      force,
			SymForce:   false,
			Star:       false,
			Hop2:       false,
			Immediate:  false,
			ConsExempt: exemptConsistency,
		}
		modifyParam = &v100.ModifySGRDFGroup{
			Suspend:         actionParam,
			Action:          action,
			ExecutionOption: v100.ExecutionOptionSynchronous,
		}
	case "Resume":
		actionParam := &v100.Resume{
			Force:        force,
			SymForce:     false,
			Star:         false,
			Hop2:         false,
			Remote:       false,
		}
		modifyParam = &v100.ModifySGRDFGroup{
			Resume:          actionParam,
			Action:          action,
			ExecutionOption: v100.ExecutionOptionSynchronous,
		}
	case "Failback":
		actionParam := &v100.Failback{
			Force:        force,
			SymForce:     false,
			Star:         false,
			Hop2:         false,
			Bypass:       false,
			Remote:       false,
		}
		modifyParam = &v100.ModifySGRDFGroup{
			Failback:        actionParam,
			Action:          action,
			ExecutionOption: v100.ExecutionOptionSynchronous,
		}
	case "Failover":
		actionParam := &v100.Failover{
			Force:     force,
			SymForce:  false,
			Star:      false,
			Hop2:      false,
			Bypass:    false,
			Remote:    false,
			Immediate: false,
			Establish: false,
			Restore:   false,
		}
		modifyParam = &v100.ModifySGRDFGroup{
			Failover:        actionParam,
			Action:          action,
			ExecutionOption: v100.ExecutionOptionSynchronous,
		}
	case "Swap":
		actionParam := &v100.Swap{
			Force:     force,
			SymForce:  false,
			Star:      false,
			Hop2:      false,
			Bypass:    false,
			HalfSwap:  false,
			RefreshR1: false,
			RefreshR2: false,
		}
		modifyParam = &v100.ModifySGRDFGroup{
			Swap:            actionParam,
			Action:          action,
			ExecutionOption: v100.ExecutionOptionSynchronous,
		}
	default:
		return fmt.Errorf("not a supported action on a protected storage group")
	}
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XStorageGroup + "/" + storageGroup + XRDFGroup + "/" + rdfGroup
	fields := map[string]interface{}{
		http.MethodPut: URL,
	}
	ctx, cancel := c.GetTimeoutContext(ctx)
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
func (c *Client) GetCreateSGReplicaPayload(remoteSymID string, rdfMode string, rdfgNo int, remoteSGName string, remoteServiceLevel string, establish, bias bool) *v100.CreateSGSRDF {

	var payload *v100.CreateSGSRDF
	switch rdfMode {
	case ASYNC:
		payload = &v100.CreateSGSRDF{
			ReplicationMode:        "Asynchronous",
			RemoteSLO:              remoteServiceLevel,
			RemoteSymmID:           remoteSymID,
			RdfgNumber:             rdfgNo,
			RemoteStorageGroupName: remoteSGName,
			Establish:              establish,
			ExecutionOption:        v100.ExecutionOptionSynchronous,
		}
	case SYNC:
		payload = &v100.CreateSGSRDF{
			ReplicationMode:        "Synchronous",
			RemoteSLO:              remoteServiceLevel,
			RemoteSymmID:           remoteSymID,
			RdfgNumber:             rdfgNo,
			RemoteStorageGroupName: remoteSGName,
			Establish:              establish,
			ExecutionOption:        v100.ExecutionOptionSynchronous,
		}
	case METRO:
		payload = &v100.CreateSGSRDF{
			ReplicationMode:        "Active",
			RemoteSLO:              remoteServiceLevel,
			RemoteSymmID:           remoteSymID,
			RdfgNumber:             rdfgNo,
			RemoteStorageGroupName: remoteSGName,
			Establish:              establish,
			MetroBias:              bias,
			ExecutionOption:        v100.ExecutionOptionSynchronous,
		}
	}
	return payload
}

// CreateSGReplica creates a storage group on remote array and protect them with given RDF Mode and a given source storage group
func (c *Client) CreateSGReplica(ctx context.Context, symID, remoteSymID, rdfMode, rdfGroupNo, sourceSG, remoteSGName, remoteServiceLevel string, bias bool) (*v100.SGRDFInfo, error) {
	defer c.TimeSpent("CreateSGReplica", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	rdfgNo, _ := strconv.Atoi(rdfGroupNo)
	createSGReplicaPayload := c.GetCreateSGReplicaPayload(remoteSymID, rdfMode, rdfgNo, remoteSGName, remoteServiceLevel, true, bias)
	Debug = true
	ifDebugLogPayload(createSGReplicaPayload)
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XStorageGroup + "/" + sourceSG + XRDFGroup

	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodPost, URL, c.getDefaultHeaders(), createSGReplicaPayload)
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rdfSG := &v100.SGRDFInfo{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(rdfSG); err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully created SG replica for %s", sourceSG))
	return rdfSG, nil
}

// GetCreateRDFPairPayload returns payload for adding a replication pair based on replication mode
func (c *Client) GetCreateRDFPairPayload(devList v100.LocalDeviceListCriteria, rdfMode, rdfType string, establish, exemptConsistency bool) *v100.CreateRDFPair {

	var payload *v100.CreateRDFPair
	switch rdfMode {
	case ASYNC:
		payload = &v100.CreateRDFPair{
			RdfMode:                 "Asynchronous",
			RdfType:                 rdfType,
			Establish:               establish,
			Exempt:                  exemptConsistency,
			LocalDeviceListCriteria: &devList,
			ExecutionOption:         v100.ExecutionOptionSynchronous,
		}
	case SYNC:
		payload = &v100.CreateRDFPair{
			RdfMode:                 "Synchronous",
			RdfType:                 rdfType,
			Establish:               establish,
			Exempt:                  exemptConsistency,
			LocalDeviceListCriteria: &devList,
			ExecutionOption:         v100.ExecutionOptionSynchronous,
		}
	case METRO:
		payload = &v100.CreateRDFPair{
			RdfMode:                 "Active",
			RdfType:                 "RDF1",
			Bias:                    true,
			Establish:               establish,
			Exempt:                  exemptConsistency,
			LocalDeviceListCriteria: &devList,
			ExecutionOption:         v100.ExecutionOptionSynchronous,
		}
	}
	return payload
}

// CreateRDFPair creates an RDF device pair in the given RDF group
func (c *Client) CreateRDFPair(ctx context.Context, symID, rdfGroupNo, deviceID, rdfMode, rdfType string, establish, exemptConsistency bool) (*v100.RDFDevicePairList, error) {
	defer c.TimeSpent("CreateRDFPair", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	var deviceList []string
	deviceList = append(deviceList, deviceID)
	devList := v100.LocalDeviceListCriteria{
		LocalDeviceList: deviceList,
	}
	createPairPayload := c.GetCreateRDFPairPayload(devList, rdfMode, rdfType, establish, exemptConsistency)
	Debug = true
	ifDebugLogPayload(createPairPayload)
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XRDFGroup + "/" + rdfGroupNo + XVolume + "/" + deviceID

	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodPost, URL, c.getDefaultHeaders(), createPairPayload)
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rdfPairList := &v100.RDFDevicePairList{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(rdfPairList); err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully created volume replica for %s", deviceID))
	return rdfPairList, nil
}

// GetRDFDevicePairInfo returns RDF volume information
func (c *Client) GetRDFDevicePairInfo(ctx context.Context, symID, rdfGroup, volumeID string) (*v100.RDFDevicePair, error) {
	defer c.TimeSpent("GetRDFDevicePairInfo", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}

	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XRDFGroup + "/" + rdfGroup + XVolume + "/" + volumeID
	resp, err := c.api.DoAndGetResponseBody(ctx, http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetRDFDevicePairInfo failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	rdfDevPairInfo := new(v100.RDFDevicePair)
	if err := json.NewDecoder(resp.Body).Decode(rdfDevPairInfo); err != nil {
		return nil, err
	}
	return rdfDevPairInfo, nil
}

// GetStorageGroupRDFInfo returns the of RDF info of protected storage group
func (c *Client) GetStorageGroupRDFInfo(ctx context.Context, symID, sgName, rdfGroupNo string) (*v100.StorageGroupRDFG, error) {
	defer c.TimeSpent("GetStorageGroupRDFInfo", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}

	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	URL := c.urlPrefix() + ReplicationX + SymmetrixX + symID + XStorageGroup + "/" + sgName + XRDFGroup + "/" + rdfGroupNo
	resp, err := c.api.DoAndGetResponseBody(ctx, http.MethodGet, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetStorageGroupRDFInfo failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	sgRdfInfo := new(v100.StorageGroupRDFG)
	if err := json.NewDecoder(resp.Body).Decode(sgRdfInfo); err != nil {
		return nil, err
	}
	return sgRdfInfo, nil
}
