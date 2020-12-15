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

package inttest

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	pmax "github.com/dell/gopowermax"
	types "github.com/dell/gopowermax/types/v90"
)

const (
	SleepTime = 10 * time.Second
)

var (
	client pmax.Pmax
	// Following values will be read from env variables set in user.env
	endpoint = "https://0.0.0.0:8443"
	// username should match an existing user in Unisphere
	username = "username"
	// password should be the value for the corresponding user in Unisphere
	password                = "password"
	apiVersion              = "XX"
	symmetrixID             = "000000000001"
	defaultFCPortGroup      = "fc-pg"
	defaultiSCSIPortGroup   = "iscsi-pg"
	defaultFCInitiator      = "FA-1D:0:00000000abcd000e"
	defaultFCInitiatorID    string
	fcInitiator1            = "00000000abcd000f"
	fcInitiator2            = "00000000abcd000g"
	defaultiSCSIInitiator   = "SE-1E:000:iqn.1993-08.org.debian:01:012a34b5cd6"
	defaultiSCSIInitiatorID string
	iscsiInitiator1         = "iqn.1993-08.org.centos:01:012a34b5cd7"
	iscsiInitiator2         = "iqn.1993-08.org.centos:01:012a34b5cd8"
	defaultSRP              = "storage-pool"
	defaultServiceLevel     = "Diamond"
	volumePrefix            = "xx"
	sgPrefix                = "zz"
	snapshotPrefix          = "snap"
	// the test run will create these for the run and clean up in the end
	defaultStorageGroup = "csi-Integration-Test"
	nonFASTManagedSG    = "csi-Integration-No-FAST"
	defaultFCHost       = "IntegrationFCHost"
	defaultiSCSIHost    = "IntegrationiSCSIHost"
)

func setDefaultVariables() {
	endpoint = setenvVariable("Endpoint", endpoint)
	username = setenvVariable("Username", username)
	password = setenvVariable("Password", password)
	apiVersion = strings.TrimSpace(setenvVariable("APIVersion", ""))
	symmetrixID = setenvVariable("SymmetrixID", symmetrixID)
	defaultFCPortGroup = setenvVariable("DefaultFCPortGroup", defaultFCPortGroup)
	defaultiSCSIPortGroup = setenvVariable("DefaultiSCSIPortGroup", defaultiSCSIPortGroup)
	defaultFCInitiator = setenvVariable("DefaultFCInitiator", defaultFCInitiator)
	defaultFCInitiatorID = strings.Split(defaultFCInitiator, ":")[2]
	fcInitiator1 = setenvVariable("FCInitiator1", fcInitiator1)
	fcInitiator2 = setenvVariable("FCInitiator2", fcInitiator2)
	defaultiSCSIInitiator = setenvVariable("DefaultiSCSIInitiator", defaultiSCSIInitiator)
	defaultiSCSIInitiatorID = strings.Join(strings.Split(defaultiSCSIInitiator, ":")[2:], ":")
	iscsiInitiator1 = setenvVariable("ISCSIInitiator1", iscsiInitiator1)
	iscsiInitiator2 = setenvVariable("ISCSIInitiator2", iscsiInitiator2)
	volumePrefix = setenvVariable("VolumePrefix", volumePrefix)
	defaultSRP = setenvVariable("DefaultStoragePool", defaultSRP)
	defaultServiceLevel = setenvVariable("DefaultServiceLevel", defaultServiceLevel)
	sgPrefix = setenvVariable("SGPrefix", sgPrefix)
	snapshotPrefix = setenvVariable("SnapPrefix", snapshotPrefix)
}

func TestMain(m *testing.M) {
	status := 0
	// Process environment variables
	setDefaultVariables()

	err := createDefaultSGAndHost() // Creates default storage group and host for the test
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if st := m.Run(); st > status {
		status = st
	}
	fmt.Printf("status %d\n", status)
	doCleanUp := setenvVariable("Cleanup", "true")
	var cleanupTests = []testing.InternalTest{}
	if doCleanUp != "false" {
		fmt.Println("========= CLEANUP ==========")
		cleanupTests = append(cleanupTests, testing.InternalTest{
			Name: "cleanupDefaultSGAndHOST",
			F:    cleanDefaultUpSGAndHOST,
		})
	}
	afterRun(cleanupTests) // Cleans up the volumes and snapshots created for replication testing purposes.
}

func setenvVariable(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if key != "Username" && key != "Password" {
			fmt.Printf("%s=%s\n", key, defaultValue)
		}
		return defaultValue
	}
	if key != "Username" && key != "Password" {
		fmt.Printf("%s=%s\n", key, value)
	}
	return value
}

func createDefaultSGAndHost() error {
	fmt.Println("Creating default SG and host...")
	// Create default SG with SRP
	_, err := createStorageGroup(symmetrixID, defaultStorageGroup, defaultSRP, defaultServiceLevel, false)
	if err != nil {
		return fmt.Errorf("failed to create SG: (%s)", err.Error())
	}

	// Create default SG without srp
	_, err = createStorageGroup(symmetrixID, nonFASTManagedSG, "none", "none", false)
	if err != nil {
		return fmt.Errorf("failed to create non fast SG: (%s)", err.Error())
	}

	// Create default FC Host
	initiators := []string{defaultFCInitiatorID}
	_, err = createHost(symmetrixID, defaultFCHost, initiators, nil)
	if err != nil {
		return fmt.Errorf("failed to create default FC host: (%s)", err.Error())
	}
	// Create default iSCSI Host
	initiators = []string{defaultiSCSIInitiatorID}
	_, err = createHost(symmetrixID, defaultiSCSIHost, initiators, nil)
	if err != nil {
		return fmt.Errorf("failed to create default FC host: (%s)", err.Error())
	}
	return err
}

