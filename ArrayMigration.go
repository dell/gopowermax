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
	"time"

	types "github.com/dell/gopowermax/v2/types/v100"
	v100 "github.com/dell/gopowermax/v2/types/v100"
	log "github.com/sirupsen/logrus"
)

const (
	XEnvironment = "/environment/"
)

//ModifyMigrationSession does modification to storage group migration session
//this is used to do commit, sync, cut over on a migration session
func (c *Client) ModifyMigrationSession(ctx context.Context, localSymID, action, storageGroup string) (*types.MigrationSession, error) {
	defer c.TimeSpent("ModifyMigrationSession", time.Now())
	if _, err := c.IsAllowedArray(localSymID); err != nil {
		return nil, err
	}
	commitEnvPayload := &types.ModifyMigrationSessionRequest{
		Action: action,
	}
	ifDebugLogPayload(commitEnvPayload)
	URL := c.urlPrefix() + XMigration + SymmetrixX + localSymID + XStorageGroup + "/" + storageGroup
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()

	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodPut, URL, c.getDefaultHeaders(), commitEnvPayload)
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	modifiedMgrSession := new(types.MigrationSession)
	if err = decoder.Decode(modifiedMgrSession); err != nil {
		return nil, err
	}
	return modifiedMgrSession, nil
}

//CreateMigrationEnvironment validates existence of or creates migration environment between local and remote arrays
func (c *Client) CreateMigrationEnvironment(ctx context.Context, localSymID, remoteSymID string) (*types.MigrationEnv, error) {
	defer c.TimeSpent("GetOrCreateMigrationEnvironment", time.Now())
	if _, err := c.IsAllowedArray(localSymID); err != nil {
		return nil, err
	}

	createEnvPayload := &types.CreateMigrationEnv{
		OtherArrayId:    remoteSymID,
		ExecutionOption: types.ExecutionOptionSynchronous,
	}
	ifDebugLogPayload(createEnvPayload)
	URL := c.urlPrefix() + XMigration + SymmetrixX + localSymID

	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()

	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodPost, URL, c.getDefaultHeaders(), createEnvPayload)
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	migEnv := new(types.MigrationEnv)
	if err = decoder.Decode(migEnv); err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("Successfully created migration environment"))

	return migEnv, nil
}

//DeleteMigrationEnvironment validates existence of or creates migration environment between source and target arrays
func (c *Client) DeleteMigrationEnvironment(ctx context.Context, localSymID, remoteSymID string) error {
	defer c.TimeSpent("DeleteMigrationEnvironment", time.Now())

	if _, err := c.IsAllowedArray(localSymID); err != nil {
		return err
	}
	URL := c.urlPrefix() + XMigration + SymmetrixX + localSymID + XEnvironment + remoteSymID

	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()

	err := c.api.Delete(ctx, URL, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Debugf("error deleting migration env: %s", err.Error())
		return err
	}
	return err
}

func (c *Client) CreateSGMigration(ctx context.Context, localSymID, remoteSymID, storageGroup string) (*types.MigrationSession, error) {
	defer c.TimeSpent("CreateSGMigration", time.Now())
	if _, err := c.IsAllowedArray(localSymID); err != nil {
		return nil, err
	}
	sgMigrationPayload := types.CreateMigrationEnv{
		OtherArrayId:    remoteSymID,
		ExecutionOption: types.ExecutionOptionSynchronous,
	}
	ifDebugLogPayload(sgMigrationPayload)

	URL := c.urlPrefix() + XMigration + SymmetrixX + localSymID + "/" + storageGroup
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodPost, URL, c.getDefaultHeaders(), sgMigrationPayload)
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	// struct pending
	sgMig := new(types.MigrationSession)
	if err = decoder.Decode(sgMig); err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("SG migration successfully done"))

	return sgMig, nil
}

// MigrateStorageGroup creates a Storage Group given the storageGroupID (name), srpID (storage resource pool), service level, and boolean for thick volumes.
// If srpID is "None" then serviceLevel and thickVolumes settings are ignored
func (c *Client) MigrateStorageGroup(ctx context.Context, symID, storageGroupID, srpID, serviceLevel string, thickVolumes bool) (*v100.StorageGroup, error) {
	defer c.TimeSpent("MigrateStorageGroup", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := c.urlPrefix() + Migration + SymmetrixX + symID + XStorageGroup
	payload := c.GetCreateStorageGroupPayload(storageGroupID, srpID, serviceLevel, thickVolumes)
	ctx, cancel := c.GetTimeoutContext(ctx)
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
	log.Info(fmt.Sprintf("Successfully Migrated SG: %s", storageGroupID))
	return storageGroup, nil
}
