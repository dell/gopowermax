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
	v100 "github.com/dell/gopowermax/v2/types/v100"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// The following constants are for internal use of the pmax library.
const (
	RESTPrefix          = "univmax/restapi/"
	StorageResourcePool = "srp"
)

var (
	// MAXJobRetryCount is the maximum number of retries to wait on a job.
	// It is a variable so that unit testing can set it lower.
	MAXJobRetryCount = 30
	// JobRetrySleepDuration is the amount of time between retries.
	JobRetrySleepDuration = 3 * time.Second
)

func (c *Client) urlPrefix() string {
	return RESTPrefix + c.version + "/"
}
func (c *Client) getSymmetrixIDListURL() string {
	return c.urlPrefix() + "system/symmetrix"
}

// Check respone to see if is nil or has bad HTTP status code.
func (c *Client) checkResponse(resp *http.Response) error {
	// parse the response
	switch {
	case resp == nil || resp.Body == nil:
		return errNilReponse
	case !(resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices):
		return c.api.ParseJSONError(resp)
	}
	return nil
}

// GetSymmetrixIDList returns a list of all the symmetrix systems known to the connected Unisphere instance.

func (c *Client) GetSymmetrixIDList(ctx context.Context) (*v100.SymmetrixIDList, error) {
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodGet, c.getSymmetrixIDListURL(), c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetSymmetrixIDList failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	symIDList := &v100.SymmetrixIDList{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(symIDList); err != nil {
		return nil, err
	}
	// we have the list of all arrays, filter out those not in the allowed arrays
	if len(c.GetAllowedArrays()) != 0 {
		allowed := make([]string, 0)
		for _, array := range symIDList.SymmetrixIDs {
			if ok, _ := c.IsAllowedArray(array); ok == true {
				allowed = append(allowed, array)
			}
		}
		symIDList.SymmetrixIDs = allowed
	}
	return symIDList, nil
}

// GetSymmetrixByID  returns the Symmetrix summary structure given a symmetrix id.
func (c *Client) GetSymmetrixByID(ctx context.Context, id string) (*v100.Symmetrix, error) {
	if _, err := c.IsAllowedArray(id); err != nil {
		return nil, err
	}
	url := c.getSymmetrixIDListURL() + "/" + id
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	resp, err := c.api.DoAndGetResponseBody(
		ctx, http.MethodGet, url, c.getDefaultHeaders(), nil)
	if err != nil {
		log.Error("GetSymmetrixIDList failed: " + err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	if err = c.checkResponse(resp); err != nil {
		return nil, err
	}

	symmetrix := &v100.Symmetrix{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(symmetrix); err != nil {
		return nil, err
	}
	return symmetrix, nil
}

// GetJobIDList returns a list of all the jobs in the symmetrix system.
// If optional statusQuery is something like JobStatusRunning it will search for running jobs.
func (c *Client) GetJobIDList(ctx context.Context, symID string, statusQuery string) ([]string, error) {
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	url := c.getSymmetrixIDListURL() + "/" + symID + "/" + "job"
	if statusQuery != "" {
		url = url + "?status=" + statusQuery
	}
	jobIDList := &v100.JobIDList{}
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	err := c.api.Get(ctx, url, c.getDefaultHeaders(), jobIDList)
	if err != nil {
		log.Error("GetJobIDList failed: " + err.Error())
		return nil, err
	}
	return jobIDList.JobIDs, nil
}

// GetJobByID returns a job given the job ID.
func (c *Client) GetJobByID(ctx context.Context, symID string, jobID string) (*v100.Job, error) {
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	maxRetry := 6
	for i := 0; i < maxRetry; i++ {
		url := c.getSymmetrixIDListURL() + "/" + symID + "/" + "job" + "/" + jobID
		job := &v100.Job{}
		err := c.api.Get(ctx, url, c.getDefaultHeaders(), job)
		if err != nil {
			if strings.Contains(err.Error(), "Cannot find role for user") {
				log.Debug(fmt.Sprintf("Retrying GetJobs: %s", err.Error()))
				time.Sleep(10 * time.Second)
				continue
			}
			log.Error("GetJobs failed: " + err.Error())
			return nil, err
		}
		return job, nil
	}
	return nil, fmt.Errorf("GetJob still failing after %d retries", maxRetry)
}

// WaitOnJobCompletion waits until a Job reaches a terminal state.
// The state may be JobStatusSucceeded or JobStatusFailed (it is the caller's responsibility to check.)
func (c *Client) WaitOnJobCompletion(ctx context.Context, symID string, jobID string) (*v100.Job, error) {
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	for i := 0; i < MAXJobRetryCount; i++ {
		job, err := c.GetJobByID(ctx, symID, jobID)
		if err != nil {
			return nil, err
		}
		log.Debug(c.JobToString(job))
		switch job.Status {
		case v100.JobStatusSucceeded:
			return job, nil
		case v100.JobStatusFailed:
			return job, nil
		}
		time.Sleep(JobRetrySleepDuration)
	}
	return nil, fmt.Errorf("Symmetrix %s Job %s timed out after %d retries", symID, jobID, MAXJobRetryCount)
}

// JobToString takes a Job and returns a string giving the job id, status, time completed, and result for easy display.
func (c *Client) JobToString(job *v100.Job) string {
	if job == nil {
		return "<nil Job>"
	}
	resourceString := ""
	resourceLinkElements := strings.Split(job.ResourceLink, "/")
	n := len(resourceLinkElements)
	if n > 5 {
		resourceString = fmt.Sprintf("%s/%s/%s", resourceLinkElements[n-3],
			resourceLinkElements[n-2], resourceLinkElements[n-1])
	}
	str := fmt.Sprintf("job id: %s status: %s completed: %s (%s) result: %s", job.JobID, job.Status, job.CompletedDate, resourceString, job.Result)
	return str
}

// GetDirectorIDList returns a list of all the directors on a given array.
func (c *Client) GetDirectorIDList(ctx context.Context, symID string) (*v100.DirectorIDList, error) {
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	directorList := &v100.DirectorIDList{}
	URL := c.getSymmetrixIDListURL() + "/" + symID + "/director"
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), directorList)
	if err != nil {
		log.Error("GetDirectorIDList failed: " + err.Error())
		return nil, err
	}

	return directorList, nil
}

// GetPortList returns a list of all the ports on a specified director/array.
func (c *Client) GetPortList(ctx context.Context, symID string, directorID string, query string) (*v100.PortList, error) {
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	portList := &v100.PortList{}
	URL := c.getSymmetrixIDListURL() + "/" + symID + "/director/" + directorID + "/port"
	if query != "" {
		URL = URL + "?" + query
	}
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), portList)
	if err != nil {
		log.Error("GetPortList failed: " + err.Error())
		return nil, err
	}

	return portList, nil
}