func cleanDefaultUpSGAndHOST(t *testing.T) {
	fmt.Println("Cleaning up SG and host...")
	// delete default SG
	err := deleteStorageGroup(symmetrixID, defaultStorageGroup)
	if err != nil {
		t.Errorf("failed to delete default SG (%s) : (%s)", defaultStorageGroup, err.Error())
	}
	// delete default non fast SG
	err = deleteStorageGroup(symmetrixID, nonFASTManagedSG)
	if err != nil {
		t.Errorf("failed to delete default non fast SG (%s) : (%s)", nonFASTManagedSG, err.Error())
	}
	// delete default FC host
	err = deleteHost(symmetrixID, defaultFCHost)
	if err != nil {
		t.Errorf("failed to delete default FC host (%s) : (%s)", defaultFCHost, err.Error())
	}
	// delete default iSCSI host
	err = deleteHost(symmetrixID, defaultiSCSIHost)
	if err != nil {
		t.Errorf("failed to delete default iSCSI host (%s) : (%s)", defaultFCHost, err.Error())
	}
}

func getClient() error {
	var err error
	client, err = pmax.NewClientWithArgs(endpoint, apiVersion, "CSI Driver for Dell EMC PowerMax v1.0",
		true, false)
	if err != nil {
		return err
	}
	err = client.Authenticate(&pmax.ConfigConnect{
		Endpoint: endpoint,
		Username: username,
		Password: password})
	if err != nil {
		return err
	}
	return nil
}

func TestAuthentication(t *testing.T) {
	_ = getClient()
}

func TestGetSymmetrixIDs(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	symIDList, err := client.GetSymmetrixIDList()
	if err != nil || symIDList == nil {
		t.Error("cannot get SymmetrixIDList: ", err.Error())
		return
	}
	if len(symIDList.SymmetrixIDs) == 0 {
		t.Error("expected at least one Symmetrix ID in list")
		return
	}
	for _, id := range symIDList.SymmetrixIDs {
		fmt.Printf("symmetrix ID: %s\n", id)
	}
}

func TestGetSymmetrix(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	symmetrix, err := client.GetSymmetrixByID(symmetrixID)
	if err != nil || symmetrix == nil {
		t.Error("cannot get Symmetrix id "+symmetrixID, err.Error())
		return
	}
	fmt.Printf("Symmetrix %s: %#v\n", symmetrixID, symmetrix)
}

func TestGetVolumeIDs(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	volumeIDList, err := client.GetVolumeIDList(symmetrixID, "", false)
	if err != nil || volumeIDList == nil {
		t.Error("cannot get volumeIDList: ", err.Error())
		return
	}
	fmt.Printf("%d volume IDs\n", len(volumeIDList))
	// Make sure no duplicates
	dupMap := make(map[string]bool)
	for i := 0; i < len(volumeIDList); i++ {
		if volumeIDList[i] == "" {
			t.Error("Got an empty volume ID")
		}
		id := volumeIDList[i]
		if dupMap[id] == true {
			t.Error("Got duplicate ID:" + id)
		}
		dupMap[id] = true
	}

	volumeIDList, err = client.GetVolumeIDList(symmetrixID, "csi", true)
	if err != nil || volumeIDList == nil {
		t.Error("cannot get volumeIDList: ", err.Error())
		return
	}
	fmt.Printf("%d CSI volume IDs\n", len(volumeIDList))
	for _, id := range volumeIDList {
		fmt.Printf("CSI volume: %s\n", id)
	}

	volumeIDList, err = client.GetVolumeIDList(symmetrixID, "ce9072c0", true)
	if err != nil || volumeIDList == nil {
		t.Error("cannot get volumeIDList: ", err.Error())
		return
	}
	fmt.Printf("%d CSI volume IDs\n", len(volumeIDList))
}

func TestGetVolume(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	// Get some CSI volumes
	volumeIDList, err := client.GetVolumeIDList(symmetrixID, "csi", true)
	if err != nil || volumeIDList == nil {
		t.Error("cannot get CSI volumeIDList: ", err.Error())
		return
	}
	if len(volumeIDList) == 0 {
		t.Error("no CSI volumes")
		return
	}
	for i, id := range volumeIDList {
		if i >= 3 {
			break
		}
		volume, err := client.GetVolumeByID(symmetrixID, id)
		if err != nil {
			t.Error("cannot retrieve Volume: " + err.Error())
		} else {
			fmt.Printf("Volume %#v\n", volume)
		}

	}
}

func TestGetNonExistentVolume(t *testing.T) {
	volume, err := client.GetVolumeByID(symmetrixID, "88888")
	if err != nil {
		fmt.Printf("TestGetNonExistentVolume: %s\n", err.Error())
	} else {
		fmt.Printf("%#v\n", volume)
		t.Error("Expected volume 88888 to be non-existent")
	}
}

func TestGetStorageGroupIDs(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	sgIDList, err := client.GetStorageGroupIDList(symmetrixID)
	if err != nil || sgIDList == nil {
		t.Error("cannot get StorageGroupIDList: ", err.Error())
		return
	}
	if len(sgIDList.StorageGroupIDs) == 0 {
		t.Error("expected at least one StorageGroup ID in list")
		return
	}
	for _, id := range sgIDList.StorageGroupIDs {
		fmt.Printf("StorageGroup ID: %s\n", id)
	}
}

