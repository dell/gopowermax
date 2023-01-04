/*
 Copyright © 2023 Dell Inc. or its subsidiaries. All Rights Reserved.

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
	types "github.com/dell/gopowermax/v2/types/v100"
	"net/http"
	"time"
)

// The following constants are for the query of performance metrics for pmax
const (
	Average      = "Average"
	Performance  = "performance"
	StorageGroup = "/StorageGroup"
	Volume       = "/Volume"
	Metrics      = "/metrics"
)

// GetStorageGroupMetrics returns a list of Storage Group performance metrics
func (c *Client) GetStorageGroupMetrics(ctx context.Context, symID string, storageGroupID string, metricsQuery []string) (*types.StorageGroupMetricsIterator, error) {
	defer c.TimeSpent("GetStorageGroupMetrics", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := RESTPrefix + Performance + StorageGroup + Metrics
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	params := types.StorageGroupMetricsParam{
		SymmetrixID:    symID,
		StartDate:      time.Now().UnixMilli() - 300000,
		EndDate:        time.Now().UnixMilli(),
		DataFormat:     Average,
		StorageGroupID: storageGroupID,
		Metrics:        metricsQuery,
	}
	resp, err := c.api.DoAndGetResponseBody(ctx, http.MethodPost, URL, c.getDefaultHeaders(), params)
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}
	metricsList := &types.StorageGroupMetricsIterator{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(metricsList); err != nil {
		return nil, err
	}
	return metricsList, nil
}

// GetVolumesMetrics returns a list of Volume performance metrics
func (c *Client) GetVolumesMetrics(ctx context.Context, symID string, storageGroups string, metricsQuery []string) (*types.VolumeMetricsIterator, error) {
	defer c.TimeSpent("GetStorageGroupMetrics", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := RESTPrefix + Performance + Volume + Metrics
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	params := types.VolumeMetricsParam{
		SystemID:                       symID,
		StartDate:                      time.Now().UnixMilli() - 300000,
		EndDate:                        time.Now().UnixMilli(),
		DataFormat:                     Average,
		CommaSeparatedStorageGroupList: storageGroups,
		Metrics:                        metricsQuery,
	}
	resp, err := c.api.DoAndGetResponseBody(ctx, http.MethodPost, URL, c.getDefaultHeaders(), params)
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}
	metricsList := &types.VolumeMetricsIterator{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(metricsList); err != nil {
		return nil, err
	}
	return metricsList, nil
}
