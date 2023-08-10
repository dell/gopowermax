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
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	pmax "github.com/dell/gopowermax/v2"
	types "github.com/dell/gopowermax/v2/types/v100"
)

var (
	snapID                     string
	sourceVolume, targetVolume *types.Volume
)

// getOrCreateVolumes creates a volume pair(source and target) for snapshot testing.
// In case either of the volumes exit, returns the already created volume.
func getOrCreateVolumes(client pmax.Pmax, target bool) (*types.Volume, *types.Volume, error) {
	if sourceVolume != nil && !target || targetVolume != nil {
		return sourceVolume, targetVolume, nil
	}
	type Vol struct {
		Volume *types.Volume
		Err    error
	}
	volChan := make(chan Vol, 2)
	maxIters := 0
	if sourceVolume == nil && target {
		maxIters = 2
	} else {
		maxIters = 1
	}
	for i := 0; i < maxIters; i++ {
		go func(i int) {
			volumeName := fmt.Sprintf("csi%s-Int%d%d", volumePrefix, time.Now().Nanosecond(), i)
			volOpts := make(map[string]interface{})
			var vol Vol
			vol.Volume, vol.Err = client.CreateVolumeInStorageGroup(context.TODO(), symmetrixID, defaultStorageGroup, volumeName, 1, volOpts)
			if vol.Err != nil && strings.Contains(vol.Err.Error(), "Failed to find newly created volume with name") {
				time.Sleep(2 * time.Second)
				ids, err := client.GetVolumeIDList(context.TODO(), symmetrixID, volumeName, false)
				if err == nil && len(ids) > 0 {
					vol.Volume, vol.Err = client.GetVolumeByID(context.TODO(), symmetrixID, ids[0])
				}
			}
			volChan <- vol
		}(i)
	}
	for j := 0; j < maxIters; j++ {
		vol := <-volChan
		if vol.Err != nil {
			return nil, nil, vol.Err
		}
		if sourceVolume == nil {
			sourceVolume = vol.Volume
		} else {
			targetVolume = vol.Volume
		}
	}

	return sourceVolume, targetVolume, nil
}

func createSnapshot(sourceVolumeList []types.VolumeList) (string, error) {
	snapshotName := fmt.Sprintf("csi%s-int%d", snapshotPrefix, time.Now().Nanosecond())
	err := client.CreateSnapshot(context.TODO(), symmetrixID, snapshotName, sourceVolumeList, 0)
	if err != nil {
		return "", err
	}
	return snapshotName, nil
}

func getOrCreateSnapshot(volumeID string, client pmax.Pmax) (string, error) {
	var err error
	if snapID == "" {
		sourceVolumeList := []types.VolumeList{{Name: volumeID}}
		snapID, err = createSnapshot(sourceVolumeList)
	}
	return snapID, err
}