func TestGetStorageGroup(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	storageGroup, err := client.GetStorageGroup(symmetrixID, defaultStorageGroup)
	if err != nil || storageGroup == nil {
		t.Error("Expected to find " + defaultStorageGroup + " but didn't")
		return
	}
	fmt.Printf("%#v\n", storageGroup)
}

func TestGetStoragePool(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	storagePool, err := client.GetStoragePool(symmetrixID, defaultSRP)
	if err != nil || storagePool == nil {
		t.Error("Expected to find " + defaultSRP + " but didn't")
		return
	}
	fmt.Printf("%#v\n", storagePool)
}

func createStorageGroup(symmetrixID, storageGroupID, srp, serviceLevel string, isThick bool) (*types.StorageGroup, error) {
	if client == nil {
		err := getClient()
		if err != nil {
			return nil, err
		}
	}
	// Check if the SG exists on array
	storageGroup, err := client.GetStorageGroup(symmetrixID, storageGroupID)
	// Storage Group already exist, returning old one
	if storageGroup != nil && err == nil {
		return storageGroup, err
	}
	// Create a new storege group
	fmt.Println("Creating a new storage group...")
	return client.CreateStorageGroup(symmetrixID, storageGroupID, srp, serviceLevel, isThick)
}

func deleteStorageGroup(symmetrixID, storageGroupID string) error {
	return client.DeleteStorageGroup(symmetrixID, storageGroupID)
}

func TestCreateStorageGroup(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	now := time.Now()
	storageGroupID := fmt.Sprintf("csi-%s-Int%d-SG", sgPrefix, now.Nanosecond())
	storageGroup, err := createStorageGroup(symmetrixID, storageGroupID,
		defaultSRP, defaultServiceLevel, false)
	if err != nil || storageGroup == nil {
		t.Error("Failed to create " + storageGroupID + " " + err.Error())
		return
	}
	fmt.Println("Fetching the newly create storage group from array")
	//Check if the SG exists on array
	storageGroup, err = client.GetStorageGroup(symmetrixID, storageGroupID)
	if err != nil || storageGroup == nil {
		t.Error("Expected to find " + storageGroupID + " but didn't")
		return
	}
	fmt.Printf("%#v\n", storageGroup)
	fmt.Println("Cleaning up the storage group: " + storageGroupID)
	err = deleteStorageGroup(symmetrixID, storageGroupID)
	if err != nil {
		t.Error("Failed to delete " + storageGroupID)
		return
	}
	//Check if the SG exists on array
	storageGroup, err = client.GetStorageGroup(symmetrixID, storageGroupID)
	if err == nil || storageGroup != nil {
		t.Error("Expected a failure in fetching " + storageGroupID + " but didn't")
		return
	}
	fmt.Println(fmt.Sprintf("Error received while fetching %s: %s", storageGroupID, err.Error()))
}

func TestCreateStorageGroupNonFASTManaged(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	now := time.Now()
	storageGroupID := fmt.Sprintf("csi-%s-Int%d-SG-No-FAST", sgPrefix, now.Nanosecond())
	storageGroup, err := createStorageGroup(symmetrixID, storageGroupID,
		"None", "None", false)
	if err != nil || storageGroup == nil {
		t.Error("Failed to create " + storageGroupID)
		return
	}
	fmt.Printf("%#v\n", storageGroup)
	if storageGroup.SRP != "" {
		t.Error("Expected no SRP but received: " + storageGroup.SRP)
	}
	fmt.Println("Cleaning up the storage group: " + storageGroupID)
	err = client.DeleteStorageGroup(symmetrixID, storageGroupID)
	if err != nil {
		t.Error("Failed to delete " + storageGroupID)
		return
	}
	//Check if the SG exists on array
	storageGroup, err = client.GetStorageGroup(symmetrixID, storageGroupID)
	if err == nil || storageGroup != nil {
		t.Error("Expected a failure in fetching " + storageGroupID + " but didn't")
		return
	}
	fmt.Println(fmt.Sprintf("Error received while fetching %s: %s", storageGroupID, err.Error()))
}

func TestGetJobs(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	jobIDList, err := client.GetJobIDList(symmetrixID, "")
	if err != nil {
		t.Error("failed to get Job ID LIst")
		return
	}
	for i, id := range jobIDList {
		if i >= 10 {
			break
		}
		job, err := client.GetJobByID(symmetrixID, id)
		if err != nil {
			t.Error("failed to get job: " + id)
			return
		}
		fmt.Printf("%s\n", client.JobToString(job))
	}

	jobIDList, err = client.GetJobIDList(symmetrixID, types.JobStatusRunning)
	if err != nil {
		t.Error("failed to get Job ID LIst")
		return
	}
	for i, id := range jobIDList {
		if i >= 10 {
			break
		}
		job, err := client.GetJobByID(symmetrixID, id)
		if err != nil {
			t.Error("failed to get job: " + id)
			return
		}
		fmt.Printf("%s\n", client.JobToString(job))
		if job.Status != types.JobStatusRunning {
			t.Error("Expected Running job: " + client.JobToString(job))
		}
	}
}

