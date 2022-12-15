package pmax

import (
	"context"
	"encoding/json"
	types "github.com/dell/gopowermax/v2/types/v100"
	"net/http"
	"time"
)

const (
	Average      = "Average"
	Performance  = "performance"
	StorageGroup = "/StorageGroup"
	Volume       = "/Volume"
	Metrics      = "/metrics"
)

func (c *Client) GetStorageGroupMetrics(ctx context.Context, symID string, storageGroupID string, metricsQuery []string) (*types.StorageGroupMetricsIterator, error) {
	defer c.TimeSpent("GetStorageGroupMetrics", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := RESTPrefix + Performance + StorageGroup + Metrics
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	params := types.StorageGroupMetricsParam{
		SymmetrixId:    symID,
		StartDate:      time.Now().UnixMilli() - 300000,
		EndDate:        time.Now().UnixMilli(),
		DataFormat:     Average,
		StorageGroupId: storageGroupID,
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

func (c *Client) GetVolumesMetrics(ctx context.Context, symID string, storageGroups string, metricsQuery []string) (*types.VolumeMetricsIterator, error) {
	defer c.TimeSpent("GetStorageGroupMetrics", time.Now())
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	URL := RESTPrefix + Performance + Volume + Metrics
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	params := types.VolumeMetricsParam{
		SystemId:                       symID,
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