func TestCrudStorageGroupSnapshot(t *testing.T) {
	snapshotSgName := "integration-test-snapshot-sg"
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	// Create a storage group
	_, err := createStorageGroup(symmetrixID, snapshotSgName, defaultSRP, defaultServiceLevel, false, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Create a volume inside the storage group
	now := time.Now()
	volumeName := fmt.Sprintf("csi%s-Int%d", volumePrefix, now.Nanosecond())
	capUnit := "TB"
	fmt.Printf("volumeName: %s\n", volumeName)
	volopts := make(map[string]interface{})
	volopts["capacityUnit"] = capUnit
	vol, errr := client.CreateVolumeInStorageGroupS(context.TODO(), symmetrixID, snapshotSgName, volumeName, "1", volopts)
	if errr != nil {
		fmt.Println(errr.Error())
		return
	}

	// Start tests
	optionalPayload := &types.CreateStorageGroupSnapshot{
		SnapshotName:    defaultSnapshotName,
		ExecutionOption: types.ExecutionOptionSynchronous,
	}
	// Create the new Snapshot
	sgSnap, err := client.CreateStorageGroupSnapshot(context.TODO(), symmetrixID, snapshotSgName, optionalPayload)
	if err != nil {
		t.Error("Error creating a storage group snapshot: " + err.Error())
		// Cleanup temp sg
		cleanupVolume(vol.VolumeID, volumeName, snapshotSgName, t)
		deleteStorageGroup(symmetrixID, snapshotSgName)
		return
	}
	fmt.Printf("Successfully created snapshot (%v)\n", sgSnap)

	// Get the list of SnapIds
	sgSnapIds, err := client.GetStorageGroupSnapshotSnapIds(context.TODO(), symmetrixID, snapshotSgName, defaultSnapshotName)
	if err != nil {
		t.Error("Error getting the snapshot snapid details on storage group: " + err.Error())
		// Cleanup temp sg
		cleanupVolume(vol.VolumeID, volumeName, snapshotSgName, t)
		deleteStorageGroup(symmetrixID, snapshotSgName)
		return
	}
	fmt.Printf("Successfully fetched snapshot ids (%v) \n\n", sgSnapIds)

	snapid := strconv.FormatInt(sgSnapIds.SnapIds[0], 10)

	// Link snap
	modifyPayloadLink := &types.ModifyStorageGroupSnapshot{
		Action:          string(pmax.Link),
		ExecutionOption: types.ExecutionOptionSynchronous,
		Link: types.LinkSnapshotAction{
			StorageGroupName: snapshotSgName + "02",
		},
	}
	modLinkSnap, err := client.ModifyStorageGroupSnapshot(context.TODO(), symmetrixID, snapshotSgName, defaultSnapshotName, snapid, modifyPayloadLink)
	if err != nil {
		t.Error("Error linking the snapshot: " + err.Error())
		//Cleanup temp sg
		cleanupVolume(vol.VolumeID, volumeName, snapshotSgName, t)
		deleteStorageGroup(symmetrixID, snapshotSgName)
		return
	}
	fmt.Printf("Successfully linked snapshot (%v)\n\n", modLinkSnap)

	// TTL snap
	modifyPayloadTTL := &types.ModifyStorageGroupSnapshot{
		Action:          string(pmax.SetTimeToLive),
		ExecutionOption: types.ExecutionOptionSynchronous,
		TimeToLive: types.TimeToLiveSnapshotAction{
			TimeToLive:  2,
			TimeInHours: true,
		},
	}
	modTTLSnap, err := client.ModifyStorageGroupSnapshot(context.TODO(), symmetrixID, snapshotSgName, defaultSnapshotName, snapid, modifyPayloadTTL)
	if err != nil {
		t.Error("Error setting TTL for the snapshot: " + err.Error())
		// Cleanup temp sg
		cleanupVolume(vol.VolumeID, volumeName, snapshotSgName, t)
		deleteStorageGroup(symmetrixID, snapshotSgName)
		return
	}
	fmt.Printf("Successfully set the TTL snapshot (%v)\n\n", modTTLSnap)

	// Unlink snap
	modifyPayloadUnlink := &types.ModifyStorageGroupSnapshot{
		Action:          string(pmax.Unlink),
		ExecutionOption: types.ExecutionOptionSynchronous,
		Unlink: types.UnlinkSnapshotAction{
			StorageGroupName: snapshotSgName + "02",
			Symforce:         true,
		},
	}
	modUnlinkSnap, err := client.ModifyStorageGroupSnapshot(context.TODO(), symmetrixID, snapshotSgName, defaultSnapshotName, snapid, modifyPayloadUnlink)
	if err != nil {
		t.Error("Error unlinking the snapshot: " + err.Error())
		// Cleanup temp sg
		cleanupVolume(vol.VolumeID, volumeName, snapshotSgName, t)
		deleteStorageGroup(symmetrixID, snapshotSgName)
		return
	}
	fmt.Printf("Successfully unlinked snapshot (%v)\n\n", modUnlinkSnap)

	// Rename Snap
	modifyPayloadRename := &types.ModifyStorageGroupSnapshot{
		Action:          string(pmax.Rename),
		ExecutionOption: types.ExecutionOptionSynchronous,
		Rename: types.RenameSnapshotAction{
			NewStorageGroupSnapshotName: "update-snapshot",
		},
	}

	modRenameSnap, err := client.ModifyStorageGroupSnapshot(context.TODO(), symmetrixID, snapshotSgName, defaultSnapshotName, snapid, modifyPayloadRename)
	if err != nil {
		t.Error("Error renaming the snapshot" + err.Error())
		// Cleanup temp sg
		cleanupVolume(vol.VolumeID, volumeName, snapshotSgName, t)
		deleteStorageGroup(symmetrixID, snapshotSgName)
		return
	}
	fmt.Printf("Successfully renamed snapshot (%v)\n\n", modRenameSnap)

	// Delete Snap
	err = client.DeleteStorageGroupSnapshot(context.TODO(), symmetrixID, snapshotSgName, "update-snapshot", snapid)
	if err != nil {
		t.Error("Error deleting snapshot group: " + err.Error())
		// Cleanup temp sg
		cleanupVolume(vol.VolumeID, volumeName, snapshotSgName, t)
		deleteStorageGroup(symmetrixID, snapshotSgName)
		return
	}
	fmt.Printf("Successfully deleted snapshot\n")
	// Cleanup temp sg
	cleanupVolume(vol.VolumeID, volumeName, snapshotSgName, t)
	deleteStorageGroup(symmetrixID, snapshotSgName)

}

func TestGetSnapVolumeList(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	snapVolumes, err := client.GetSnapVolumeList(context.TODO(), symmetrixID, nil)
	if err != nil {
		t.Error("Error getting the list of volumes with snapshots: " + err.Error())
		return
	}
	fmt.Printf("Volumes with snapshots: %v\n", snapVolumes)
}

func TestCreateSnapshot(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	srcVolume, _, err := getOrCreateVolumes(client, false)
	if err != nil {
		t.Error(err)
		return
	}
	sourceVolumeList := []types.VolumeList{{Name: srcVolume.VolumeID}}
	snapID = fmt.Sprintf("csi%s-int%d", snapshotPrefix, time.Now().Nanosecond())
	err = client.CreateSnapshot(context.TODO(), symmetrixID, snapID, sourceVolumeList, 0)
	if err != nil {
		t.Errorf("Error creating a snapshot(%s) on a volumes %v\n", snapID, sourceVolumeList)
		return
	}
	volumeSnapshot, err := client.GetSnapshotInfo(context.TODO(), symmetrixID, srcVolume.VolumeID, snapID)
	if err != nil {
		t.Errorf("Error %s fetching created snapshot(%s) on a volumes %v\n", err.Error(), snapID, sourceVolumeList)
		return
	}
	fmt.Printf("Snapshot(%s) created successfully\n", volumeSnapshot.SnapshotName)
}

func TestGetVolumeSnapInfo(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	volume, _, err := getOrCreateVolumes(client, false)
	if err != nil {
		t.Error(err)
		return
	}
	sourceVolumeList := []types.VolumeList{{Name: volume.VolumeID}}
	snapshotName, err := createSnapshot(sourceVolumeList)
	snapshotVolumeGeneration, err := client.GetVolumeSnapInfo(context.TODO(), symmetrixID, volume.VolumeID)
	if err != nil {
		t.Errorf("Error getting the snapshot on volume %s: %s", volume.VolumeIdentifier, err.Error())
		return
	}
	fmt.Printf("Snapshots on volume (%s): %v\n", volume.VolumeIdentifier, snapshotVolumeGeneration)
	if snapshotName != "" {
		err := client.DeleteSnapshot(context.TODO(), symmetrixID, snapshotName, sourceVolumeList, int64(0))
		if err != nil {
			t.Error(err)
		}
	}
}

func TestGetSnapshotInfo(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	}
	volume, _, err := getOrCreateVolumes(client, false)
	if err != nil {
		t.Error(err)
		return
	}
	snapshotName, err := getOrCreateSnapshot(volume.VolumeID, client)
	if err != nil {
		t.Error(err)
		return
	}
	volumeSnapshot, err := client.GetSnapshotInfo(context.TODO(), symmetrixID, volume.VolumeID, snapshotName)
	if err != nil {
		t.Errorf("Error getting snapshot(%s) details: %s", snapshotName, err.Error())
		return
	}
	fmt.Printf("Snapshot(%s): %v\n", snapshotName, volumeSnapshot)
}