func TestGetStoragePoolList(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	spList, err := client.GetStoragePoolList(symmetrixID)
	if err != nil {
		t.Error("Failed to get StoragePoolList: " + err.Error())
		return
	}
	for _, value := range spList.StoragePoolIDs {
		fmt.Printf("Storage Resource Pool: %s\n", value)
	}
}

func TestGetMaskingViews(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	mvList, err := client.GetMaskingViewList(symmetrixID)
	if err != nil {
		t.Error("Failed to get MaskingViewList: " + err.Error())
		return
	}
	for _, mvID := range mvList.MaskingViewIDs {
		fmt.Printf("Masking View: %s\n", mvID)
		mv, err := client.GetMaskingViewByID(symmetrixID, mvID)
		if err != nil {
			t.Error("Failed to GetMaskingViewByID: ", err.Error())
			return
		}
		fmt.Printf("%#v\n", mv)
		conns, err := client.GetMaskingViewConnections(symmetrixID, mvID, "")
		if err != nil {
			t.Error("Failed to GetMaskingViewConnections: ", err.Error())
			return
		}
		for _, conn := range conns {
			fmt.Printf("mv connection VolumeID %s HostLUNAddress %s InitiatorID %s DirectorPort %s\n",
				conn.VolumeID, conn.HostLUNAddress, conn.InitiatorID, conn.DirectorPort)
		}
	}
}

func TestCreateVolumeInStorageGroup1(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	now := time.Now()
	volumeName := fmt.Sprintf("csi%s-Int%d", volumePrefix, now.Nanosecond())
	fmt.Printf("volumeName: %s\n", volumeName)
	payload := client.GetCreateVolInSGPayload(1, volumeName, false)

	payloadBytes, err := json.Marshal(&payload)
	if err != nil {
		t.Error("Encoding error on json")
	}
	fmt.Printf("payload: %s\n", string(payloadBytes))

	job, err := client.UpdateStorageGroup(symmetrixID, defaultStorageGroup, payload)
	if err != nil {
		t.Error("Error returned from UpdateStorageGroup")
		return
	}
	jobID := job.JobID
	job, err = client.WaitOnJobCompletion(symmetrixID, jobID)
	if err == nil {
		idlist, err := client.GetVolumeIDList(symmetrixID, volumeName, false)
		if err != nil {
			t.Error("Error getting volume IDs: " + err.Error())
		}
		for _, id := range idlist {
			cleanupVolume(id, volumeName, defaultStorageGroup, t)
		}
	}
}

func TestCreateVolumeInStorageGroup2(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	now := time.Now()
	volumeName := fmt.Sprintf("csi%s-Int%d", volumePrefix, now.Nanosecond())
	fmt.Printf("volumeName: %s\n", volumeName)
	vol, err := client.CreateVolumeInStorageGroup(symmetrixID, defaultStorageGroup, volumeName, 1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("volume:\n%#v\n", vol)
	cleanupVolume(vol.VolumeID, volumeName, defaultStorageGroup, t)
}

func TestAddVolumesInStorageGroup(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	now := time.Now()
	volumeName := fmt.Sprintf("csi%s-Int%d", volumePrefix, now.Nanosecond())
	fmt.Printf("volumeName: %s\n", volumeName)
	vol, err := client.CreateVolumeInStorageGroup(symmetrixID, defaultStorageGroup, volumeName, 1)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("volume:\n%#v\n", vol)
	err = client.AddVolumesToStorageGroup(symmetrixID, nonFASTManagedSG, vol.VolumeID)
	if err != nil {
		t.Error(err)
		return
	}
	sg, err := client.GetStorageGroup(symmetrixID, nonFASTManagedSG)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("SG after adding volume: %#v\n", sg)
	//Remove the volume from SG as part of cleanup
	sg, err = client.RemoveVolumesFromStorageGroup(symmetrixID, nonFASTManagedSG, vol.VolumeID)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("SG after removing volume: %#v\n", sg)
	cleanupVolume(vol.VolumeID, volumeName, defaultStorageGroup, t)
}

func cleanupVolume(volumeID string, volumeName string, storageGroup string, t *testing.T) {
	if volumeName != "" {
		vol, err := client.RenameVolume(symmetrixID, volumeID, "_DEL"+volumeName)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("volume Renamed: %s\n", vol.VolumeIdentifier)
	}
	sg, err := client.RemoveVolumesFromStorageGroup(symmetrixID, storageGroup, volumeID)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("SG after removing volume: %#v\n", sg)
	pmax.Debug = true
	fmt.Printf("Initiating removal of tracks\n")
	job, err := client.InitiateDeallocationOfTracksFromVolume(symmetrixID, volumeID)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("Waiting on job: %s\n", client.JobToString(job))
	job, err = client.WaitOnJobCompletion(symmetrixID, job.JobID)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("Job completion status: %s\n", client.JobToString(job))
	switch job.Status {
	case "SUCCEEDED":
	case "FAILED":
		if strings.Contains(job.Result, "The device is already in the requested state") {
			break
		}
		t.Error("Track deallocation job failed: " + job.Result)
	}
	err = client.DeleteVolume(symmetrixID, volumeID)
	if err != nil {
		t.Error("DeleteVolume failed: " + err.Error())
	}
	// Test deletion of the volume again... should return an error
	err = client.DeleteVolume(symmetrixID, volumeID)
	if err == nil {
		t.Error("Expected an error saying volume was not found, but no error")
	}
	fmt.Printf("Received expected error: %s\n", err.Error())
}

func TestCreateVolumeInStorageGroupInParallel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping this test in short mode")
	}
	// make sure we have a client
	getClient()
	CreateVolumesInParallel(5, t)
}