// GetPort returns port details.
func (c *Client) GetPort(ctx context.Context, symID string, directorID string, portID string) (*v100.Port, error) {
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	port := &v100.Port{}
	URL := c.getSymmetrixIDListURL() + "/" + symID + "/director/" + directorID + "/port/" + portID
	ctx, cancel := c.GetTimeoutContext(ctx)
	defer cancel()
	err := c.api.Get(ctx, URL, c.getDefaultHeaders(), port)
	if err != nil {
		log.Error("GetPort failed: " + err.Error())
		return nil, err
	}

	return port, nil
}

// GetListOfTargetAddresses returns list of target addresses
func (c *Client) GetListOfTargetAddresses(ctx context.Context, symID string) ([]string, error) {
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	ipAddr := []string{}
	// Get list of all directors
	directors, err := c.GetDirectorIDList(ctx, symID)
	if err != nil {
		return []string{}, err
	}

	// for each director, get list of ports with iscsi_target=true
	for _, d := range directors.DirectorIDs {

		ports, err := c.GetPortList(ctx, symID, d, "type=Gige")
		if err != nil {
			// Ignore the error and continue
			continue
		}

		// for each port, get the details
		for _, p := range ports.SymmetrixPortKey {
			port, err := c.GetPort(ctx, symID, d, p.PortID)
			if err != nil {
				// Ignore the error and continue
				continue
			}
			if len(port.SymmetrixPort.IPAddresses) > 0 {
				ipAddr = append(ipAddr, port.SymmetrixPort.IPAddresses...)
			}

		}
	}

	return ipAddr, nil
}

// GetISCSITargets returns list of target addresses
func (c *Client) GetISCSITargets(ctx context.Context, symID string) ([]ISCSITarget, error) {
	if _, err := c.IsAllowedArray(symID); err != nil {
		return nil, err
	}
	targets := make([]ISCSITarget, 0)
	// Get list of all directors
	directors, err := c.GetDirectorIDList(ctx, symID)
	if err != nil {
		return []ISCSITarget{}, err
	}

	for _, d := range directors.DirectorIDs {
		// Check if director is ISCSI
		// To do this, check if any ports have ports with GigE enabled
		ports, err := c.GetPortList(ctx, symID, d, "type=Gige")
		if err != nil {
			// Ignore the error and continue
			log.Errorf("Failed to get ports of type GigE for director: %s. Error: %s",
				d, err.Error())
			continue
		}
		if len(ports.SymmetrixPortKey) > 0 {
			// This is a director with ISCSI port(s)
			// Query for iscsi_targets
			virtualPorts, err := c.GetPortList(ctx, symID, d, "iscsi_target=true")
			if err != nil {
				return []ISCSITarget{}, err
			}
			// we have a list of virtual director ports which have ISCSI targets
			// and portal IPs associated with it
			for _, vp := range virtualPorts.SymmetrixPortKey {
				port, err := c.GetPort(ctx, symID, vp.DirectorID, vp.PortID)
				if err != nil {
					// Ignore the error and continue
					log.Errorf("Failed to fetch port details for %s:%s. Error: %s",
						vp.DirectorID, vp.PortID, err.Error())
					continue
				}
				// this should always be set
				if port.SymmetrixPort.Identifier != "" {
					tgt := ISCSITarget{
						IQN:       port.SymmetrixPort.Identifier,
						PortalIPs: port.SymmetrixPort.IPAddresses,
					}
					targets = append(targets, tgt)
				}
			}
		}
	}
	return targets, nil
}

// SetAllowedArrays sets the list of arrays which can be manipulated
// an empty list will allow all arrays to be accessed
func (c *Client) SetAllowedArrays(arrays []string) error {
	c.allowedArrays = arrays
	return nil
}

// GetAllowedArrays returns a slice of arrays that can be manipulated
func (c *Client) GetAllowedArrays() []string {
	return c.allowedArrays
}

// IsAllowedArray checks to see if we can manipulate the specified array
func (c *Client) IsAllowedArray(array string) (bool, error) {
	// if no list has been specified, allow all arrays
	if len(c.allowedArrays) == 0 {
		return true, nil
	}
	// check to see if the specified array in in the list
	for _, a := range c.allowedArrays {
		if a == array {
			return true, nil
		}
	}
	// we did not find the array
	return false, fmt.Errorf("the requested array (%s) is ignored as it is not managed", array)
}