func TestGetSnapshotGenerations(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	}
	volume, _, err := getOrCreateVolumes(client, false)
	if err != nil {
		t.Error(err)
		return
	}
	snapshotName, err := getOrCreateSnapshot(volume.VolumeID, client)
	if err != nil {
		t.Error(err)
		return
	}
	volumeSnapshotGenerations, err := client.GetSnapshotGenerations(context.TODO(), symmetrixID, volume.VolumeID, snapshotName)
	if err != nil {
		t.Errorf("Error fetching generations on the snapshot(%s): %s\n", snapshotName, err.Error())
	}
	fmt.Printf("Snapshot(%s) generations fetched successfully: %v\n", snapshotName, volumeSnapshotGenerations)
}

func TestGetSnapshotGenerationInfo(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Errorf(err.Error())
			return
		}
	}
	volume, _, err := getOrCreateVolumes(client, false)
	if err != nil {
		t.Error(err)
		return
	}
	snapshotName, err := getOrCreateSnapshot(volume.VolumeID, client)
	if err != nil {
		t.Error(err)
		return
	}
	var generation int64
	volumeSnapshotGeneration, err := client.GetSnapshotGenerationInfo(context.TODO(), symmetrixID, volume.VolumeID, snapshotName, generation)
	if err != nil {
		t.Errorf("Error fetching generation(%d) info on the snapshot(%s): %s\n", generation, snapshotName, err.Error())
	}
	if volumeSnapshotGeneration.Generation == generation {
		fmt.Printf("Snapshot(%s) generation(%d) fetched successfully: %v\n", snapshotName, generation, volumeSnapshotGeneration)
	} else {
		t.Errorf("Returned generation is not same as the expected one.")
	}
}