func CreateVolumesInParallel(nVols int, t *testing.T) {
	fmt.Printf("testing CreateVolumeInStorageGroup with %d parallel requests\n", nVols)
	volIDList := make([]string, nVols)
	// make channels for communication
	idchan := make(chan string, nVols)
	errchan := make(chan error, nVols)
	t0 := time.Now()

	// create a temporary storage group
	now := time.Now()
	storageGroupName := fmt.Sprintf("pmax-%s-Int%d-SG", sgPrefix, now.Nanosecond())
	_, err := createStorageGroup(symmetrixID, storageGroupName,
		defaultSRP, defaultServiceLevel, false)
	if err != nil {
		t.Errorf("Unable to create temporary Storage Group: %s", storageGroupName)
	}
	// Send requests
	for i := 0; i < nVols; i++ {
		name := fmt.Sprintf("pmax-Int%d-Scale%d", now.Nanosecond(), i)
		go func(volumeName string, idchan chan string, errchan chan error) {
			var err error
			resp, err := client.CreateVolumeInStorageGroup(symmetrixID, storageGroupName, volumeName, 1)
			if resp != nil {
				fmt.Printf("ID %s Name %s\n%#v\n", resp.VolumeID, volumeName, resp)
				idchan <- resp.VolumeID
			} else {
				idchan <- ""
			}
			errchan <- err
		}(name, idchan, errchan)
	}
	// Wait on complete, collecting ids and errors
	nerrors := 0
	for i := 0; i < nVols; i++ {
		var id string
		var err error
		id = <-idchan
		if id != "" {
			volIDList[i] = id
		}
		err = <-errchan
		if err != nil {
			err = fmt.Errorf("create volume received error: %s", err.Error())
			t.Error(err.Error())
			nerrors++
		}
	}
	t1 := time.Now()
	fmt.Printf("Create volume time for %d volumes %d errors: %v %v\n", nVols, nerrors, t1.Sub(t0).Seconds(), t1.Sub(t0).Seconds()/float64(nVols))
	fmt.Printf("%v\n", volIDList)
	time.Sleep(SleepTime)
	// Cleanup the volumes
	for _, id := range volIDList {
		cleanupVolume(id, "", storageGroupName, t)
	}
	// remove the temporary storage group
	err = client.DeleteStorageGroup(symmetrixID, storageGroupName)
	if err != nil {
		t.Errorf("Unable to delete temporary Storage Group: %s", storageGroupName)
	}
}

func TestGetPorts(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	dirName := "FA-1D"
	portName := "4"
	port, err := client.GetPort(symmetrixID, dirName, portName)
	if err != nil {
		t.Errorf("Unable to read FC storage port %s %s: %s", dirName, portName, err)
		return
	}
	fmt.Printf("port %s:%s %#v\n", dirName, portName, port)
	dirName = "SE-1E"
	portName = "0"
	port, err = client.GetPort(symmetrixID, dirName, portName)
	if err != nil {
		t.Errorf("Unable to read iSCSI storage port %s %s: %s", dirName, portName, err)
		return
	}
	fmt.Printf("port %s:%s %#v\n", dirName, portName, port)

}

func TestGetPortGroupIDs(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	pgList, err := client.GetPortGroupList(symmetrixID, "")
	if err != nil || pgList == nil {
		t.Error("cannot get PortGroupList: ", err.Error())
		return
	}
	if len(pgList.PortGroupIDs) == 0 {
		t.Error("expected at least one PortGroup ID in list")
		return
	}
	pgList, err = client.GetPortGroupList(symmetrixID, "fibre")
	if err != nil || pgList == nil {
		t.Error("cannot get FC PortGroupList: ", err.Error())
		return
	}
	if len(pgList.PortGroupIDs) == 0 {
		t.Error("expected at least one FC PortGroup ID in list")
		return
	}
	pgList, err = client.GetPortGroupList(symmetrixID, "iscsi")
	if err != nil || pgList == nil {
		t.Error("cannot get iSCSI PortGroupList: ", err.Error())
		return
	}
	if len(pgList.PortGroupIDs) == 0 {
		t.Error("expected at least one iSCSI PortGroup ID in list")
		return
	}

}
func TestGetPortGroupByFCID(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	portGroup, err := client.GetPortGroupByID(symmetrixID, defaultFCPortGroup)
	if err != nil || portGroup == nil {
		t.Error("Expected to find " + defaultFCPortGroup + " but didn't")
		return
	}
}
func TestGetPortGroupByiSCSIID(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	portGroup, err := client.GetPortGroupByID(symmetrixID, defaultiSCSIPortGroup)
	if err != nil || portGroup == nil {
		t.Error("Expected to find " + defaultiSCSIPortGroup + " but didn't")
		return
	}
	fmt.Printf("PortGroup: %#v\n", portGroup)
}

