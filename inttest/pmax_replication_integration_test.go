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
	"fmt"
	"strings"
	"testing"
	"time"

	pmax "github.com/dell/gopowermax"
	"github.com/dell/gopowermax/types/v90"
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
			var vol Vol
			vol.Volume, vol.Err = client.CreateVolumeInStorageGroup(symmetrixID, defaultStorageGroup, volumeName, 1)
			if vol.Err != nil && strings.Contains(vol.Err.Error(), "Failed to find newly created volume with name") {
				time.Sleep(2 * time.Second)
				ids, err := client.GetVolumeIDList(symmetrixID, volumeName, false)
				if err == nil && len(ids) > 0 {
					vol.Volume, vol.Err = client.GetVolumeByID(symmetrixID, ids[0])
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
	err := client.CreateSnapshot(symmetrixID, snapshotName, sourceVolumeList, 0)
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

func TestGetSnapVolumeList(t *testing.T) {
	if client == nil {
		err := getClient()
		if err != nil {
			t.Error(err)
			return
		}
	}
	snapVolumes, err := client.GetSnapVolumeList(symmetrixID, nil)
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
	err = client.CreateSnapshot(symmetrixID, snapID, sourceVolumeList, 0)
	if err != nil {
		t.Errorf("Error creating a snapshot(%s) on a volumes %v\n", snapID, sourceVolumeList)
		return
	}
	volumeSnapshot, err := client.GetSnapshotInfo(symmetrixID, srcVolume.VolumeID, snapID)
	if err != nil {
		t.Errorf("Error fetching created snapshot(%s) on a volumes %v\n", snapID, sourceVolumeList)
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
	snapshotVolumeGeneration, err := client.GetVolumeSnapInfo(symmetrixID, volume.VolumeID)
	if err != nil {
		t.Errorf("Error getting the snapshot on volume %s: %s", volume.VolumeIdentifier, err.Error())
		return
	}
	fmt.Printf("Snapshots on volume (%s): %v\n", volume.VolumeIdentifier, snapshotVolumeGeneration)
	if snapshotName != "" {
		err := client.DeleteSnapshot(symmetrixID, snapshotName, sourceVolumeList, int64(0))
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
	volumeSnapshot, err := client.GetSnapshotInfo(symmetrixID, volume.VolumeID, snapshotName)
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
	volumeSnapshotGenerations, err := client.GetSnapshotGenerations(symmetrixID, volume.VolumeID, snapshotName)
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
	volumeSnapshotGeneration, err := client.GetSnapshotGenerationInfo(symmetrixID, volume.VolumeID, snapshotName, generation)
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
	err := client.ModifySnapshot(symmetrixID, sourceVolumeList, targetVolumeList, snapshotName, operation, "", 0)
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
	volumeSnapshot, err := client.GetSnapshotInfo(symmetrixID, srcVolume.VolumeID, newSnapID)
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
	volumeSnapshot, err = client.GetSnapshotInfo(symmetrixID, srcVolume.VolumeID, snapshotName)
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
	err := client.ModifySnapshot(symmetrixID, sourceVolumeList, targetVolumeList, snapshotName, "Rename", newSnapID, 0)
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
	privateVolume, err := client.GetPrivVolumeByID(symmetrixID, volume.VolumeID)
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

	rdfGrpInfo, err := client.GetRDFGroup(symmetrixID, localRDFGrpNo)
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

	rdfSgInfo, err := client.GetProtectedStorageGroup(symmetrixID, defaultProtectedStorageGroup)
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

	rdfSgInfo, err := client.GetStorageGroupRDFInfo(symmetrixID, defaultProtectedStorageGroup, localRDFGrpNo)
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
	rdfPair, err := client.GetRDFDevicePairInfo(symmetrixID, localRDFGrpNo, localVol.VolumeID)
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
	vol, err := client.CreateVolumeInProtectedStorageGroupS(symmetrixID, remoteSymmetrixID, defaultProtectedStorageGroup, defaultProtectedStorageGroup, volumeName, 30)
	if err != nil {
		t.Errorf("Error Creating Volume in Protected Storage Group: %s", err.Error())
		return
	}
	fmt.Printf("Volume in Protected Storage Group created successfully: %v\n", vol)
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
	vol, err := client.CreateVolumeInStorageGroup(symmetrixID, nonFASTManagedSG, volumeName, 50)
	if err != nil {
		t.Error(err)
		return
	}
	err = client.AddVolumesToProtectedStorageGroup(symmetrixID, defaultProtectedStorageGroup, remoteSymmetrixID, defaultProtectedStorageGroup, true, vol.VolumeID)
	if err != nil {
		t.Error(err)
		return
	}
	sg, err := client.RemoveVolumesFromStorageGroup(symmetrixID, nonFASTManagedSG, true, vol.VolumeID)
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
	vol, err := client.CreateVolumeInProtectedStorageGroupS(symmetrixID, remoteSymmetrixID, defaultProtectedStorageGroup, defaultProtectedStorageGroup, volumeName, 30)
	if err != nil {
		t.Errorf("Error Creating Volume in Protected Storage Group: %s", err.Error())
		return
	}
	fmt.Printf("Volume in Protected Storage Group created successfully: %v\n", vol)

	err = client.ExecuteReplicationActionOnSG(symmetrixID, "Suspend", defaultProtectedStorageGroup, localRDFGrpNo, true, true)
	if err != nil {
		t.Errorf("Error in suspending the RDF relation in Protected Storage Group: %s", err.Error())
		return
	}

	err = client.ExecuteReplicationActionOnSG(symmetrixID, "Resume", defaultProtectedStorageGroup, localRDFGrpNo, true, true)
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
	err = client.DeleteSnapshot(symmetrixID, snapshotName, sourceVolumeList, int64(0))
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
	symReplicationCapabilities, err := client.GetReplicationCapabilities()
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
			vol, err := client.RenameVolume(symmetrixID, volumeID, "_DEL"+volumeName)
			if err != nil {
				t.Error(err)
				return
			}
			fmt.Printf("volume Renamed: %s\n", vol.VolumeIdentifier)
		}
		sg, err := client.RemoveVolumesFromStorageGroup(symmetrixID, storageGroup, true, volumeID)
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Printf("SG after removing volume: %#v\n", sg)
		pmax.Debug = true
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
}
func cleanupRDFPair(volumeID string, volumeName string, storageGroup string, t *testing.T) {
	fmt.Println("Cleaning up RDF Pair...")

	//Retrieving remote volume information

	rdfPair, err := client.GetRDFDevicePairInfo(symmetrixID, localRDFGrpNo, volumeID)
	if err != nil {
		t.Errorf("Error retrieving RDF device pair information: %s", err.Error())
		return
	}

	//Terminating the Pair and removing the volumes from local SG and remote SG

	_, err = client.RemoveVolumesFromProtectedStorageGroup(symmetrixID, defaultProtectedStorageGroup, remoteSymmetrixID, defaultProtectedStorageGroup, true, volumeID)
	if err != nil {
		t.Errorf("failed to remove volumes from default Protected SG (%s) : (%s)", defaultProtectedStorageGroup, err.Error())
	}

	//Deleting local volume

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

	//Deleting remote volume

	err = client.DeleteVolume(remoteSymmetrixID, rdfPair.RemoteVolumeName)
	if err != nil {
		t.Error("DeleteVolume failed: " + err.Error())
	}
	// Test deletion of the volume again... should return an error
	err = client.DeleteVolume(remoteSymmetrixID, rdfPair.RemoteVolumeName)
	if err == nil {
		t.Error("Expected an error saying volume was not found, but no error")
	}
	fmt.Printf("Received expected error: %s\n", err.Error())

}