func TestSnapshotLinkage(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	srcVolume, tgtVolume, err := getOrCreateVolumes(client, true)
	if err != nil {
		t.Error(err)
		return
	}
	sourceVolumeList := []types.VolumeList{{Name: srcVolume.VolumeID}}
	targetVolumeList := []types.VolumeList{{Name: tgtVolume.VolumeID}}
	snapshotName, err := getOrCreateSnapshot(sourceVolume.VolumeID, client)
	if err != nil {
		t.Error(err)
		return
	}

	// Link snapshot
	err = modifySnapshotLink(sourceVolumeList, targetVolumeList, "Link", snapshotName, t)
	if err != nil {
		t.Error(err)
		return
	}
	//Unlink snapshot
	err = modifySnapshotLink(sourceVolumeList, targetVolumeList, "Unlink", snapshotName, t)
	if err != nil {
		t.Error(err)
		return
	}
}

func modifySnapshotLink(sourceVolumeList, targetVolumeList []types.VolumeList, operation, snapshotName string, t *testing.T) error {
	err := client.ModifySnapshot(context.TODO(), symmetrixID, sourceVolumeList, targetVolumeList, snapshotName, operation, "", 0)
	if err != nil {
		return fmt.Errorf("Error %sing snapshot(%s)", strings.ToLower(operation), snapshotName)
	}

	fmt.Printf("Snapshot(%s) %sed successfully\n", strings.ToLower(operation), snapshotName)
	return nil
}

func TestSnapshotRenaming(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	srcVolume, tgtVolume, err := getOrCreateVolumes(client, true)
	if err != nil {
		t.Error(err)
		return
	}
	sourceVolumeList := []types.VolumeList{{Name: srcVolume.VolumeID}}
	targetVolumeList := []types.VolumeList{{Name: tgtVolume.VolumeID}}
	snapshotName, err := getOrCreateSnapshot(sourceVolume.VolumeID, client)
	if err != nil {
		t.Error(err)
		return
	}
	newSnapID := snapshotName + "_renamed"
	err = renameSnapshot(symmetrixID, snapshotName, newSnapID, 0, sourceVolumeList, targetVolumeList)
	if err != nil {
		t.Error(err)
		return
	}
	volumeSnapshot, err := client.GetSnapshotInfo(context.TODO(), symmetrixID, srcVolume.VolumeID, newSnapID)
	if err != nil {
		t.Errorf("Error fetching renamed snapshot(%s) on a volumes %v\n", snapID, sourceVolumeList)
		return
	}
	if volumeSnapshot.SnapshotName == newSnapID {
		fmt.Printf("Snapshot(%s) renamed to %s successully\n", snapshotName, newSnapID)
	} else {
		fmt.Printf("Renaming snapshot failed")
	}

	// Revert back the change
	err = renameSnapshot(symmetrixID, newSnapID, snapshotName, 0, sourceVolumeList, targetVolumeList)
	if err != nil {
		t.Error(err)
		return
	}
	volumeSnapshot, err = client.GetSnapshotInfo(context.TODO(), symmetrixID, srcVolume.VolumeID, snapshotName)
	if err != nil {
		t.Errorf("Error fetching renamed snapshot(%s) on a volumes %v\n", snapID, sourceVolumeList)
		return
	}
	if volumeSnapshot.SnapshotName == snapshotName {
		fmt.Printf("Snapshot(%s) renamed to %s successully\n", newSnapID, snapshotName)
	} else {
		fmt.Printf("Renaming snapshot failed")
	}
}

func renameSnapshot(symmetrixID, snapshotName, newSnapID string, generation int, sourceVolumeList, targetVolumeList []types.VolumeList) error {
	err := client.ModifySnapshot(context.TODO(), symmetrixID, sourceVolumeList, targetVolumeList, snapshotName, "Rename", newSnapID, 0)
	if err != nil {
		return fmt.Errorf("Error renaming snapshot: %s", err.Error())
	}
	return nil
}