func TestGetInitiatorIDs(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	initList, err := client.GetInitiatorList(symmetrixID, "", true, false)
	if err != nil || initList == nil {
		t.Error("cannot get Initiator List: ", err.Error())
		return
	}
	if len(initList.InitiatorIDs) == 0 {
		t.Error("expected at least one Initiator ID in list")
		return
	}
	// Get the FC initiator list for the default initiator HBA
	initList, err = client.GetInitiatorList(symmetrixID, defaultFCInitiatorID, false, true)
	if err != nil {
		t.Error("Receieved error : ", err.Error())
	}
	if len(initList.InitiatorIDs) != 0 {
		fmt.Println(initList.InitiatorIDs)
	} else {
		fmt.Println("Received an empty FC list")
		t.Error("Expected to find atleast one FC initiator")
	}
	// Get the iSCSI initiator list for the default IQN
	initList, err = client.GetInitiatorList(symmetrixID, defaultiSCSIInitiatorID, true, true)
	if err != nil {
		t.Error("Receieved error : ", err.Error())
	}
	if len(initList.InitiatorIDs) != 0 {
		fmt.Println(initList.InitiatorIDs)
	} else {
		fmt.Println("Received an empty iSCSI list")
		t.Error("Expected to find atleast one iSCSI initiator")
	}

	// Get the initiator list for an IQN not on the array
	initList, err = client.GetInitiatorList(symmetrixID, "iqn.1993-08.org.desian:01:5ae293b352a2", true, true)
	if err != nil {
		t.Error("Received error : ", err.Error())
	}
	if len(initList.InitiatorIDs) != 0 {
		fmt.Println(initList.InitiatorIDs)
	} else {
		fmt.Println("Received an empty list as expected for unknown IQN")
	}
}
func TestGetInitiatorByFCID(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	initiator, err := client.GetInitiatorByID(symmetrixID, defaultFCInitiator)
	if err != nil || initiator == nil {
		t.Error("Expected to find " + defaultFCInitiator + " but didn't")
		return
	}
	fmt.Printf("defaultFCInitator: %#v\n", initiator)
}

func TestGetInitiatorByiSCSIID(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	initiator, err := client.GetInitiatorByID(symmetrixID, defaultiSCSIInitiator)
	if err != nil || initiator == nil {
		t.Error("Expected to find " + defaultiSCSIInitiator + " but didn't")
		return
	}
	fmt.Printf("defaultiSCSIInitator: %#v\n", initiator)
}

func TestFCGetInitiators(t *testing.T) {
	// Get all the initiators and print the FC ones
	initList, err := client.GetInitiatorList(symmetrixID, "", false, false)
	if err != nil || initList == nil {
		t.Error("cannot get Initiator List: ", err.Error())
		return
	}

	// Read our FC initiators from the /sys/class/fc_host directory
	cmd := exec.Command("/bin/sh", "-c", "cd /sys/class/fc_host; cat */port_name")
	bytes, err := cmd.Output()
	if err != nil {
		return
	}

	// Look for our initiators on the array
	ourInits := strings.Split(string(bytes), "\n")
	if len(ourInits) == 0 {
		// We have any initiators that we know
		return
	}
	fcInitiators := make([]string, 0)
	for _, ourInit := range ourInits {
		ourInit := strings.TrimSpace(strings.Replace(ourInit, "0x", "", 1))
		if ourInit == "" {
			continue
		}
		//fmt.Printf("ourInit: %s\n", ourInit)
		for _, init := range initList.InitiatorIDs {
			if strings.HasSuffix(init, ourInit) {
				fmt.Printf("initiator: %s\n", init)
				fcInitiators = append(fcInitiators, init)
			}
		}
	}

	// Print the matching initiator structures from the Symmetrix
	for _, fcInit := range fcInitiators {
		initiator, err := client.GetInitiatorByID(symmetrixID, fcInit)
		if err != nil || initiator == nil {
			t.Errorf("Unable to read FC initiator: %s", err)
			return
		}
		fmt.Printf("%#v\n", initiator)
	}
}

func TestGetHostIDs(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	hostList, err := client.GetHostList(symmetrixID)
	for _, id := range hostList.HostIDs {
		fmt.Printf("Host ID: %s\n", id)
	}
	if err != nil || hostList == nil {
		t.Error("cannot get Host List: ", err.Error())
		return
	}
	if len(hostList.HostIDs) == 0 {
		t.Error("expected at least one Host ID in list")
		return
	}
}
func TestGetHostByFCID(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	host, err := client.GetHostByID(symmetrixID, defaultFCHost)
	if err != nil || host == nil {
		t.Error("Expected to find FC Host" + defaultFCHost + " but didn't")
		return
	}
	fmt.Printf("defaultHost: %#v\n", host)
}

func TestGetHostByiSCSIID(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	host, err := client.GetHostByID(symmetrixID, defaultiSCSIHost)
	if err != nil || host == nil {
		t.Error("Expected to find " + defaultiSCSIHost + " but didn't")
		return
	}
	fmt.Printf("defaultHost: %#v\n", host)
}

func createHost(symmetrixID, hostID string, initiatorKeys []string, hostFlag *types.HostFlags) (*types.Host, error) {
	if client == nil {
		err := getClient()
		if err != nil {
			return nil, err
		}
	}
	// Check and return if the host exist on the array
	host, err := client.GetHostByID(symmetrixID, hostID)
	if host != nil && err == nil {
		return host, nil
	}
	// host not found, create the host
	fmt.Println("Creating a new host...")
	return client.CreateHost(symmetrixID, hostID, initiatorKeys, hostFlag)
}

func deleteHost(symmetrixID, hostID string) error {
	return client.DeleteHost(symmetrixID, hostID)
}