/*func TestSnapshotRestore(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	srcVolume, tgtVolume, err := getOrCreateVolumes(client, true)
	if err != nil {
		t.Error(err)
		return
	}
	sourceVolumeList := []types.VolumeList{{Name: srcVolume.VolumeID}}
	targetVolumeList := []types.VolumeList{{Name: tgtVolume.VolumeID}}

	snapshotName, err := getOrCreateSnapshot(sourceVolume.VolumeID, client)
	if err != nil {
		t.Error(err)
		return
	}
	err = client.ModifySnapshot(symmetrixID, sourceVolumeList, targetVolumeList, snapshotName, "Restore", "", 0)
	if err != nil {
		t.Errorf("Error restoring the snapshot(%s): %s", snapshotName, err.Error())
		return
	}
	err = client.DeleteSnapshot(symmetrixID, snapshotName, sourceVolumeList, int64(0))
	if err != nil {
		t.Errorf("Failed to terminate restore session for snapshot(%s)", snapshotName)
		return
	}
	fmt.Printf("Snapshot(%s) restored successfully\n", snapshotName)
}*/

/*func TestPrivateVolumeListing(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	volume, _, err := getOrCreateVolumes(client, false)
	if err != nil {
		t.Error(err)
		return
	}
	privateVolumeList, err := client.GetPrivVolumeIDList(symmetrixID, volume.VolumeID, true)
	if err != nil {
		t.Errorf("Error fetching the list of private volumes: %s", err.Error())
		return
	}
	fmt.Printf("Private volumes fetched successfully: %v\n", privateVolumeList[0])
}*/

func TestGetPrivVolumeByID(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	volume, _, err := getOrCreateVolumes(client, false)
	if err != nil {
		t.Error(err)
		return
	}
	privateVolume, err := client.GetPrivVolumeByID(context.TODO(), symmetrixID, volume.VolumeID)
	if err != nil {
		t.Errorf("Error fetching a private volume: %s", err.Error())
		return
	}
	fmt.Printf("Private volume fetched successfully: %v\n", privateVolume)
}

func TestGetRDFGroup(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}

	rdfGrpInfo, err := client.GetRDFGroupByID(context.TODO(), symmetrixID, localRDFGrpNo)
	if err != nil {
		t.Errorf("Error fetching RDF Group Information : %s", err.Error())
		return
	}
	fmt.Printf("RDF Information fetched successfully: %v\n", rdfGrpInfo)

}

func TestGetProtectedStorageGroup(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}

	rdfSgInfo, err := client.GetProtectedStorageGroup(context.TODO(), symmetrixID, defaultProtectedStorageGroup)
	if err != nil {
		t.Errorf("Error fetching Protected Storage Group Information : %s", err.Error())
		return
	}
	fmt.Printf("Protected Storage Group Information fetched successfully: %v\n", rdfSgInfo)
}

func TestGetStorageGroupRDFInfo(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}

	rdfSgInfo, err := client.GetStorageGroupRDFInfo(context.TODO(), symmetrixID, defaultProtectedStorageGroup, localRDFGrpNo)
	if err != nil {
		t.Errorf("Error fetching RDF Information for storage group: %s", err.Error())
		return
	}
	fmt.Printf("RDF Information for storage group fetched successfully: %v\n", rdfSgInfo)
}
func TestGetRDFDevicePairInfo(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	rdfPair, err := client.GetRDFDevicePairInfo(context.TODO(), symmetrixID, localRDFGrpNo, localVol.VolumeID)
	if err != nil {
		t.Errorf("Error retrieving RDF device pair information: %s", err.Error())
		return
	}

	fmt.Printf("RDF device pair information retrieved successfully: %v\n", rdfPair)
}

func TestCreateVolumeInProtectedStorageGroupS(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	now := time.Now()
	volumeName := fmt.Sprintf("csi%s-Int%d", volumePrefix, now.Nanosecond())
	volOpts := make(map[string]interface{})

	vol, err := client.CreateVolumeInProtectedStorageGroupS(context.TODO(), symmetrixID, remoteSymmetrixID, defaultProtectedStorageGroup, defaultProtectedStorageGroup, volumeName, 30, volOpts)
	if err != nil {
		t.Errorf("Error Creating Volume in Protected Storage Group: %s", err.Error())
		return
	}
	fmt.Printf("Volume in Protected Storage Group created successfully: %v\n", vol)
	fmt.Printf("Waiting for 8 minutes \n")
	time.Sleep(500 * time.Second)
	cleanupRDFPair(vol.VolumeID, volumeName, defaultProtectedStorageGroup, t)
}
func TestAddVolumesToProtectedStorageGroup(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	now := time.Now()
	volumeName := fmt.Sprintf("csi%s-Int%d", volumePrefix, now.Nanosecond())
	volOpts := make(map[string]interface{})
	vol, err := client.CreateVolumeInStorageGroup(context.TODO(), symmetrixID, nonFASTManagedSG, volumeName, 50, volOpts)
	if err != nil {
		t.Error(err)
		return
	}
	err = client.AddVolumesToProtectedStorageGroup(context.TODO(), symmetrixID, defaultProtectedStorageGroup, remoteSymmetrixID, defaultProtectedStorageGroup, true, vol.VolumeID)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("Waiting for 5 minutes \n")
	time.Sleep(300 * time.Second)
	sg, err := client.RemoveVolumesFromStorageGroup(context.TODO(), symmetrixID, nonFASTManagedSG, true, vol.VolumeID)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("SG after removing volume: %#v\n", sg)
	cleanupRDFPair(vol.VolumeID, volumeName, defaultProtectedStorageGroup, t)
}