func TestCreateFCHost(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			return
		}
	}
	initiatorKeys := make([]string, 0)
	initiatorKeys = append(initiatorKeys, fcInitiator1)
	host, err := createHost(symmetrixID, "IntTestFCHost", initiatorKeys, nil)
	if err != nil || host == nil {
		t.Error("Expected to create FC host but didn't: " + err.Error())
		return
	}
	fmt.Printf("%#v\n, FC host", host)
	err = deleteHost(symmetrixID, "IntTestFCHost")
	if err != nil {
		t.Error("Could not delete FC Host: " + err.Error())
	}
}

func TestCreateiSCSIHost(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			return
		}
	}
	initiatorKeys := make([]string, 0)
	initiatorKeys = append(initiatorKeys, iscsiInitiator1)
	host, err := createHost(symmetrixID, "IntTestiSCSIHost", initiatorKeys, nil)
	if err != nil || host == nil {
		t.Error("Expected to create host but didn't: " + err.Error())
		return
	}
	fmt.Printf("%#v\n, host", host)
	err = deleteHost(symmetrixID, "IntTestiSCSIHost")
	if err != nil {
		t.Error("Could not delete Host: " + err.Error())
	}
}

func TestCreateFCMaskingView(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping this test in short mode")
	}
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	hostID := "IntTestFCMV-Host"
	// In case a prior test left the Host in the system...
	client.DeleteHost(symmetrixID, hostID)

	// create a Host with some initiators
	initiatorKeys := make([]string, 0)
	initiatorKeys = append(initiatorKeys, fcInitiator2)
	fmt.Println("Setting up a host before creation of masking view")
	host, err := createHost(symmetrixID, hostID, initiatorKeys, nil)
	if err != nil || host == nil {
		t.Error("Expected to create host but didn't: " + err.Error())
		return
	}
	fmt.Printf("%#v\n, host", host)
	// add a volume in defaultStorageGroup
	volumeName := fmt.Sprintf("csi%s-Int%d", volumePrefix, time.Now().Nanosecond())
	fmt.Printf("volumeName: %s\n", volumeName)
	vol, err := client.CreateVolumeInStorageGroup(symmetrixID, defaultStorageGroup, volumeName, 1)
	if err != nil {
		t.Error("Expected to create a volume but didn't" + err.Error())
		return
	}
	fmt.Printf("volume:\n%#v\n", vol)
	maskingViewID := "IntTestFCMV"
	maskingView, err := client.CreateMaskingView(symmetrixID, maskingViewID, defaultStorageGroup,
		hostID, true, defaultFCPortGroup)
	if err != nil {
		t.Error("Expected to create MV with FC initiator and port but didn't: " + err.Error())
		cleanupHost(symmetrixID, hostID, t)
		return
	}
	fmt.Println("Fetching the newly created masking view from array")
	//Check if the MV exists on array
	maskingView, err = client.GetMaskingViewByID(symmetrixID, maskingViewID)
	if err != nil || maskingView == nil {
		t.Error("Expected to find " + maskingViewID + " but didn't")
		cleanupHost(symmetrixID, hostID, t)
		return
	}
	fmt.Printf("%#v\n", maskingView)
	fmt.Println("Cleaning up the masking view")
	err = client.DeleteMaskingView(symmetrixID, maskingViewID)
	if err != nil {
		t.Error("Failed to delete " + maskingViewID)
		return
	}
	fmt.Println("Sleeping for 20 seconds")
	time.Sleep(20 * time.Second)
	// Confirm if the masking view got deleted
	maskingView, err = client.GetMaskingViewByID(symmetrixID, maskingViewID)
	if err == nil {
		t.Error("Expected a failure in fetching MV: " + maskingViewID + "but didn't")
		fmt.Printf("%#v\n", maskingView)
		return
	}
	fmt.Println(fmt.Sprintf("Error in fetching %s: %s", maskingViewID, err.Error()))
	cleanupVolume(vol.VolumeID, volumeName, defaultStorageGroup, t)
	cleanupHost(symmetrixID, hostID, t)
}

func TestCreatePortGroup(t *testing.T) {
	t.Skip("Skipping this test until Delete Port Group is implemented")
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	portGroupID := "IntTestPG"
	portKeys := make([]types.PortKey, 0)
	portKey := types.PortKey{
		DirectorID: "FA-1D",
		PortID:     "4",
	}
	portKeys = append(portKeys, portKey)
	portGroup, err := client.CreatePortGroup(symmetrixID, portGroupID, portKeys)
	if err != nil {
		t.Error("Couldn't create port group")
		return
	}
	fmt.Println(portGroup.PortGroupID)
}

func TestCreateiSCSIMaskingView(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping this test in short mode")
	}
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf("Unable to get/create pmax client: (%s)", err.Error())
			return
		}
	}
	hostID := "IntTestiSCSIMV-Host"
	// In case a prior test left the Host in the system...
	client.DeleteHost(symmetrixID, hostID)

	// create a Host with some initiators
	initiatorKeys := make([]string, 0)
	initiatorKeys = append(initiatorKeys, iscsiInitiator2)
	fmt.Println("Setting up a host before creation of masking view")
	host, err := createHost(symmetrixID, hostID, initiatorKeys, nil)
	if err != nil || host == nil {
		t.Error("Expected to create iscsi host but didn't: " + err.Error())
		return
	}
	fmt.Printf("%#v\n, host", host)
	// add a volume in defaultStorageGroup
	volumeName := fmt.Sprintf("csi%s-Int%d", volumePrefix, time.Now().Nanosecond())
	fmt.Printf("volumeName: %s\n", volumeName)
	vol, err := client.CreateVolumeInStorageGroup(symmetrixID, defaultStorageGroup, volumeName, 1)
	if err != nil {
		t.Error("Expected to create a volume but didn't" + err.Error())
		return
	}
	fmt.Printf("volume:\n%#v\n", vol)
	// create the masking view
	maskingViewID := "IntTestiSCSIMV"
	maskingView, err := client.CreateMaskingView(symmetrixID, maskingViewID, defaultStorageGroup,
		hostID, true, defaultiSCSIPortGroup)
	if err != nil {
		t.Error("Expected to create MV with iscsi initiator host and port but didn't: " + err.Error())
		cleanupHost(symmetrixID, hostID, t)
		return
	}
	fmt.Println("Fetching the newly created masking view from array")
	//Check if the MV exists on array
	maskingView, err = client.GetMaskingViewByID(symmetrixID, maskingViewID)
	if err != nil || maskingView == nil {
		t.Error("Expected to find " + maskingViewID + " but didn't")
		cleanupHost(symmetrixID, hostID, t)
		return
	}
	fmt.Printf("%#v\n", maskingView)
	fmt.Println("Cleaning up the masking view")
	err = client.DeleteMaskingView(symmetrixID, maskingViewID)
	if err != nil {
		t.Error("Failed to delete " + maskingViewID)
		return
	}
	fmt.Println("Sleeping for 20 seconds")
	time.Sleep(20 * time.Second)
	// Confirm if the masking view got deleted
	maskingView, err = client.GetMaskingViewByID(symmetrixID, maskingViewID)
	if err == nil {
		t.Error("Expected a failure in fetching MV: " + maskingViewID + "but didn't")
		fmt.Printf("%#v\n", maskingView)
		return
	}
	fmt.Println(fmt.Sprintf("Error in fetching %s: %s", maskingViewID, err.Error()))
	cleanupVolume(vol.VolumeID, volumeName, defaultStorageGroup, t)
	cleanupHost(symmetrixID, hostID, t)
}

func cleanupHost(symmetrixID string, hostID string, t *testing.T) {
	fmt.Println("Cleaning up the host")
	err := client.DeleteHost(symmetrixID, hostID)
	if err != nil {
		t.Error("Failed to delete " + hostID)
	}
	return
}

func TestUpdateHostInitiators(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err.Error())
			return
		}
	}

	// In case a prior test left the Host in the system...
	client.DeleteHost(symmetrixID, "IntTestHost")

	// create a Host with some initiators
	initiatorKeys := make([]string, 0)
	initiatorKeys = append(initiatorKeys, iscsiInitiator1)
	host, err := client.CreateHost(symmetrixID, "IntTestHost", initiatorKeys, nil)
	if err != nil || host == nil {
		t.Error("Expected to create host but didn't: " + err.Error())
		return
	}
	fmt.Printf("%#v\n, host", host)

	// change the list of initiators and update the host
	updatedInitiators := make([]string, 0)
	updatedInitiators = append(updatedInitiators, iscsiInitiator1, iscsiInitiator2)
	host, err = client.UpdateHostInitiators(symmetrixID, host, updatedInitiators)
	if err != nil || host == nil {
		t.Error("Expected to update host but didn't: " + err.Error())
		return
	}
	fmt.Printf("%#v\n, host", host)

	// validate that we have the right number of intiators
	if len(host.Initiators) != len(updatedInitiators) {
		msg := fmt.Sprintf("Expected %d initiators but received %d", len(updatedInitiators), len(host.Initiators))
		t.Error(msg)
		return
	}

	// delete the host
	err = client.DeleteHost(symmetrixID, "IntTestHost")
	if err != nil {
		t.Error("Could not delete Host: " + err.Error())
	}
}

func TestGetTargetAddresses(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
	addresses, err := client.GetListOfTargetAddresses(symmetrixID)
	if err != nil {
		t.Error("Error calling GetListOfTargetAddresses " + err.Error())
		return
	}
	fmt.Printf("Addresses: %v\n", addresses)
}

func TestGetISCSITargets(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
	targets, err := client.GetISCSITargets(symmetrixID)
	if err != nil {
		t.Error("Error calling GetISCSITargets " + err.Error())
		return
	}
	fmt.Printf("Targets: %v\n", targets)
}

func TestExpandVolume(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
	//create a volume
	now := time.Now()
	volumeName := fmt.Sprintf("csi%s-Int%d", volumePrefix, now.Nanosecond())
	fmt.Printf("volumeName: %s\n", volumeName)
	vol, err := client.CreateVolumeInStorageGroup(symmetrixID, defaultStorageGroup, volumeName, 26)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("volume:\n%#v\n", vol)
	//expand Volume
	expandedSize := 30
	fmt.Println("Doing VolumeExpansion")
	expandedVol, err := client.ExpandVolume(symmetrixID, vol.VolumeID, expandedSize)
	if err != nil {
		t.Error("Error in Volume Expansion: " + err.Error())
		return
	}
	//check expand size
	if expandedVol.CapacityCYL != expandedSize {
		t.Error("Size mismatch after Expansion: " + err.Error())
		return
	}
	fmt.Printf("volume:\n%#v\n", expandedVol)
	fmt.Printf("Expanded Volume Size:\n%d\n", expandedVol.CapacityCYL)
	//all ok delete Volume
	cleanupVolume(vol.VolumeID, volumeName, defaultStorageGroup, t)
}