func TestExecuteReplicationActionOnSG(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	now := time.Now()
	volumeName := fmt.Sprintf("csi%s-Int%d", volumePrefix, now.Nanosecond())
	volOpts := make(map[string]interface{})
	vol, err := client.CreateVolumeInProtectedStorageGroupS(context.TODO(), symmetrixID, remoteSymmetrixID, defaultProtectedStorageGroup, defaultProtectedStorageGroup, volumeName, 30, volOpts)
	if err != nil {
		t.Errorf("Error Creating Volume in Protected Storage Group: %s", err.Error())
		return
	}
	fmt.Printf("Volume in Protected Storage Group created successfully: %v\n", vol)

	fmt.Printf("Waiting for 10 minutes \n")
	time.Sleep(600 * time.Second)

	err = client.ExecuteReplicationActionOnSG(context.TODO(), symmetrixID, "Suspend", defaultProtectedStorageGroup, localRDFGrpNo, true, true, false)
	if err != nil {
		t.Errorf("Error in suspending the RDF relation in Protected Storage Group: %s", err.Error())
		return
	}

	err = client.ExecuteReplicationActionOnSG(context.TODO(), symmetrixID, "Resume", defaultProtectedStorageGroup, localRDFGrpNo, true, true, false)
	if err != nil {
		t.Errorf("Error in resuming the RDF relation in Protected Storage Group: %s", err.Error())
		return
	}

	cleanupRDFPair(vol.VolumeID, volumeName, defaultProtectedStorageGroup, t)
}

/*func TestSnapSessionFetch(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err.Error())
			return
		}
	}
	srcVolume, tgtVolume, err := getOrCreateVolumes(client, true)
	if err != nil {
		t.Error(err)
		return
	}
	snapshotName, err := getOrCreateSnapshot(sourceVolume.VolumeID, client)
	if err != nil {
		t.Error(err)
		return
	}
	sourceVolumeList := []types.VolumeList{{Name: srcVolume.VolumeID}}
	targetVolumeList := []types.VolumeList{{Name: tgtVolume.VolumeID}}
	modifySnapshotLink(sourceVolumeList, targetVolumeList, "Link", snapshotName, t)

	snapSession, targetSession, err := client.GetSnapSessions(symmetrixID, sourceVolume.VolumeID)
	if err != nil {
		t.Errorf("Error fetching the snapshot sessions: %s", err.Error())
		return
	}
	modifySnapshotLink(sourceVolumeList, targetVolumeList, "Unlink", snapshotName, t)
	fmt.Printf("Snap and target sessions fetched successfully: %v\v%v\n", snapSession, targetSession)
}*/

func TestDeleteSnapshot(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	volume, _, err := getOrCreateVolumes(client, false)
	if err != nil {
		t.Error(err)
		return
	}
	sourceVolumeList := []types.VolumeList{{Name: volume.VolumeID}}
	snapshotName, err := getOrCreateSnapshot(volume.VolumeID, client)
	if err != nil {
		t.Error(err)
		return
	}
	err = client.DeleteSnapshot(context.TODO(), symmetrixID, snapshotName, sourceVolumeList, int64(0))
	if err != nil {
		t.Errorf("Failed to get snapshot (%s) deletion status: %s", snapshotName, err.Error())
		return
	}
	fmt.Printf("Snapshot (%s) deleted successfully\n", snapshotName)
	snapID = ""
}

func TestGetReplicationCapabilities(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	symReplicationCapabilities, err := client.GetReplicationCapabilities(context.TODO())
	if err != nil {
		t.Errorf("Failed to fetch replication capabilities: %s", err.Error())
		return
	}
	fmt.Printf("Replication capabilities: %v\n", symReplicationCapabilities)
}

// afterRun gets invoked after the tests run, to clean the snapshot
// and volumes created for testing purposes.
func afterRun(tests []testing.InternalTest) {
	cleanup := []testing.InternalTest{}
	if snapID != "" {
		cleanup = append(cleanup, testing.InternalTest{
			Name: "TestSnapshotDeletion",
			F:    TestDeleteSnapshot,
		})
	}
	if sourceVolume != nil {
		cleanup = append(cleanup, testing.InternalTest{
			Name: "volumeCleanup",
			F:    volumeCleanup(sourceVolume.VolumeID, sourceVolume.VolumeIdentifier, defaultStorageGroup),
		})
	}
	if targetVolume != nil {
		cleanup = append(cleanup, testing.InternalTest{
			Name: "volumeCleanup",
			F:    volumeCleanup(targetVolume.VolumeID, targetVolume.VolumeIdentifier, defaultStorageGroup),
		})
	}
	cleanup = append(cleanup, tests...)
	testing.Main(func(pat, str string) (bool, error) {
		return true, nil
	}, cleanup, nil, nil)
}

func volumeCleanup(volumeID string, volumeName string, storageGroup string) func(*testing.T) {
	return func(t *testing.T) {
		if volumeName != "" {
			vol, err := client.RenameVolume(context.TODO(), symmetrixID, volumeID, "_DEL"+volumeName)
			if err != nil {
				t.Error(err)
				return
			}
			fmt.Printf("volume Renamed: %s\n", vol.VolumeIdentifier)
		}
		sg, err := client.RemoveVolumesFromStorageGroup(context.TODO(), symmetrixID, storageGroup, true, volumeID)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("SG after removing volume: %#v\n", sg)
		pmax.Debug = true
		err = client.DeleteVolume(context.TODO(), symmetrixID, volumeID)
		if err != nil {
			t.Error("DeleteVolume failed: " + err.Error())
		}
		// Test deletion of the volume again... should return an error
		err = client.DeleteVolume(context.TODO(), symmetrixID, volumeID)
		if err == nil {
			t.Error("Expected an error saying volume was not found, but no error")
		}
		fmt.Printf("Received expected error: %s\n", err.Error())
	}
}
func cleanupRDFPair(volumeID string, volumeName string, storageGroup string, t *testing.T) {
	fmt.Println("Cleaning up RDF Pair...")

	//Retrieving remote volume information

	rdfPair, err := client.GetRDFDevicePairInfo(context.TODO(), symmetrixID, localRDFGrpNo, volumeID)
	if err != nil {
		t.Errorf("Error retrieving RDF device pair information: %s", err.Error())
		return
	}

	//Terminating the Pair and removing the volumes from local SG and remote SG

	_, err = client.RemoveVolumesFromProtectedStorageGroup(context.TODO(), symmetrixID, defaultProtectedStorageGroup, remoteSymmetrixID, defaultProtectedStorageGroup, true, volumeID)
	if err != nil {
		t.Errorf("failed to remove volumes from default Protected SG (%s) : (%s)", defaultProtectedStorageGroup, err.Error())
	}

	//Deleting local volume

	err = client.DeleteVolume(context.TODO(), symmetrixID, volumeID)
	if err != nil {
		t.Error("DeleteVolume failed: " + err.Error())
	}

	// Test deletion of the volume again... should return an error
	err = client.DeleteVolume(context.TODO(), symmetrixID, volumeID)
	if err == nil {
		t.Error("Expected an error saying volume was not found, but no error")
	}
	fmt.Printf("Received expected error: %s\n", err.Error())

	//Deleting remote volume

	err = client.DeleteVolume(context.TODO(), remoteSymmetrixID, rdfPair.RemoteVolumeName)
	if err != nil {
		t.Error("DeleteVolume failed: " + err.Error())
	}
	// Test deletion of the volume again... should return an error
	err = client.DeleteVolume(context.TODO(), remoteSymmetrixID, rdfPair.RemoteVolumeName)
	if err == nil {
		t.Error("Expected an error saying volume was not found, but no error")
	}
	fmt.Printf("Received expected error: %s\n", err.Error())

}

func TestGetFreeLocalAndRemoteRDFg(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}

	rdfGrpInfo, err := client.GetFreeLocalAndRemoteRDFg(context.TODO(), symmetrixID, remoteSymmetrixID)
	if err != nil {
		t.Errorf("Error fetching Free RDF Group Information : %s", err.Error())
		return
	}
	fmt.Printf(" Free RDF Information fetched successfully: %v\n", rdfGrpInfo)
}

func TestGetLocalOnlineRDFDirs(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}

	rdfGrpInfo, err := client.GetLocalOnlineRDFDirs(context.TODO(), symmetrixID)
	if err != nil {
		t.Errorf("Error fetching local Online RDF Dirs Information : %s", err.Error())
		return
	}
	fmt.Printf("Local Online RDF Dirs fetched successfully: %v\n", rdfGrpInfo)
}

func TestCreateSnapshotPolicy(t *testing.T) {

	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	localSnapshotPolicyDetails := &types.LocalSnapshotPolicyDetails{
		Secure:        false,
		SnapshotCount: 24,
	}

	optionalPayload := make(map[string]interface{})
	optionalPayload["localSnapshotPolicyDetails"] = localSnapshotPolicyDetails

	targets, err := client.CreateSnapshotPolicy(context.TODO(), symmetrixID, "WeeklyDefaultnewTest", "1 Hour", 10, 2, 2, optionalPayload)
	if err != nil {
		t.Error("Error Creating snapshot policy" + err.Error())
		return
	}
	fmt.Printf("Created snapshot policy: %v\n", targets.SnapshotPolicyName)
}

func TestGetSnapshotPolicy(t *testing.T) {

	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}

	targets, err := client.GetSnapshotPolicy(context.TODO(), symmetrixID, "WeeklyDefaultnewTest")
	if err != nil {
		t.Error("Error calling GetSnapshotPolicy " + err.Error())
		return
	}
	fmt.Printf("Snapshot Policy name: %v\n", targets.SnapshotPolicyName)
}

func TestUpdateSnapshotPolicyForModify(t *testing.T) {

	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	modifySnapshotPolicyParam := &types.ModifySnapshotPolicyParam{
		SnapshotPolicyName: "WeeklyDefaultnewTest1",
		IntervalMinutes:    60,
		OffsetMins:         10,
	}

	optionalPayload := make(map[string]interface{})
	optionalPayload["modify"] = modifySnapshotPolicyParam

	error := client.UpdateSnapshotPolicy(context.TODO(), symmetrixID, "Modify", "WeeklyDefaultnewTest", optionalPayload)
	if error != nil {
		t.Error("Error Updating snapshot policy " + error.Error())
		return
	}
	fmt.Printf("Updated Snapshot Policy: Modify")
}

func TestUpdateSnapshotPolicyForAddStorageGroup(t *testing.T) {

	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	snapshotPolicySgName := "test-snapshot-policy-sg"
	// Create a storage group
	_, err := createStorageGroup(symmetrixID, snapshotPolicySgName, "None", defaultServiceLevel, false, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	associateStorageGroupParam := &types.AssociateStorageGroupParam{
		StorageGroupName: []string{snapshotPolicySgName},
	}

	optionalPayload := make(map[string]interface{})
	optionalPayload["associateStorageGroupParam"] = associateStorageGroupParam

	error := client.UpdateSnapshotPolicy(context.TODO(), symmetrixID, "AssociateToStorageGroups", "WeeklyDefaultnewTest1", optionalPayload)
	if error != nil {
		t.Error("Error Updating snapshot policy " + error.Error())
		deleteStorageGroup(symmetrixID, snapshotPolicySgName)
		return
	}
	fmt.Printf("Updated Snapshot Policy: AssociateToStorageGroups")
}
func TestUpdateSnapshotPolicyForRemoveStorageGroup(t *testing.T) {

	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	snapshotPolicySgName := "test-snapshot-policy-sg"
	disassociateStorageGroupParam := &types.DisassociateStorageGroupParam{
		StorageGroupName: []string{snapshotPolicySgName},
	}

	optionalPayload := make(map[string]interface{})
	optionalPayload["disassociateStorageGroupParam"] = disassociateStorageGroupParam

	error := client.UpdateSnapshotPolicy(context.TODO(), symmetrixID, "DisassociateFromStorageGroups", "WeeklyDefaultnewTest1", optionalPayload)
	if error != nil {
		t.Error("Error Updating snapshot policy " + error.Error())
		return
	}
	fmt.Printf("Updated Snapshot Policy: DisassociateFromStorageGroups")
}

func TestGetSnapshotPolicyList(t *testing.T) {

	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	targets, err := client.GetSnapshotPolicyList(context.TODO(), symmetrixID)
	if err != nil {
		t.Error("Error calling GetSnapshotPolicyList " + err.Error())
		return
	}
	fmt.Printf("Snapshot Policy names: %v\n", targets.SnapshotPolicyIds)
}

func TestDeleteSnapshotPolicy(t *testing.T) {

	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}

	error := client.DeleteSnapshotPolicy(context.TODO(), symmetrixID, "WeeklyDefaultnewTest1")
	if error != nil {
		t.Error("Error Deleting Snapshot Policy " + error.Error())
		return
	}
	fmt.Printf("Deleted snapshot policy")
}
