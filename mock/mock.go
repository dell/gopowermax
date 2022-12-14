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

package mock

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	types "github.com/dell/gopowermax/v2/types/v100"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
)

// constants
const (
	APIVersion                   = "{apiversion}"
	PREFIX                       = "/univmax/restapi/" + APIVersion
	PREFIXNOVERSION              = "/univmax/restapi"
	PRIVATEPREFIX                = "/univmax/restapi/private/" + APIVersion
	defaultUsername              = "username"
	defaultPassword              = "password"
	Debug                        = false
	DefaultStorageGroup          = "CSI-Test-SG-1"
	DefaultStorageGroup1         = "CSI-Test-SG-2"
	DefaultProtectedStorageGroup = "CSI-no-srp-async-test-13"
	DefaultSymmetrixID           = "000197900046"
	DefaultRemoteSymID           = "000000000013"
	PostELMSRSymmetrixID         = "000197900047"
	DefaultStoragePool           = "SRP_1"
	DefaultServiceLevel          = "Optimized"
	DefaultFcStoragePortWWN      = "5000000000000001"
	DefaultRDFGNo                = 13
	DefaultRemoteRDFGNo          = 13
	RemoteArrayHeaderKey         = "RemoteArray"
	RemoteArrayHeaderValue       = "true"
)

const (
	_ = 1 << (10 * iota)
	// KiB ...
	KiB
	// MiB ...
	MiB
	// GiB ...
	GiB
	// TiB ...
	TiB
	// PiB ...
	PiB
)

var mockCacheMutex sync.Mutex

// Data are internal tables the Mock Unisphere uses to provide functionality.
var Data struct {
	VolumeIDToIdentifier          map[string]string
	VolumeIDToSize                map[string]int
	VolumeIDIteratorList          []string
	VolumeIDToSGList              map[string][]string
	MaskingViewIDToHostID         map[string]string
	MaskingViewIDToHostGroupID    map[string]string
	MaskingViewIDToPortGroupID    map[string]string
	MaskingViewIDToStorageGroupID map[string]string
	StorageGroupIDToMaskingViewID map[string]string
	JobIDToMockJob                map[string]*JobInfo
	StorageGroupIDToNVolumes      map[string]int
	StorageGroupIDToStorageGroup  map[string]*types.StorageGroup
	StorageGroupIDToVolumes       map[string][]string
	MaskingViewIDToMaskingView    map[string]*types.MaskingView
	InitiatorIDToInitiator        map[string]*types.Initiator
	HostIDToHost                  map[string]*types.Host
	PortGroupIDToPortGroup        map[string]*types.PortGroup
	PortIDToSymmetrixPortType     map[string]*types.SymmetrixPortType
	VolumeIDToVolume              map[string]*types.Volume
	HostGroupIDToHostGroup        map[string]*types.HostGroup
	JSONDir                       string
	InitiatorHost                 string

	// Snapshots
	VolIDToSnapshots  map[string]map[string]*types.Snapshot
	SnapIDToLinkedVol map[string]map[string]*types.LinkedVolumes

	// SRDF
	StorageGroupIDToRDFStorageGroup map[string]*types.RDFStorageGroup
	RDFGroup                        *types.RDFGroup
	SGRDFInfo                       *types.SGRDFInfo
}

// InducedErrors constants
var InducedErrors struct {
	NoConnection                       bool
	InvalidJSON                        bool
	BadHTTPStatus                      int
	GetSymmetrixError                  bool
	GetVolumeIteratorError             bool
	GetVolumeError                     bool
	UpdateVolumeError                  bool
	DeleteVolumeError                  bool
	DeviceInSGError                    bool
	GetStorageGroupError               bool
	GetStorageGroupSnapshotPolicyError bool
	InvalidResponse                    bool
	GetStoragePoolError                bool
	UpdateStorageGroupError            bool
	GetJobError                        bool
	JobFailedError                     bool
	VolumeNotCreatedError              bool
	GetJobCannotFindRoleForUser        bool
	CreateStorageGroupError            bool
	StorageGroupAlreadyExists          bool
	DeleteStorageGroupError            bool
	GetStoragePoolListError            bool
	GetPortGroupError                  bool
	GetPortError                       bool
	GetSpecificPortError               bool
	GetPortISCSITargetError            bool
	GetPortGigEError                   bool
	GetDirectorError                   bool
	GetInitiatorError                  bool
	GetInitiatorByIDError              bool
	GetHostError                       bool
	CreateHostError                    bool
	DeleteHostError                    bool
	UpdateHostError                    bool
	GetMaskingViewError                bool
	CreateMaskingViewError             bool
	UpdateMaskingViewError             bool
	MaskingViewAlreadyExists           bool
	DeleteMaskingViewError             bool
	PortGroupNotFoundError             bool
	InitiatorGroupNotFoundError        bool
	StorageGroupNotFoundError          bool
	VolumeNotAddedError                bool
	GetMaskingViewConnectionsError     bool
	ResetAfterFirstError               bool
	CreateSnapshotError                bool
	DeleteSnapshotError                bool
	LinkSnapshotError                  bool
	RenameSnapshotError                bool
	GetSymVolumeError                  bool
	GetVolSnapsError                   bool
	GetGenerationError                 bool
	GetPrivateVolumeIterator           bool
	SnapshotNotLicensed                bool
	UnisphereMismatchError             bool
	TargetNotDefinedError              bool
	SnapshotExpired                    bool
	InvalidSnapshotName                bool
	GetPrivVolumeByIDError             bool
	CreatePortGroupError               bool
	UpdatePortGroupError               bool
	DeletePortGroupError               bool
	ExpandVolumeError                  bool
	MaxSnapSessionError                bool
	GetSRDFInfoError                   bool
	VolumeRdfTypesError                bool
	GetSRDFPairInfoError               bool
	GetProtectedStorageGroupError      bool
	CreateSGReplicaError               bool
	GetRDFGroupError                   bool
	GetSGOnRemote                      bool
	GetSGWithVolOnRemote               bool
	RDFGroupHasPairError               bool
	GetRemoteVolumeError               bool
	InvalidLocalVolumeError            bool
	InvalidRemoteVolumeError           bool
	FetchResponseError                 bool
	RemoveVolumesFromSG                bool
	ModifyMobilityError                bool
	GetHostGroupError                  bool
	CreateHostGroupError               bool
	DeleteHostGroupError               bool
	UpdateHostGroupError               bool
}

// hasError checks to see if the specified error (via pointer)
// is set. If so it returns true, else false.
// Additionally if ResetAfterFirstError is set, the first error
// condition will be reset to no longer be an error condition.
func hasError(errorType *bool) bool {
	if *errorType {
		if InducedErrors.ResetAfterFirstError {
			*errorType = false
			InducedErrors.ResetAfterFirstError = false
		}
		return true
	}
	return false
}

// Reset : re-initializes the variables
func Reset() {
	InducedErrors.NoConnection = false
	InducedErrors.InvalidJSON = false
	InducedErrors.BadHTTPStatus = 0
	InducedErrors.GetSymmetrixError = false
	InducedErrors.GetVolumeIteratorError = false
	InducedErrors.GetVolumeError = false
	InducedErrors.UpdateVolumeError = false
	InducedErrors.DeleteVolumeError = false
	InducedErrors.DeviceInSGError = false
	InducedErrors.GetStorageGroupError = false
	InducedErrors.GetStorageGroupSnapshotPolicyError = false
	InducedErrors.InvalidResponse = false
	InducedErrors.UpdateStorageGroupError = false
	InducedErrors.ModifyMobilityError = false
	InducedErrors.GetJobError = false
	InducedErrors.JobFailedError = false
	InducedErrors.VolumeNotCreatedError = false
	InducedErrors.GetJobCannotFindRoleForUser = false
	InducedErrors.CreateStorageGroupError = false
	InducedErrors.StorageGroupAlreadyExists = false
	InducedErrors.DeleteStorageGroupError = false
	InducedErrors.GetStoragePoolListError = false
	InducedErrors.GetStoragePoolError = false
	InducedErrors.GetPortGroupError = false
	InducedErrors.GetPortError = false
	InducedErrors.GetSpecificPortError = false
	InducedErrors.GetPortISCSITargetError = false
	InducedErrors.GetPortGigEError = false
	InducedErrors.GetDirectorError = false
	InducedErrors.GetInitiatorError = false
	InducedErrors.GetInitiatorByIDError = false
	InducedErrors.GetHostError = false
	InducedErrors.CreateHostError = false
	InducedErrors.DeleteHostError = false
	InducedErrors.UpdateHostError = false
	InducedErrors.GetMaskingViewError = false
	InducedErrors.CreateMaskingViewError = false
	InducedErrors.UpdateMaskingViewError = false
	InducedErrors.MaskingViewAlreadyExists = false
	InducedErrors.DeleteMaskingViewError = false
	InducedErrors.PortGroupNotFoundError = false
	InducedErrors.InitiatorGroupNotFoundError = false
	InducedErrors.StorageGroupNotFoundError = false
	InducedErrors.VolumeNotAddedError = false
	InducedErrors.GetMaskingViewConnectionsError = false
	InducedErrors.ResetAfterFirstError = false
	InducedErrors.CreateSnapshotError = false
	InducedErrors.LinkSnapshotError = false
	InducedErrors.GetSymVolumeError = false
	InducedErrors.GetVolSnapsError = false
	InducedErrors.DeleteSnapshotError = false
	InducedErrors.RenameSnapshotError = false
	InducedErrors.GetGenerationError = false
	InducedErrors.GetPrivateVolumeIterator = false
	InducedErrors.SnapshotNotLicensed = false
	InducedErrors.UnisphereMismatchError = false
	InducedErrors.TargetNotDefinedError = false
	InducedErrors.SnapshotExpired = false
	InducedErrors.InvalidSnapshotName = false
	InducedErrors.GetPrivVolumeByIDError = false
	InducedErrors.CreatePortGroupError = false
	InducedErrors.UpdatePortGroupError = false
	InducedErrors.DeletePortGroupError = false
	InducedErrors.ExpandVolumeError = false
	InducedErrors.MaxSnapSessionError = false
	InducedErrors.GetSRDFInfoError = false
	InducedErrors.VolumeRdfTypesError = false
	InducedErrors.GetSRDFPairInfoError = false
	InducedErrors.GetProtectedStorageGroupError = false
	InducedErrors.CreateSGReplicaError = false
	InducedErrors.GetRDFGroupError = false
	InducedErrors.GetSGOnRemote = false
	InducedErrors.GetSGWithVolOnRemote = false
	InducedErrors.RDFGroupHasPairError = false
	InducedErrors.InvalidLocalVolumeError = false
	InducedErrors.InvalidRemoteVolumeError = false
	InducedErrors.GetRemoteVolumeError = false
	InducedErrors.FetchResponseError = false
	InducedErrors.RemoveVolumesFromSG = false
	InducedErrors.GetHostGroupError = false
	InducedErrors.CreateHostGroupError = false
	InducedErrors.DeleteHostGroupError = false
	InducedErrors.UpdateHostGroupError = false
	Data.JSONDir = "mock"
	Data.VolumeIDToIdentifier = make(map[string]string)
	Data.VolumeIDToSize = make(map[string]int)
	Data.VolumeIDIteratorList = make([]string, 0)
	Data.VolumeIDToSGList = make(map[string][]string)
	Data.MaskingViewIDToHostID = make(map[string]string)
	Data.MaskingViewIDToHostGroupID = make(map[string]string)
	Data.MaskingViewIDToPortGroupID = make(map[string]string)
	Data.MaskingViewIDToStorageGroupID = make(map[string]string)
	Data.StorageGroupIDToMaskingViewID = make(map[string]string)
	Data.JobIDToMockJob = make(map[string]*JobInfo)
	Data.StorageGroupIDToNVolumes = make(map[string]int)
	Data.StorageGroupIDToNVolumes[DefaultStorageGroup] = 0
	Data.StorageGroupIDToStorageGroup = make(map[string]*types.StorageGroup)
	Data.MaskingViewIDToMaskingView = make(map[string]*types.MaskingView)
	Data.InitiatorIDToInitiator = make(map[string]*types.Initiator)
	Data.HostIDToHost = make(map[string]*types.Host)
	Data.PortGroupIDToPortGroup = make(map[string]*types.PortGroup)
	Data.PortIDToSymmetrixPortType = make(map[string]*types.SymmetrixPortType)
	Data.VolumeIDToVolume = make(map[string]*types.Volume)
	Data.StorageGroupIDToVolumes = make(map[string][]string)
	Data.VolIDToSnapshots = make(map[string]map[string]*types.Snapshot)
	Data.SnapIDToLinkedVol = make(map[string]map[string]*types.LinkedVolumes)
	Data.StorageGroupIDToRDFStorageGroup = make(map[string]*types.RDFStorageGroup)
	Data.HostGroupIDToHostGroup = make(map[string]*types.HostGroup)
	Data.RDFGroup = &types.RDFGroup{
		RdfgNumber:          DefaultRDFGNo,
		Label:               "RG_13",
		RemoteRdfgNumber:    DefaultRDFGNo,
		RemoteSymmetrix:     DefaultRemoteSymID,
		NumDevices:          0,
		TotalDeviceCapacity: 0.0,
		Modes:               []string{"Asynchronous"},
		Type:                "VASA_ASYNC",
		Async:               true,
	}
	Data.SGRDFInfo = &types.SGRDFInfo{
		RdfGroupNumber: DefaultRDFGNo,
		VolumeRdfTypes: []string{"R1"},
		States:         []string{"Consistent"},
		Modes:          []string{"Asynchronous"},
		LargerRdfSides: []string{"Equal"},
	}
	initMockCache()
}

func initMockCache() {
	// Initialize SGs
	AddStorageGroup("CSI-Test-SG-1", "SRP_1", "Diamond")       // #nosec G20
	AddStorageGroup("CSI-Test-SG-2", "SRP_1", "Diamond")       // #nosec G20
	AddStorageGroup("CSI-Test-SG-3", "SRP_2", "Silver")        // #nosec G20
	AddStorageGroup("CSI-Test-SG-4", "SRP_2", "Optimized")     // #nosec G20
	AddStorageGroup("CSI-Test-SG-5", "SRP_2", "None")          // #nosec G20
	AddStorageGroup("CSI-Test-SG-6", "None", "None")           // #nosec G20
	AddStorageGroup("CSI-Test-Fake-Remote-SG", "None", "None") // #nosec G20
	// Initialize protected SG
	AddStorageGroup(DefaultProtectedStorageGroup, "None", "None")        // #nosec G20
	AddRDFStorageGroup(DefaultProtectedStorageGroup, DefaultRemoteSymID) // #nosec G20
	// ISCSI directors
	iscsiDir1 := "SE-1E"
	iscsidir1PortKey1 := iscsiDir1 + ":" + "4"
	//iscsiDir2 := "SE-2E"
	// FC directors
	fcDir1 := "FA-1D"
	fcDir2 := "FA-2D"
	fcDir1PortKey1 := fcDir1 + ":" + "5"
	fcDir2PortKey1 := fcDir2 + ":" + "1"
	// Add Port groups
	AddPortGroup("csi-pg", "Fibre", []string{fcDir1PortKey1, fcDir2PortKey1}) // #nosec G20
	// Initialize initiators
	// Initialize Hosts
	initNode1List := make([]string, 0)
	iqnNode1 := "iqn.1993-08.org.centos:01:5ae577b352a0"
	initNode1 := iscsidir1PortKey1 + ":" + iqnNode1
	initNode1List = append(initNode1List, iqnNode1)
	AddInitiator(initNode1, iqnNode1, "GigE", []string{iscsidir1PortKey1}, "") // #nosec G20
	AddHost("CSI-Test-Node-1", "iSCSI", initNode1List)                         // #nosec G20
	initNode2List := make([]string, 0)
	iqn1Node2 := "iqn.1993-08.org.centos:01:5ae577b352a1"
	iqn2Node2 := "iqn.1993-08.org.centos:01:5ae577b352a2"
	init1Node2 := iscsidir1PortKey1 + ":" + iqn1Node2
	init2Node2 := iscsidir1PortKey1 + ":" + iqn2Node2
	initNode2List = append(initNode2List, iqn1Node2)
	initNode2List = append(initNode2List, iqn2Node2)
	AddInitiator(init1Node2, iqn1Node2, "GigE", []string{iscsidir1PortKey1}, "")       // #nosec G20
	AddInitiator(init2Node2, iqn2Node2, "GigE", []string{iscsidir1PortKey1}, "")       // #nosec G20
	AddHost("CSI-Test-Node-2", "iSCSI", initNode2List)                                 // #nosec G20
	AddMaskingView("CSI-Test-MV-1", "CSI-Test-SG-1", "CSI-Test-Node-1", "iscsi_ports") // #nosec G20

	initNode3List := make([]string, 0)
	hba1Node3 := "20000090fa9278dd"
	hba2Node3 := "20000090fa9278dc"
	init1Node3 := fcDir1PortKey1 + ":" + hba1Node3
	init2Node3 := fcDir2PortKey1 + ":" + hba1Node3
	init3Node3 := fcDir1PortKey1 + ":" + hba2Node3
	init4Node3 := fcDir2PortKey1 + ":" + hba2Node3
	AddInitiator(init1Node3, hba1Node3, "Fibre", []string{fcDir1PortKey1}, "") // #nosec G20
	AddInitiator(init2Node3, hba1Node3, "Fibre", []string{fcDir2PortKey1}, "") // #nosec G20
	AddInitiator(init3Node3, hba2Node3, "Fibre", []string{fcDir1PortKey1}, "") // #nosec G20
	AddInitiator(init4Node3, hba2Node3, "Fibre", []string{fcDir2PortKey1}, "") // #nosec G20
	initNode3List = append(initNode3List, hba1Node3)
	initNode3List = append(initNode3List, hba2Node3)
	AddHost("CSI-Test-Node-3-FC", "Fibre", initNode3List) // #nosec G20
	AddTempSnapshots()
}

var mockRouter http.Handler

// GetHandler returns the http handler
func GetHandler() http.Handler {
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if Debug {
				log.Printf("handler called: %s %s", r.Method, r.URL)
			}
			if InducedErrors.InvalidJSON {
				w.Write([]byte(`this is not json`)) // #nosec G20
			} else if InducedErrors.NoConnection {
				writeError(w, "No Connection", http.StatusRequestTimeout)
			} else if InducedErrors.BadHTTPStatus != 0 {
				writeError(w, "Internal Error", InducedErrors.BadHTTPStatus)
			} else {
				if mockRouter != nil {
					mockRouter.ServeHTTP(w, r)
				} else {
					getRouter().ServeHTTP(w, r)
				}
			}
		})
	return handler
}

func getRouter() http.Handler {
	router := mux.NewRouter()
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/host/{id}", handleHost)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/host", handleHost)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/hostgroup/{id}", handleHostGroup)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/hostgroup", handleHostGroup)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/initiator/{id}", handleInitiator)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/initiator", handleInitiator)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/portgroup/{id}", handlePortGroup)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/portgroup", handlePortGroup)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/storagegroup/{id}", handleStorageGroup)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/storagegroup", handleStorageGroup)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/maskingview/{mvID}/connections", handleMaskingViewConnections)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/maskingview/{mvID}", handleMaskingView)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/maskingview", handleMaskingView)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/srp/{id}", handleStorageResourcePool)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/srp", handleStorageResourcePool)
	router.HandleFunc(PREFIXNOVERSION+"/common/Iterator/{iterId}/page", handleIterator)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/volume/{volID}", handleVolume)
	router.HandleFunc(PREFIX+"/sloprovisioning/symmetrix/{symid}/volume", handleVolume)
	router.HandleFunc(PRIVATEPREFIX+"/sloprovisioning/symmetrix/{symid}/volume", handlePrivVolume)
	router.HandleFunc(PREFIX+"/system/symmetrix/{symid}/director/{director}/port/{id}", handlePort)
	router.HandleFunc(PREFIX+"/system/symmetrix/{symid}/director/{director}/port", handlePort)
	router.HandleFunc(PREFIX+"/system/symmetrix/{symid}/director/{id}", handleDirector)
	router.HandleFunc(PREFIX+"/system/symmetrix/{symid}/director", handleDirector)
	router.HandleFunc(PREFIX+"/system/symmetrix/{symid}/job/{jobID}", handleJob)
	router.HandleFunc(PREFIX+"/system/symmetrix/{symid}/job", handleJob)
	router.HandleFunc(PREFIX+"/system/symmetrix/{id}", handleSymmetrix)
	router.HandleFunc(PREFIX+"/system/symmetrix", handleSymmetrix)
	router.HandleFunc(PREFIX+"/system/version", handleVersion)
	router.HandleFunc(PREFIX+"/version", handleVersion)
	router.HandleFunc(PREFIXNOVERSION+"/version", handleVersion)
	router.HandleFunc("/", handleNotFound)

	//Snapshot
	router.HandleFunc(PRIVATEPREFIX+"/replication/symmetrix/{symid}/snapshot/{SnapID}", handleSnapshot)
	router.HandleFunc(PRIVATEPREFIX+"/replication/symmetrix/{symid}/volume", handleSymVolumes)
	router.HandleFunc(PRIVATEPREFIX+"/replication/symmetrix/{symid}/volume/{volID}/snapshot", handleVolSnaps)
	router.HandleFunc(PRIVATEPREFIX+"/replication/symmetrix/{symid}/volume/{volID}/snapshot/{SnapID}", handleVolSnaps)
	router.HandleFunc(PRIVATEPREFIX+"/replication/symmetrix/{symid}/volume/{volID}/snapshot/{SnapID}/generation", handleGenerations)
	router.HandleFunc(PRIVATEPREFIX+"/replication/symmetrix/{symid}/volume/{volID}/snapshot/{SnapID}/generation/{genID}", handleGenerations)
	router.HandleFunc(PREFIX+"/replication/capabilities/symmetrix", handleCapabilities)
	router.HandleFunc(PREFIX+"/replication/symmetrix/{symID}/snapshot_policy/{snapshotPolicyID}/storagegroup/{storageGroupID}", handleStorageGroupSnapshotPolicy)

	// SRDF
	router.HandleFunc(PREFIX+"/replication/symmetrix/{symid}/rdf_group/{rdf_no}", handleRDFGroup)
	router.HandleFunc(PREFIX+"/replication/symmetrix/{symid}/storagegroup/{id}", handleRDFStorageGroup)
	router.HandleFunc(PREFIX+"/replication/symmetrix/{symid}/storagegroup/{id}/rdf_group", handleRDFStorageGroup)
	router.HandleFunc(PREFIX+"/replication/symmetrix/{symid}/storagegroup/{id}/rdf_group/{rdf_no}", handleSGRDFInfo)
	router.HandleFunc(PREFIX+"/replication/symmetrix/{symid}/rdf_group/{rdf_no}/volume/{volume_id}", handleRDFDevicePair)

	mockRouter = router
	return router
}

// NewVolume creates a new mock volume with the specified characteristics.
func NewVolume(volumeID, volumeIdentifier string, size int, sgList []string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	Data.VolumeIDToIdentifier[volumeID] = volumeIdentifier
	fmt.Printf("NewVolume: id %s name %s\n", volumeID, volumeIdentifier)
	Data.VolumeIDToSize[volumeID] = size
	Data.VolumeIDToSGList[volumeID] = sgList
}

// TO be used for the endpoints that don't have handlers yet
func handleTODO(w http.ResponseWriter, r *http.Request) {
	writeError(w, "Endpoint not implemented yet", http.StatusNotImplemented)
}

// GET, POST /univmax/restapi/APIVersion/replication/symmetrix/{symID}/rdf_group/{rdf_no}/volume/{volume_id}
func handleRDFDevicePair(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleRDFDevicePairInfo(w, r)
	case http.MethodPost:
		handleRDFDevicePairCreation(w, r)
	default:
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleRDFDevicePairCreation(w http.ResponseWriter, r *http.Request) {
	// TODO: Update mock cache based on the request payload.
	routeParams := mux.Vars(r)
	rdfPairs := new(types.RDFDevicePairList)
	rdfPairs.RDFDevicePair = []types.RDFDevicePair{
		{
			RemoteVolumeName:     routeParams["volume_id"],
			LocalVolumeName:      routeParams["volume_id"],
			RemoteSymmID:         routeParams["symid"],
			LocalSymmID:          routeParams["symid"],
			LocalRdfGroupNumber:  DefaultRDFGNo,
			RemoteRdfGroupNumber: DefaultRemoteRDFGNo,
		},
	}
	writeJSON(w, rdfPairs)
}

// GET /univmax/restapi/APIVersion/replication/symmetrix/{symID}/rdf_group/{rdf_no}/volume/{volume_id}
func handleRDFDevicePairInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if InducedErrors.GetSRDFPairInfoError {
		writeError(w, "Could not retrieve pair info", http.StatusBadRequest)
		return
	}
	routeParams := mux.Vars(r)
	var volumeConfig string
	if routeParams["symid"] == Data.RDFGroup.RemoteSymmetrix {
		volumeConfig = "RDF2+TDEV"
	} else {
		volumeConfig = "RDF1+TDEV"
	}
	rdfDevicePairInfo := &types.RDFDevicePair{
		LocalRdfGroupNumber:  Data.RDFGroup.RdfgNumber,
		RemoteRdfGroupNumber: Data.RDFGroup.RdfgNumber,
		LocalSymmID:          routeParams["symid"],
		RemoteSymmID:         Data.RDFGroup.RemoteSymmetrix,
		LocalVolumeName:      routeParams["volume_id"],
		RemoteVolumeName:     routeParams["volume_id"],
		VolumeConfig:         volumeConfig,
		RdfMode:              Data.RDFGroup.Modes[0],
		RdfpairState:         "Consistent",
		LargerRdfSide:        "Equal",
	}
	writeJSON(w, rdfDevicePairInfo)
}

// GET /univmax/restapi/APIVersion/replication/symmetrix/{symID}/rdf_group/{rdf_no}
func handleRDFGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if InducedErrors.GetRDFGroupError {
		writeError(w, "the specified RA group does not exist: induced error", http.StatusNotFound)
		return
	}
	routeParams := mux.Vars(r)
	rdfGroupNumber := routeParams["rdf_no"]
	if rdfGroupNumber != fmt.Sprintf("%d", Data.RDFGroup.RdfgNumber) {
		writeError(w, "The specified RA group is not valid", http.StatusNotFound)
	} else {
		if InducedErrors.RDFGroupHasPairError {
			Data.RDFGroup.NumDevices = 1
		}
		writeJSON(w, Data.RDFGroup)
	}
}

// GET /univmax/restapi/APIVersion/replication/symmetrix/{symid}/storagegroup/{id}
// POST /univmax/restapi/APIVersion/replication/symmetrix/{symid}/storagegroup/{id}/rdf_group
func handleRDFStorageGroup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if InducedErrors.GetProtectedStorageGroupError {
			writeError(w, "The requested storage group cannot be found: induced error", http.StatusNotFound)
			return
		}
		if InducedErrors.FetchResponseError {
			writeError(w, "Error fetching response", http.StatusBadRequest)
		}
		handleSGRDFFetch(w, r)
	case http.MethodPost:
		if InducedErrors.CreateSGReplicaError {
			writeError(w, "Failed to create SG replica: induced error", http.StatusNotFound)
			return
		}
		handleSGRDFCreation(w, r)
	default:
		writeError(w, "Method["+r.Method+"] not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSGRDFFetch(w http.ResponseWriter, r *http.Request) {
	routeParams := mux.Vars(r)
	storageGroupID := routeParams["id"]
	symmetrixID := routeParams["symid"]
	var (
		rdfStorageGroup *types.RDFStorageGroup
		ok              bool
	)
	if _, ok = Data.StorageGroupIDToStorageGroup[storageGroupID]; !ok {
		writeError(w, "The requested storage group does not exist", http.StatusNotFound)
		return
	}
	if rdfStorageGroup, ok = Data.StorageGroupIDToRDFStorageGroup[storageGroupID]; !ok {
		rdfStorageGroup = &types.RDFStorageGroup{
			SymmetrixID: symmetrixID,
			Name:        storageGroupID,
			Rdf:         false,
		}
	}
	if InducedErrors.RDFGroupHasPairError {
		rdfStorageGroup.NumDevicesNonGk = 1
	}
	writeJSON(w, rdfStorageGroup)
}

func handleSGRDFCreation(w http.ResponseWriter, r *http.Request) {
	sgsrdf := new(types.CreateSGSRDF)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(sgsrdf); err != nil {
		writeError(w, "invalid json", http.StatusBadRequest)
		return
	}
	routeParams := mux.Vars(r)
	storageGroupName := routeParams["id"]
	symmetrixID := routeParams["symid"]
	if _, err := AddRDFStorageGroup(storageGroupName, symmetrixID); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, volumeID := range Data.StorageGroupIDToVolumes[storageGroupName] {
		if _, ok := Data.VolumeIDToVolume[volumeID]; !ok {
			continue
		}
		volume := Data.VolumeIDToVolume[volumeID]
		volume.Type = "RDF1+TDEV"
		volume.RDFGroupIDList = []types.RDFGroupID{
			{RDFGroupNumber: Data.RDFGroup.RdfgNumber},
		}
	}
	sgrdfInfo := new(types.SGRDFInfo)
	err := copier.Copy(sgrdfInfo, Data.SGRDFInfo)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sgrdfInfo.SymmetrixID = symmetrixID
	sgrdfInfo.StorageGroupName = storageGroupName
	writeJSON(w, sgrdfInfo)
}

// GET, PUT /replication/symmetrix/{symid}/storagegroup/{id}/rdf_group/{rdf_no}
func handleSGRDF(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleSGRDFInfo(w, r)
	case http.MethodPut:
		handleSGRDFAction(w, r)
	default:
		writeError(w, "Method["+r.Method+"] not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSGRDFInfo(w http.ResponseWriter, r *http.Request) {
	if InducedErrors.GetSRDFInfoError {
		writeError(w, "Error retrieving SRDF Info(%s): induced error", http.StatusRequestTimeout)
		return
	}
	routeParams := mux.Vars(r)
	if routeParams["rdf_no"] != fmt.Sprintf("%d", Data.RDFGroup.RdfgNumber) {
		writeError(w, "The specified RA group is not valid", http.StatusNotFound)
	} else {
		sgrdfInfo := new(types.SGRDFInfo)
		err := copier.Copy(sgrdfInfo, Data.SGRDFInfo)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if InducedErrors.VolumeRdfTypesError {
			sgrdfInfo.VolumeRdfTypes = []string{"invalid"}
		}
		sgrdfInfo.SymmetrixID = routeParams["symid"]
		sgrdfInfo.StorageGroupName = routeParams["id"]
		writeJSON(w, sgrdfInfo)
	}
}

func handleSGRDFAction(w http.ResponseWriter, r *http.Request) {
	// TODO: execute actions by updating the memory cache
	w.WriteHeader(200)
}

// GET /univmax/restapi/system/version
func handleVersion(w http.ResponseWriter, r *http.Request) {
	auth := defaultUsername + ":" + defaultPassword
	authExpected := fmt.Sprintf("Basic " + base64.StdEncoding.EncodeToString([]byte(auth)))
	// Check for valid credentials
	authSupplied := r.Header.Get("Authorization")
	if authExpected != authSupplied {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	apiversion := vars["apiversion"]
	// check the apiversion
	switch apiversion {
	case "90":
		w.Write([]byte(`{ "version": "V9.0.1.6" }`)) // #nosec G20
		break
	case "": // for version 91, as URL does not have apiversion in V9.1
		w.Write([]byte(`{ "version": "V9.1.0.2" }`)) // #nosec G20
		break
	default:
		writeError(w, "Unsupport API version: "+apiversion, http.StatusServiceUnavailable)
	}
}

// GET /univmax/restapi/APIVersion/system/symmetrix/{id}"
// GET /univmax/restapi/APIVersion/system/symmetrix"
func handleSymmetrix(w http.ResponseWriter, r *http.Request) {
	if InducedErrors.GetSymmetrixError {
		writeError(w, "Error retrieving Symmetrix: induced error", http.StatusRequestTimeout)
		return
	}
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		returnJSONFile(Data.JSONDir, "symmetrixList.json", w, nil)
	}
	if id != "000197900046" && id != "000197900047" {
		writeError(w, "Symmetrix not found", http.StatusNotFound)
		return
	}
	if id == "000197900046" {
		returnJSONFile(Data.JSONDir, "symmetrix46.json", w, nil)
	} else if id == "000197900047" {
		returnJSONFile(Data.JSONDir, "symmetrix47.json", w, nil)
	}
}

func handleStorageResourcePool(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	srpID := vars["id"]
	if InducedErrors.GetStoragePoolListError {
		writeError(w, "Error retrieving StoragePools: induced error", http.StatusRequestTimeout)
		return
	}
	if InducedErrors.GetStoragePoolError {
		writeError(w, "Error retrieving Storage Pool(s): induced error", http.StatusRequestTimeout)
		return
	}
	if srpID == "" {
		returnJSONFile(Data.JSONDir, "storageResourcePool.json", w, nil)
	}
	replacements := make(map[string]string)
	replacements["__SRP_ID__"] = "SRP_1"
	returnJSONFile(Data.JSONDir, "storage_pool_template.json", w, replacements)
}

// GET /univmax/restapi/API_VERSON/sloprovisioning/symmetrix/{id}/volume/{id}
// GET /univmax/restapi/API_VERSON/sloprovisioning/symmetrix/{id}/volume
func handleVolume(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volID := vars["volID"]
	switch r.Method {
	case http.MethodGet:
		mockCacheMutex.Lock()
		defer mockCacheMutex.Unlock()
		if volID == "" {
			if InducedErrors.GetVolumeIteratorError {
				writeError(w, "Error getting VolumeIterator: induced error", http.StatusRequestTimeout)
				return
			}
			// Here we want a volume iterator.
			var like bool
			queryParams := r.URL.Query()
			volumeIdentifier := queryParams.Get("volume_identifier")
			if strings.Contains(volumeIdentifier, "<like>") {
				like = true
				volumeIdentifier = strings.TrimPrefix(volumeIdentifier, "<like>")
			}
			// Copy data to Data.VolumeIDIteratorList, while checking for volumeIdentifier match if needed
			Data.VolumeIDIteratorList = make([]string, 0)
			for _, vol := range Data.VolumeIDToVolume {
				if volumeIdentifier != "" {
					if like {
						if !strings.Contains(vol.VolumeIdentifier, volumeIdentifier) {
							continue
						}
					} else {
						if vol.VolumeIdentifier != volumeIdentifier {
							continue
						}
					}
				}
				Data.VolumeIDIteratorList = append(Data.VolumeIDIteratorList, vol.VolumeID)
			}
			if Debug {
				fmt.Printf("Data.VolumeIDIteratorList %#v", Data.VolumeIDIteratorList)
			}
			iter := &types.VolumeIterator{
				Count:          len(Data.VolumeIDIteratorList),
				ID:             "Volume",
				MaxPageSize:    10,
				ExpirationTime: 0,
			}
			numberToDo := len(Data.VolumeIDIteratorList)
			if numberToDo > iter.MaxPageSize {
				numberToDo = iter.MaxPageSize
			}
			iter.ResultList.From = 1
			iter.ResultList.To = numberToDo
			for i := iter.ResultList.From - 1; i <= iter.ResultList.To-1; i++ {
				volIDList := types.VolumeIDList{VolumeIDs: Data.VolumeIDIteratorList[i]}
				iter.ResultList.VolumeList = append(iter.ResultList.VolumeList, volIDList)
			}
			if Debug {
				fmt.Printf("iter: %#v\n", iter)
			}
			encoder := json.NewEncoder(w)
			err := encoder.Encode(iter)
			if err != nil {
				writeError(w, "json encoding error", http.StatusInternalServerError)
			}
			return
		}
		if InducedErrors.GetVolumeError {
			writeError(w, "Error retrieving Volume: induced error", http.StatusRequestTimeout)
			return
		}
		if volID != "" {
			if vars["symid"] == Data.RDFGroup.RemoteSymmetrix {
				returnVolume(w, volID, true)
			} else {
				returnVolume(w, volID, false)
			}
		}

	case http.MethodPut:
		if InducedErrors.UpdateVolumeError {
			writeError(w, "Error updating Volume: induced error", http.StatusRequestTimeout)
			return
		}
		if volID == "" {
			writeError(w, "Volume ID must be supplied", http.StatusBadRequest)
			return
		}
		decoder := json.NewDecoder(r.Body)
		updateVolumePayload := &types.EditVolumeParam{}
		err := decoder.Decode(updateVolumePayload)
		if err != nil {
			writeError(w, "problem decoding PUT Volume payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("PUT volume payload: %#v\n", updateVolumePayload)
		executionOption := updateVolumePayload.ExecutionOption
		if updateVolumePayload.EditVolumeActionParam.FreeVolumeParam != nil {
			FreeVolume(w, updateVolumePayload.EditVolumeActionParam.FreeVolumeParam, volID, executionOption)
			return
		}
		if updateVolumePayload.EditVolumeActionParam.EnableMobilityIDParam != nil {
			ModifyMobility(w, updateVolumePayload.EditVolumeActionParam.EnableMobilityIDParam, volID, executionOption)
			return
		}
		if updateVolumePayload.EditVolumeActionParam.ModifyVolumeIdentifierParam != nil {
			if vars["symid"] == Data.RDFGroup.RemoteSymmetrix {
				RenameVolume(w, updateVolumePayload.EditVolumeActionParam.ModifyVolumeIdentifierParam, volID, executionOption, true)
			} else {
				RenameVolume(w, updateVolumePayload.EditVolumeActionParam.ModifyVolumeIdentifierParam, volID, executionOption, false)
			}
			return
		}
		if updateVolumePayload.EditVolumeActionParam.ExpandVolumeParam != nil {
			ExpandVolume(w, updateVolumePayload.EditVolumeActionParam.ExpandVolumeParam, volID, executionOption)
			return
		}
	case http.MethodDelete:
		if InducedErrors.DeleteVolumeError {
			writeError(w, "Error deleting Volume: induced error", http.StatusRequestTimeout)
			return
		}
		if InducedErrors.DeviceInSGError {
			writeError(w, "Error deleting Volume: induced error - device is a member of a storage group", http.StatusForbidden)
			return
		}
		DeleteVolume(volID) // #nosec G20
	}
}

// DeleteVolume - Deletes volume from cache
func DeleteVolume(volID string) error {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	return deleteVolume(volID)
}

func deleteVolume(volID string) error {
	vol, ok := Data.VolumeIDToVolume[volID]
	if ok {
		if vol.NumberOfStorageGroups > 0 {
			return errors.New("Volume present in storage group. Can't be deleted")
		}
		Data.VolumeIDToVolume[volID] = nil
	} else {
		return errors.New("Volume not found")
	}
	return nil
}

func returnVolume(w http.ResponseWriter, volID string, remote bool) {
	if volID != "" {
		if vol, ok := Data.VolumeIDToVolume[volID]; ok {
			newVol := new(types.Volume)
			err := copier.Copy(newVol, vol)
			if err != nil {
				writeError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Printf("volume: %#v\n", vol)
			if InducedErrors.InvalidLocalVolumeError {
				newVol.StorageGroupIDList = nil
			}
			if remote {
				if InducedErrors.FetchResponseError {
					writeError(w, "Error fetching response", http.StatusBadRequest)
					return
				}
				if InducedErrors.GetRemoteVolumeError {
					writeError(w, "Volume cannot be found", http.StatusNotFound)
					return
				}
				if InducedErrors.InvalidRemoteVolumeError {
					newVol.StorageGroupIDList = nil
				}
				if !strings.Contains(vol.Type, "RDF") {
					writeError(w, "Volume not found", http.StatusNotFound)
					return
				}
				newVol.Type = strings.ReplaceAll(newVol.Type, "RDF1", "RDF2")
				newVol.VolumeIdentifier = ""
			}
			writeJSON(w, newVol)
			return
		}
		writeError(w, "Volume cannot be found: "+volID, http.StatusNotFound)
	}
}

// FreeVolume - handler for free volume job
func FreeVolume(w http.ResponseWriter, param *types.FreeVolumeParam, volID string, executionOption string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	freeVolume(w, param, volID, executionOption)
}

// This returns a job for freeing space in a volume
func freeVolume(w http.ResponseWriter, param *types.FreeVolumeParam, volID string, executionOption string) {
	if executionOption != types.ExecutionOptionAsynchronous {
		writeError(w, "expected ASYNCHRONOUS", http.StatusBadRequest)
		return
	}
	// Make a job to return
	resourceLink := fmt.Sprintf("sloprovisioning/system/%s/volume/%s", DefaultSymmetrixID, volID)
	if InducedErrors.JobFailedError {
		newMockJob(volID, types.JobStatusRunning, types.JobStatusFailed, resourceLink)
	} else {
		newMockJob(volID, types.JobStatusRunning, types.JobStatusSucceeded, resourceLink)
	}
	returnJobByID(w, volID)
}

// RenameVolume - renames volume in cache
func RenameVolume(w http.ResponseWriter, param *types.ModifyVolumeIdentifierParam, volID string, executionOption string, remote bool) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	renameVolume(w, param, volID, executionOption, remote)
}

// This returns the volume itself after renaming
func renameVolume(w http.ResponseWriter, param *types.ModifyVolumeIdentifierParam, volID string, executionOption string, remote bool) {
	if executionOption != types.ExecutionOptionSynchronous {
		writeError(w, "expected SYNCHRONOUS", http.StatusBadRequest)
		return
	}
	Data.VolumeIDToVolume[volID].VolumeIdentifier = param.VolumeIdentifier.IdentifierName
	returnVolume(w, volID, remote)
}

// ModifyMobility modifes the mobility-id-enabled parameter of volume
func ModifyMobility(w http.ResponseWriter, param *types.EnableMobilityIDParam, volID string, executionOption string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	modifyMobility(w, param, volID, executionOption)
}

func modifyMobility(w http.ResponseWriter, param *types.EnableMobilityIDParam, volID string, executionOption string) {
	if InducedErrors.ModifyMobilityError {
		writeError(w, "Error modifying mobility for volume: induced error", http.StatusRequestTimeout)
		return
	}
	Data.VolumeIDToVolume[volID].MobilityIDEnabled = param.EnableMobilityID
	returnVolume(w, volID, false)
}

// ExpandVolume - Expands volume size in cache
func ExpandVolume(w http.ResponseWriter, param *types.ExpandVolumeParam, volID string, executionOption string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	expandVolume(w, param, volID, executionOption)
}

// This returns the volume itself after expanding the volume's size
func expandVolume(w http.ResponseWriter, param *types.ExpandVolumeParam, volID string, executionOption string) {
	if InducedErrors.ExpandVolumeError {
		writeError(w, "Error expanding volume: induced error", http.StatusRequestTimeout)
		return
	}
	if executionOption != types.ExecutionOptionSynchronous {
		writeError(w, "expected SYNCHRONOUS", http.StatusBadRequest)
		return
	}

	newSize, err := strconv.ParseFloat(param.VolumeAttribute.VolumeSize, 64)
	switch param.VolumeAttribute.CapacityUnit {
	case "MB":
		newSize = newSize * MiB / GiB
	case "TB":
		newSize = newSize * TiB / GiB
	case "PB":
		newSize = newSize * PiB / GiB
	case "GB":
	}

	if err == nil {
		Data.VolumeIDToVolume[volID].CapacityGB = newSize
	} else {
		writeError(w, fmt.Sprintf("Could not convert expand size parameter in request (%s)", param.VolumeAttribute.VolumeSize), http.StatusBadRequest)
		return
	}
	returnVolume(w, volID, false)
}

// JobInfo is used to simulate a job in Unisphere.
// The first call to read it returns Status as the InitialState.
// Subsequent calls return the Status as the FinalState.
type JobInfo struct {
	Job          types.Job
	InitialState string
	FinalState   string
}

// NewMockJob creates a JobInfo that can be queried
func NewMockJob(jobID string, initialState string, finalState string, resourceLink string) *JobInfo {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	return newMockJob(jobID, initialState, finalState, resourceLink)
}

func newMockJob(jobID string, initialState string, finalState string, resourceLink string) *JobInfo {
	job := new(JobInfo)
	job.Job.JobID = jobID
	job.InitialState = initialState
	job.FinalState = finalState
	job.Job.Status = "SCHEDULED"
	job.Job.ResourceLink = resourceLink
	Data.JobIDToMockJob[jobID] = job
	return job
}

func handleJob(w http.ResponseWriter, r *http.Request) {
	if InducedErrors.GetJobError {
		writeError(w, "Error getting Job(s): induced error", http.StatusRequestTimeout)
		return
	}
	vars := mux.Vars(r)
	jobID := vars["jobID"]
	if jobID == "" {
		queryParams := r.URL.Query()
		// Return a job id list
		jobIDList := new(types.JobIDList)
		jobIDList.JobIDs = make([]string, 0)
		mockCacheMutex.Lock()
		defer mockCacheMutex.Unlock()
		for key := range Data.JobIDToMockJob {
			job := Data.JobIDToMockJob[key].Job
			if queryParams.Get("status") == "" || queryParams.Get("status") == job.Status {
				jobIDList.JobIDs = append(jobIDList.JobIDs, key)
			}
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(jobIDList) // #nosec G20
		return
	}
	// Return a specific job
	if InducedErrors.GetJobCannotFindRoleForUser {
		InducedErrors.GetJobCannotFindRoleForUser = false
		writeError(w, "Cannot find role for user", http.StatusInternalServerError)
		return
	}
	ReturnJobByID(w, jobID)
}

// ReturnJobByID - Returns job based on ID from mock cache
func ReturnJobByID(w http.ResponseWriter, jobID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	returnJobByID(w, jobID)
}

func returnJobByID(w http.ResponseWriter, jobID string) {
	job := Data.JobIDToMockJob[jobID]
	if job == nil {
		// Not found
		writeError(w, "Job not found: "+jobID, http.StatusNotFound)
		return
	}
	if job.Job.Status == job.InitialState {
		job.Job.Status = job.FinalState
		job.Job.CompletedDate = time.Now().String()
		job.Job.Result = "Mock job completed"
	} else {
		job.Job.Status = job.InitialState
		job.Job.Result = "Mock job in-progress"
	}
	encoder := json.NewEncoder(w)
	err := encoder.Encode(&job.Job)
	if err != nil {
		writeError(w, "json encoding error", http.StatusInternalServerError)
	}
}

// /unixvmax/restapi/common/Iterator/{iterID]/page}
func handleIterator(w http.ResponseWriter, r *http.Request) {
	var err error
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	switch r.Method {
	case http.MethodGet:
		vars := mux.Vars(r)
		queryParams := r.URL.Query()
		from := queryParams.Get("from")
		to := queryParams.Get("to")
		fmt.Printf("mux iterId %s from %s to %s\n", vars["iterId"], from, to)
		result := &types.VolumeResultList{}
		result.From, err = strconv.Atoi(from)
		if err != nil {
			writeError(w, "bad from query parameter", http.StatusBadRequest)
		}
		result.To, err = strconv.Atoi(to)
		if err != nil {
			writeError(w, "bad from query parameter", http.StatusBadRequest)
		}
		for i := result.From - 1; i < result.To-1; i++ {
			volIDList := types.VolumeIDList{VolumeIDs: Data.VolumeIDIteratorList[i]}
			result.VolumeList = append(result.VolumeList, volIDList)
		}
		if Debug {
			fmt.Printf("volumeResultList: %#v\n", result)
		}
		encoder := json.NewEncoder(w)
		err := encoder.Encode(result)
		if err != nil {
			writeError(w, "volumeResultList json encoding error", http.StatusInternalServerError)
		}
	case http.MethodDelete:
		// Nothing to do, will return
	}
}

func handleStorageGroupSnapshotPolicy(w http.ResponseWriter, r *http.Request) {
	if InducedErrors.GetStorageGroupSnapshotPolicyError {
		writeError(w, "Error retrieving storage group snapshot policy: induced error", http.StatusRequestTimeout)
		return
	}
	returnJSONFile(Data.JSONDir, "storage_group_snapshot_policy.json", w, nil)
}

// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/storagegroup/{id}
// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/storagegroup
// /univmax/restapi/91/sloprovisioning/symmetrix/{symid}/storagegroup/{id}
// /univmax/restapi/91/sloprovisioning/symmetrix/{symid}/storagegroup
func handleStorageGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sgID := vars["id"]
	apiversion := vars["apiversion"]
	switch r.Method {
	case http.MethodGet:
		if InducedErrors.GetStorageGroupError {
			writeError(w, "Error retrieving Storage Group(s): induced error", http.StatusRequestTimeout)
			return
		}
		if vars["symid"] == Data.RDFGroup.RemoteSymmetrix && strings.Contains(sgID, "rep") {
			ReturnStorageGroup(w, sgID, true)
		} else {
			ReturnStorageGroup(w, sgID, false)
		}

	case http.MethodPut:
		if InducedErrors.UpdateStorageGroupError {
			writeError(w, "Error updating Storage Group: induced error", http.StatusRequestTimeout)
			return
		}
		if sgID == "" {
			writeError(w, "storage group ID must be supplied", http.StatusBadRequest)
			return
		}
		decoder := json.NewDecoder(r.Body)
		if apiversion == "90" {
			updateSGPayload := &types.UpdateStorageGroupPayload{}
			err := decoder.Decode(updateSGPayload)
			if err != nil {
				writeError(w, "problem decoding PUT StorageGroup payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			fmt.Printf("PUT StorageGroup payload: %#v\n", updateSGPayload)
			editPayload := updateSGPayload.EditStorageGroupActionParam
			if editPayload.ExpandStorageGroupParam != nil {
				expandPayload := editPayload.ExpandStorageGroupParam
				addVolumeParam := expandPayload.AddVolumeParam
				if addVolumeParam != nil {
					name := addVolumeParam.VolumeIdentifier.IdentifierName
					var size string
					if len(addVolumeParam.VolumeAttributes) > 0 {
						size = addVolumeParam.VolumeAttributes[0].VolumeSize
					}
					size = addVolumeParam.VolumeAttributes[0].VolumeSize
					AddVolumeToStorageGroupTest(w, name, size, sgID)
				}
				addSpecificVolumeParam := expandPayload.AddSpecificVolumeParam
				if addSpecificVolumeParam != nil {
					AddSpecificVolumeToStorageGroup(w, addSpecificVolumeParam.VolumeIDs, sgID)
				}
			}
			if editPayload.RemoveVolumeParam != nil {
				RemoveVolumeFromStorageGroup(w, editPayload.RemoveVolumeParam.VolumeIDs, sgID)

			}
		} else {
			// for apiVersion 91
			updateSGPayload := &types.UpdateStorageGroupPayload{}
			err := decoder.Decode(updateSGPayload)
			if err != nil {
				writeError(w, "problem decoding PUT StorageGroup payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			fmt.Printf("PUT StorageGroup payload: %#v\n", updateSGPayload)
			editPayload := updateSGPayload.EditStorageGroupActionParam
			if editPayload.ExpandStorageGroupParam != nil {
				expandPayload := editPayload.ExpandStorageGroupParam
				addVolumeParam := expandPayload.AddVolumeParam
				if addVolumeParam != nil {
					name := addVolumeParam.VolumeAttributes[0].VolumeIdentifier.IdentifierName
					size := addVolumeParam.VolumeAttributes[0].VolumeSize
					AddVolumeToStorageGroupTest(w, name, size, sgID)
				}
				addSpecificVolumeParam := expandPayload.AddSpecificVolumeParam
				if addSpecificVolumeParam != nil {
					AddSpecificVolumeToStorageGroup(w, addSpecificVolumeParam.VolumeIDs, sgID)
				}
			}
			if editPayload.RemoveVolumeParam != nil {
				RemoveVolumeFromStorageGroup(w, editPayload.RemoveVolumeParam.VolumeIDs, sgID)
			}
		}
	case http.MethodPost:
		if InducedErrors.CreateStorageGroupError {
			writeError(w, "Error creating Storage Group: induced error", http.StatusRequestTimeout)
			return
		}
		if InducedErrors.StorageGroupAlreadyExists {
			writeError(w, "The requested storage group resource already exists", http.StatusConflict)
			return
		}
		decoder := json.NewDecoder(r.Body)
		createSGPayload := &types.CreateStorageGroupParam{}
		err := decoder.Decode(createSGPayload)
		if err != nil {
			writeError(w, "problem decoding POST StorageGroup payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("POST StorageGroup payload: %#v\n", createSGPayload)
		sgID := createSGPayload.StorageGroupID
		// Data.StorageGroupIDToNVolumes[sgID] = 0
		// fmt.Println("SG Name: ", sgID)
		AddStorageGroupFromCreateParams(createSGPayload)
		if vars["symid"] == Data.RDFGroup.RemoteSymmetrix {
			ReturnStorageGroup(w, sgID, true)
		} else {
			ReturnStorageGroup(w, sgID, false)
		}

	case http.MethodDelete:
		if InducedErrors.DeleteStorageGroupError {
			writeError(w, "Error deleting storage group: induced error", http.StatusRequestTimeout)
			return
		}
		RemoveStorageGroup(w, sgID)

	default:
		writeError(w, "Invalid Method", http.StatusBadRequest)
	}
}

// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/maskingview/{id}/connections
func handleMaskingViewConnections(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		queryParams := r.URL.Query()
		volID := queryParams.Get("volume_id")
		if InducedErrors.GetMaskingViewConnectionsError {
			writeError(w, "Error retrieving Masking View Connections: induced error", http.StatusRequestTimeout)
			return
		}
		if volID == "" {
			// Return a response for all volumes
			index := 1
			result := &types.MaskingViewConnectionsResult{
				MaskingViewConnections: make([]*types.MaskingViewConnection, 0),
			}
			for id := range Data.VolumeIDToVolume {
				conn1 := &types.MaskingViewConnection{
					VolumeID:       id,
					HostLUNAddress: fmt.Sprintf("%4d", index),
					CapacityGB:     "0.1",
					InitiatorID:    "iqn.1993-08.org.debian:01:8f21cc8ad2a7",
					DirectorPort:   "SE-1E:000",
					LoggedIn:       false,
					OnFabric:       true,
				}
				result.MaskingViewConnections = append(result.MaskingViewConnections, conn1)
				conn2 := &types.MaskingViewConnection{
					VolumeID:       id,
					HostLUNAddress: fmt.Sprintf("%4d", index),
					CapacityGB:     "0.1",
					InitiatorID:    "iqn.1993-08.org.debian:01:8f21cc8ad2a7",
					DirectorPort:   "SE-2E:000",
					LoggedIn:       false,
					OnFabric:       true,
				}
				result.MaskingViewConnections = append(result.MaskingViewConnections, conn2)
				index++
			}
			writeJSON(w, result)
			return
		}
		replacements := make(map[string]string)
		replacements["__VOLUME_ID__"] = volID
		returnJSONFile(Data.JSONDir, "masking_view_connections_template.json", w, replacements)
	}
}

// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/maskingview/{id}
// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/maskingview
func handleMaskingView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mvID := vars["mvID"]
	switch r.Method {
	case http.MethodGet:
		if InducedErrors.GetMaskingViewError {
			writeError(w, "Error retrieving Masking View(s): induced error", http.StatusRequestTimeout)
			return
		}
		returnMaskingView(w, mvID)

	case http.MethodPost:
		if InducedErrors.CreateMaskingViewError {
			writeError(w, "Failed to create masking view: induced error", http.StatusRequestTimeout)
			return
		} else if InducedErrors.MaskingViewAlreadyExists {
			writeError(w, "The requested masking view resource already exists", http.StatusConflict)
			return
		} else if InducedErrors.PortGroupNotFoundError {
			writeError(w, "Port Group on Symmetrix cannot be found", http.StatusInternalServerError)
		} else if InducedErrors.InitiatorGroupNotFoundError {
			writeError(w, "Initiator Group on Symmetrix cannot be found", http.StatusInternalServerError)
		} else if InducedErrors.StorageGroupNotFoundError {
			writeError(w, "Storage Group on Symmetrix cannot be found", http.StatusInternalServerError)
		}
		decoder := json.NewDecoder(r.Body)
		createMVPayload := &types.MaskingViewCreateParam{}
		err := decoder.Decode(createMVPayload)
		if err != nil {
			writeError(w, "problem decoding POST Masking View payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("POST MaskingView payload: %#v\n", createMVPayload)
		mvID := createMVPayload.MaskingViewID
		//Data.StorageGroupIDToNVolumes[sgID] = 0
		fmt.Println("MV Name: ", mvID)
		addMaskingViewFromCreateParams(createMVPayload)
		returnMaskingView(w, mvID)

	case http.MethodPut:
		if InducedErrors.UpdateMaskingViewError {
			writeError(w, "Error updating Masking View: induced error", http.StatusRequestTimeout)
			return
		}
		// if mvID == "" {
		// 	writeError(w, "Masking View ID must be supplied", http.StatusBadRequest)
		// 	return
		// }
		decoder := json.NewDecoder(r.Body)
		updateMaskingViewPayload := &types.EditMaskingViewParam{}
		err := decoder.Decode(updateMaskingViewPayload)
		if err != nil {
			writeError(w, "problem decoding PUT Masking View payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("PUT masking view payload: %#v\n", updateMaskingViewPayload)
		executionOption := updateMaskingViewPayload.ExecutionOption
		if &updateMaskingViewPayload.EditMaskingViewActionParam.RenameMaskingViewParam != nil {
			RenameMaskingView(w, &updateMaskingViewPayload.EditMaskingViewActionParam.RenameMaskingViewParam, mvID, executionOption)
			return
		}

	case http.MethodDelete:
		if InducedErrors.DeleteMaskingViewError {
			writeError(w, "Error deleting Masking view: induced error", http.StatusRequestTimeout)
			return
		}
		RemoveMaskingView(w, mvID)

	default:
		writeError(w, "Invalid Method", http.StatusBadRequest)
	}
}

func newStorageGroup(storageGroupID string, maskingViewID string, storageResourcePoolID string,
	serviceLevel string, numOfVolumes int) {
	numOfMaskingViews := 0
	if maskingViewID != "" {
		numOfMaskingViews = 1
	}
	childStorageGroups := []string{}
	maskingViews := make([]string, 0)
	if maskingViewID != "" {
		maskingViews = append(maskingViews, maskingViewID)
	}
	storageGroup := &types.StorageGroup{
		StorageGroupID:    storageGroupID,
		SLO:               serviceLevel,
		SRP:               storageResourcePoolID,
		Workload:          "None",
		SLOCompliance:     "STABLE",
		NumOfVolumes:      numOfVolumes,
		NumOfChildSGs:     0,
		NumOfParentSGs:    0,
		NumOfMaskingViews: numOfMaskingViews,
		NumOfSnapshots:    0,
		CapacityGB:        234.5,
		DeviceEmulation:   "FBA",
		Type:              "Standalone",
		Unprotected:       true,
		ChildStorageGroup: childStorageGroups,
		MaskingView:       maskingViews,
	}
	Data.StorageGroupIDToStorageGroup[storageGroupID] = storageGroup
	volumes := make([]string, 0)
	Data.StorageGroupIDToVolumes[storageGroupID] = volumes
}

func newMaskingView(maskingViewID string, storageGroupID string, hostID string, portGroupID string) {
	maskingView := &types.MaskingView{
		MaskingViewID:  maskingViewID,
		HostID:         hostID,
		HostGroupID:    "",
		PortGroupID:    portGroupID,
		StorageGroupID: storageGroupID,
	}
	Data.MaskingViewIDToMaskingView[maskingViewID] = maskingView
}

// AddStorageGroup - Adds a storage group to the mock data cache
func AddStorageGroup(storageGroupID string, storageResourcePoolID string,
	serviceLevel string) (*types.StorageGroup, error) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	if _, ok := Data.StorageGroupIDToStorageGroup[storageGroupID]; ok {
		return nil, errors.New("The requested storage group resource already exists")
	}
	newStorageGroup(storageGroupID, "", storageResourcePoolID, serviceLevel, 0)
	return Data.StorageGroupIDToStorageGroup[storageGroupID], nil
}

// AddRDFStorageGroup ...
func AddRDFStorageGroup(storageGroupID, symmetrixID string) (*types.RDFStorageGroup, error) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	if _, ok := Data.StorageGroupIDToRDFStorageGroup[storageGroupID]; ok {
		return nil, fmt.Errorf("rdfStorageGroup already exists")
	}
	Data.StorageGroupIDToStorageGroup[storageGroupID].Unprotected = false
	rdfSG := &types.RDFStorageGroup{
		Name:        storageGroupID,
		SymmetrixID: symmetrixID,
		Rdf:         true,
	}
	Data.StorageGroupIDToRDFStorageGroup[storageGroupID] = rdfSG
	return rdfSG, nil
}

// RemoveStorageGroup - Removes a storage group from the mock data cache
func RemoveStorageGroup(w http.ResponseWriter, storageGroupID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	removeStorageGroup(w, storageGroupID)
}

func removeStorageGroup(w http.ResponseWriter, storageGroupID string) {
	if InducedErrors.GetSGOnRemote {
		storageGroupID = "CSI-Test-Fake-Remote-SG"
	}
	sg, ok := Data.StorageGroupIDToStorageGroup[storageGroupID]
	if !ok {
		fmt.Println("Storage Group " + storageGroupID + " doesn't exist")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if sg.NumOfMaskingViews != 0 {
		fmt.Println("Can't delete a storage group which is part of masking view")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	volumes := Data.StorageGroupIDToVolumes[storageGroupID]
	if InducedErrors.RemoveVolumesFromSG {
		volumes = nil
	}
	if len(volumes) > 0 {
		fmt.Println("Can't delete a storage group which has volumes")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	delete(Data.StorageGroupIDToStorageGroup, storageGroupID)
	delete(Data.StorageGroupIDToStorageGroup, storageGroupID+"-remote")
	delete(Data.StorageGroupIDToRDFStorageGroup, storageGroupID)
	delete(Data.StorageGroupIDToRDFStorageGroup, storageGroupID+"-remote")
}

func addMaskingViewFromCreateParams(createParams *types.MaskingViewCreateParam) {
	mvID := createParams.MaskingViewID
	hostID := ""
	hostGroupID := ""
	if createParams.HostOrHostGroupSelection.UseExistingHostParam != nil {
		hostID = createParams.HostOrHostGroupSelection.UseExistingHostParam.HostID
	} else if createParams.HostOrHostGroupSelection.UseExistingHostGroupParam != nil {
		hostGroupID = createParams.HostOrHostGroupSelection.UseExistingHostGroupParam.HostGroupID
	}
	portGroupID := createParams.PortGroupSelection.UseExistingPortGroupParam.PortGroupID
	sgID := createParams.StorageGroupSelection.UseExistingStorageGroupParam.StorageGroupID
	if hostID != "" {
		AddMaskingView(mvID, sgID, hostID, portGroupID) // #nosec G20
	} else if hostGroupID != "" {
		AddMaskingView(mvID, sgID, hostGroupID, portGroupID) // #nosec G20
	}
}

// AddMaskingView - Adds a masking view to the mock data cache
func AddMaskingView(maskingViewID string, storageGroupID string, hostID string, portGroupID string) (*types.MaskingView, error) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	return addMaskingView(maskingViewID, storageGroupID, hostID, portGroupID)
}

func addMaskingView(maskingViewID string, storageGroupID string, hostID string, portGroupID string) (*types.MaskingView, error) {
	if _, ok := Data.MaskingViewIDToMaskingView[maskingViewID]; ok {
		return nil, errors.New("Error! Masking View already exists")
	}
	if _, ok := Data.StorageGroupIDToStorageGroup[storageGroupID]; !ok {
		return nil, errors.New("Storage Group doesn't exist")
	}
	/*if _, ok := Data.PortGroupIDToPortGroup[portGroupID]; !ok {
		return errors.New("Port Group doesn't exist")
	}*/
	if _, ok := Data.HostIDToHost[hostID]; !ok {
		return nil, errors.New("Host doesn't exist")
	}
	newMaskingView(maskingViewID, storageGroupID, hostID, portGroupID)
	// Update host
	Data.HostIDToHost[hostID].MaskingviewIDs = append(Data.HostIDToHost[hostID].MaskingviewIDs, maskingViewID)
	Data.HostIDToHost[hostID].NumberMaskingViews++
	// Update Storage Group
	currentMaskingViewIDs := Data.StorageGroupIDToStorageGroup[storageGroupID].MaskingView
	Data.StorageGroupIDToStorageGroup[storageGroupID].MaskingView = append(
		currentMaskingViewIDs, maskingViewID)
	Data.StorageGroupIDToStorageGroup[storageGroupID].NumOfMaskingViews++
	// Update the volume cache
	for _, volumeID := range Data.StorageGroupIDToVolumes[storageGroupID] {
		Data.VolumeIDToVolume[volumeID].NumberOfFrontEndPaths = 1
	}
	return Data.MaskingViewIDToMaskingView[maskingViewID], nil
}

// RenameMaskingView - Renames masking view
func RenameMaskingView(w http.ResponseWriter, param *types.RenameMaskingViewParam, maskingViewID string, executionOption string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	renameMaskingView(w, param, maskingViewID, executionOption)
}

func renameMaskingView(w http.ResponseWriter, param *types.RenameMaskingViewParam, maskingViewID string, executionOption string) {
	if executionOption != types.ExecutionOptionSynchronous {
		writeError(w, "expected SYNCHRONOUS", http.StatusBadRequest)
		return
	}
	Data.MaskingViewIDToMaskingView[maskingViewID].MaskingViewID = param.NewMaskingViewName
	returnMaskingView(w, maskingViewID)
}

// RemoveMaskingView - Removes a masking view from the mock data cache
func RemoveMaskingView(w http.ResponseWriter, maskingViewID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	removeMaskingView(w, maskingViewID)
}

func removeMaskingView(w http.ResponseWriter, maskingViewID string) {
	mv, ok := Data.MaskingViewIDToMaskingView[maskingViewID]
	if !ok {
		fmt.Println("Masking View " + maskingViewID + " doesn't exist")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Handle storage groups
	storageGroupID := mv.StorageGroupID
	Data.StorageGroupIDToStorageGroup[storageGroupID].NumOfMaskingViews--
	currentMaskingViewIDs := Data.StorageGroupIDToStorageGroup[storageGroupID].MaskingView
	newMaskingViewIDs := make([]string, 0)
	for _, mvID := range currentMaskingViewIDs {
		if mvID != maskingViewID {
			newMaskingViewIDs = append(newMaskingViewIDs, mvID)
		}
	}
	Data.StorageGroupIDToStorageGroup[storageGroupID].MaskingView = newMaskingViewIDs
	// Handle Hosts
	hostID := mv.HostID
	Data.HostIDToHost[hostID].NumberMaskingViews--
	currentMaskingViewIDs = Data.HostIDToHost[hostID].MaskingviewIDs
	newMaskingViewIDs = make([]string, 0)
	for _, mvID := range currentMaskingViewIDs {
		if mvID != maskingViewID {
			newMaskingViewIDs = append(newMaskingViewIDs, mvID)
		}
	}
	Data.HostIDToHost[hostID].MaskingviewIDs = newMaskingViewIDs
	// Check if we need to update the number of front end paths for volumes
	// Loop through volumes of this particular SG
	if volumeIDs, ok := Data.StorageGroupIDToVolumes[storageGroupID]; ok {
		// First construct a list of all SGs
		tempSGList := make([]string, 0)
		for _, volumeID := range volumeIDs {
			if vol, ok1 := Data.VolumeIDToVolume[volumeID]; ok1 {
				tempSGList = append(tempSGList, vol.StorageGroupIDList...)
			}
		}
		// Remove duplicates
		tempSGList = uniqueElements(tempSGList)
		// Filter out SGs in masking Views
		sgIDsInMaskingView := make([]string, 0)
		for _, sgID := range tempSGList {
			if sg, ok1 := Data.StorageGroupIDToStorageGroup[sgID]; ok1 {
				if sg.NumOfMaskingViews > 0 {
					sgIDsInMaskingView = append(sgIDsInMaskingView, sgID)
				}
			}
		}
		// Now Update the number of front end paths
		for _, volumeID := range volumeIDs {
			if vol, ok1 := Data.VolumeIDToVolume[volumeID]; ok1 {
				update := compareAndCheck(vol.StorageGroupIDList, sgIDsInMaskingView)
				if update {
					vol.NumberOfFrontEndPaths = 0
				}
			}
		}
	}
	delete(Data.StorageGroupIDToStorageGroup, maskingViewID)
}

// compareAndCheck - compares two string slices and returns true if the slices are equal or false if they aren't
func compareAndCheck(slice1 []string, slice2 []string) bool {
	for _, item := range slice1 {
		for _, item1 := range slice2 {
			if item == item1 {
				return false
			}
		}
	}
	return true
}

// uniqueElements - Removes duplicates from a string slice and returns a slice containing unique elements only
func uniqueElements(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// newVolume creates a new mock volume with the specified characteristics.
func newVolume(volumeID, volumeIdentifier string, size int, sgList []string) {
	volume := &types.Volume{
		VolumeID:              volumeID,
		Type:                  "TDEV",
		Emulation:             "FBA",
		SSID:                  "FFFFFFFF",
		AllocatedPercent:      0,
		CapacityGB:            0.0,
		FloatCapacityMB:       0.0,
		CapacityCYL:           size,
		Status:                "Ready",
		Reserved:              false,
		Pinned:                false,
		VolumeIdentifier:      volumeIdentifier,
		WWN:                   "600009700001979000465330303" + volumeID,
		EffectiveWWN:          "600009700001979000465330303" + volumeID,
		Encapsulated:          false,
		NumberOfStorageGroups: 1,
		NumberOfFrontEndPaths: 0,
		StorageGroupIDList:    sgList,
	}
	if _, ok := Data.StorageGroupIDToRDFStorageGroup[sgList[0]]; ok {
		volume.Type = "RDF1+TDEV"
		volume.RDFGroupIDList = []types.RDFGroupID{
			{RDFGroupNumber: Data.RDFGroup.RdfgNumber},
		}
	}
	Data.VolumeIDToVolume[volumeID] = volume
}

// AddNewVolume - Add a volume to the mock data cache
func AddNewVolume(volumeID, volumeIdentifier string, size int, storageGroupID string) error {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	return addNewVolume(volumeID, volumeIdentifier, size, storageGroupID)
}

func addNewVolume(volumeID, volumeIdentifier string, size int, storageGroupID string) error {
	if _, ok := Data.VolumeIDToVolume[volumeID]; ok {
		return errors.New("The requested volume already exists")
	}
	if _, ok := Data.StorageGroupIDToStorageGroup[storageGroupID]; !ok {
		return errors.New("The requested storage group resource doesn't exist")
	}
	sgList := []string{storageGroupID}
	newVolume(volumeID, volumeIdentifier, size, sgList)
	Data.StorageGroupIDToStorageGroup[storageGroupID].NumOfVolumes++
	currentVolumes := Data.StorageGroupIDToVolumes[storageGroupID]
	newVolumes := append(currentVolumes, volumeID)
	Data.StorageGroupIDToVolumes[storageGroupID] = newVolumes
	return nil
}

func newInitiator(initiatorID string, initiatorName string, initiatorType string, dirPortKeys []types.PortKey, hostID string) {
	//maskingViewIDs := []string{}
	initiator := &types.Initiator{
		InitiatorID:          initiatorName,
		SymmetrixPortKey:     dirPortKeys,
		InitiatorType:        initiatorType,
		FCID:                 "0",
		IPAddress:            "192.168.1.175",
		HostID:               hostID,
		HostGroupIDs:         []string{},
		LoggedIn:             true,
		OnFabric:             true,
		FlagsInEffect:        "Common_Serial_Number(C), SCSI_3(SC3), SPC2_Protocol_Version(SPC2)",
		NumberVols:           1,
		NumberHostGroups:     0,
		NumberMaskingViews:   0,
		NumberPowerPathHosts: 0,
	}
	Data.InitiatorIDToInitiator[initiatorID] = initiator
}

// AddInitiator - Adds an initiator to the mock data cache
func AddInitiator(initiatorID string, initiatorName string, initiatorType string, dirPortKeys []string, hostID string) (*types.Initiator, error) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	if _, ok := Data.InitiatorIDToInitiator[initiatorID]; ok {
		return nil, errors.New("Error! Initiator already exists")
	}
	// if host id is supplied, check for existence of host
	if hostID != "" {
		if _, ok := Data.HostIDToHost[hostID]; !ok {
			return nil, errors.New("Error! Host doesn't exist")
		}
	}
	portKeys := make([]types.PortKey, 0)
	for _, dirPortKey := range dirPortKeys {
		dirPortDetails := strings.Split(dirPortKey, ":")
		portKey := types.PortKey{
			DirectorID: dirPortDetails[0],
			PortID:     dirPortKey,
		}
		portKeys = append(portKeys, portKey)
	}
	newInitiator(initiatorID, initiatorName, initiatorType, portKeys, hostID)
	return Data.InitiatorIDToInitiator[initiatorID], nil
}

// ReturnInitiator - Returns initiator from mock cache based on initiator id
func ReturnInitiator(w http.ResponseWriter, initiatorID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	returnInitiator(w, initiatorID)
}

func returnInitiator(w http.ResponseWriter, initiatorID string) {
	if initiatorID != "" {
		if init, ok := Data.InitiatorIDToInitiator[initiatorID]; ok {
			writeJSON(w, init)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		initIDs := make([]string, 0)
		for k := range Data.InitiatorIDToInitiator {
			initIDs = append(initIDs, k)
		}
		initiatorIDList := &types.InitiatorList{
			InitiatorIDs: initIDs,
		}
		writeJSON(w, initiatorIDList)
	}
}

func newHost(hostID string, hostType string, initiatorIDs []string) {
	maskingViewIDs := []string{}
	host := &types.Host{
		HostID:             hostID,
		NumberMaskingViews: 0,
		NumberInitiators:   int64(len(initiatorIDs)),
		NumberHostGroups:   0,
		PortFlagsOverride:  false,
		ConsistentLun:      false,
		EnabledFlags:       "",
		DisabledFlags:      "",
		HostType:           hostType,
		Initiators:         initiatorIDs,
		MaskingviewIDs:     maskingViewIDs,
		NumPowerPathHosts:  0,
	}
	Data.HostIDToHost[hostID] = host
}

// AddHost - Adds a host to the mock data cache
func AddHost(hostID string, hostType string, initiatorIDs []string) (*types.Host, error) {
	if _, ok := Data.HostIDToHost[hostID]; ok {
		return nil, errors.New("Error! Host already exists")
	}
	validInitiators := false
	// Check if initiators exist
	for _, initID := range initiatorIDs {
		for _, v := range Data.InitiatorIDToInitiator {
			if v.InitiatorID == initID {
				if v.HostID == "" {
					validInitiators = true
					break
				}
			}
		}
		if !validInitiators {
			break
		}
	}
	if !validInitiators {
		errormsg := "Error! Some initiators don't exist or are not valid"
		fmt.Println(errormsg)
		return nil, fmt.Errorf(errormsg)
	}
	newHost(hostID, hostType, initiatorIDs)
	//Update the initiators
	for _, initID := range initiatorIDs {
		for k, v := range Data.InitiatorIDToInitiator {
			if v.InitiatorID == initID {
				Data.InitiatorIDToInitiator[k].HostID = hostID
				break
			}
		}
	}
	fmt.Println(Data.HostIDToHost[hostID])
	return Data.HostIDToHost[hostID], nil
}

// RemoveHost - Removes host from mock cache
func RemoveHost(hostID string) error {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	return removeHost(hostID)
}

// removeHost - Remove a host from the mock data cache
func removeHost(hostID string) error {
	host, ok := Data.HostIDToHost[hostID]
	if !ok {
		return errors.New("Error! Host doesn't exist")
	}
	if host.NumberMaskingViews > 0 {
		return errors.New("Error! Host is part of a masking view")
	}
	Data.HostIDToHost[hostID] = nil
	return nil
}

func newPortGroup(portGroupID string, portGroupType string, portKeys []types.PortKey) {
	portGroup := &types.PortGroup{
		PortGroupID:        portGroupID,
		SymmetrixPortKey:   portKeys,
		NumberPorts:        int64(len(portKeys)),
		NumberMaskingViews: 0,
		PortGroupType:      portGroupType,
	}
	Data.PortGroupIDToPortGroup[portGroupID] = portGroup
}

// addPortGroup - Adds a port group to the mock data cache
func addPortGroup(portGroupID string, portGroupType string, portKeys []types.PortKey) (*types.PortGroup, error) {
	if _, ok := Data.PortGroupIDToPortGroup[portGroupID]; ok {
		return nil, errors.New("Error! Port Group already exists")
	}
	newPortGroup(portGroupID, portGroupType, portKeys)
	return Data.PortGroupIDToPortGroup[portGroupID], nil
}

// updatePortGroup - Update PortGroup by ID by adding 'addKeys' and removing 'removeKeys'
func updatePortGroup(portGroupID string, editPayload *types.EditPortGroupActionParam) (*types.PortGroup, error) {
	pg, ok := Data.PortGroupIDToPortGroup[portGroupID]
	if !ok {
		return nil, fmt.Errorf("error! PortGroup %s does not exist", portGroupID)
	}

	// Collect the ports to add (if any)
	addKeys := make([]types.PortKey, 0)
	if editPayload.AddPortParam != nil {
		addKeys = convertToPortKeys(editPayload.AddPortParam.Ports)
	}

	// Collect the ports to remove (if any)
	removeKeys := make([]types.PortKey, 0)
	if editPayload.RemovePortParam != nil {
		removeKeys = convertToPortKeys(editPayload.RemovePortParam.Ports)
	}

	// Add to the list of ports
	pg.SymmetrixPortKey = append(pg.SymmetrixPortKey, addKeys...)

	// Remove from the list of ports in the PortGroup
	for _, key := range removeKeys {
		pg.SymmetrixPortKey = removePortKey(pg.SymmetrixPortKey, key)
	}

	if editPayload.RenamePortGroupParam != nil && editPayload.RenamePortGroupParam.NewPortGroupName != "" {
		portGroupID = editPayload.RenamePortGroupParam.NewPortGroupName
		pg.PortGroupID = portGroupID
	}

	// Update the PortGroup mapping with the update PortGroup
	Data.PortGroupIDToPortGroup[portGroupID] = pg
	return pg, nil
}

// convertToPortKeys - Convert a slice of types.SymmetrixPortKeyType to slice of types.PortKey
func convertToPortKeys(symmPorts []types.SymmetrixPortKeyType) []types.PortKey {
	if symmPorts == nil || len(symmPorts) == 0 {
		return make([]types.PortKey, 0)
	}

	out := make([]types.PortKey, len(symmPorts))
	for idx, it := range symmPorts {
		out[idx] = types.PortKey{
			DirectorID: it.DirectorID,
			PortID:     it.PortID,
		}
	}

	return out
}

// removePortKey - delete PortKey 'key' from the slice
func removePortKey(slice []types.PortKey, keyToRemove types.PortKey) []types.PortKey {
	index := -1
	// Find the index in the slice that has the match
	for it, thisKey := range slice {
		if thisKey.DirectorID == keyToRemove.DirectorID && thisKey.PortID == keyToRemove.PortID {
			index = it
			break
		}
	}
	if index != -1 {
		// Found the index with matching port
		copy(slice[index:], slice[index+1:])
		return slice[:len(slice)-1]
	}
	// No match was found, return unchanged slice
	return slice

}

// UpdatePortGroupFromParams - Updates PortGroup given an EditPortGroup payload
func UpdatePortGroupFromParams(portGroupID string, updateParams *types.EditPortGroup) {
	updatePortGroup(portGroupID, updateParams.EditPortGroupActionParam) // #nosec G20
}

// DeletePortGroup - Remove PortGroup by ID 'portGroupID'
func DeletePortGroup(portGroupID string) (*types.PortGroup, error) {
	pg, ok := Data.PortGroupIDToPortGroup[portGroupID]
	if !ok {
		return nil, fmt.Errorf("error! PortGroup %s does not exist", portGroupID)
	}
	delete(Data.PortGroupIDToPortGroup, portGroupID)
	return pg, nil
}

// AddPortGroupFromCreateParams - Adds a storage group from create params
func AddPortGroupFromCreateParams(createParams *types.CreatePortGroupParams) {
	portGroupID := createParams.PortGroupID
	portKeys := createParams.SymmetrixPortKey
	addPortGroup(portGroupID, "Fibre", portKeys) // #nosec G20
}

// AddPortGroup - Adds a port group to the mock data cache
func AddPortGroup(portGroupID string, portGroupType string, portIdentifiers []string) (*types.PortGroup, error) {
	portKeys := make([]types.PortKey, 0)
	for _, dirPortKey := range portIdentifiers {
		dirPortDetails := strings.Split(dirPortKey, ":")
		if len(dirPortDetails) != 2 {
			errormsg := fmt.Sprintf("Invalid dir port specified: %s", dirPortKey)
			log.Error(errormsg)
			return nil, fmt.Errorf(errormsg)
		}
		portKey := types.PortKey{
			DirectorID: dirPortDetails[0],
			PortID:     dirPortKey,
		}
		portKeys = append(portKeys, portKey)
	}
	if _, ok := Data.PortGroupIDToPortGroup[portGroupID]; ok {
		return nil, errors.New("Error! Port Group already exists")
	}
	newPortGroup(portGroupID, portGroupType, portKeys)
	return Data.PortGroupIDToPortGroup[portGroupID], nil
}

// AddStorageGroupFromCreateParams - Adds a storage group from create params
func AddStorageGroupFromCreateParams(createParams *types.CreateStorageGroupParam) {
	sgID := createParams.StorageGroupID
	srpID := createParams.SRPID
	serviceLevel := "None"
	if srpID != "None" {
		sloBasedParams := createParams.SLOBasedStorageGroupParam
		serviceLevel = sloBasedParams[0].SLOID
	} else {
		srpID = ""
	}
	AddStorageGroup(sgID, srpID, serviceLevel) // #nosec G20
}

// keys - Return keys of the given map
func keys(m map[string]*types.StorageGroup) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ReturnStorageGroup - Returns storage group information from mock cache
func ReturnStorageGroup(w http.ResponseWriter, sgID string, remote bool) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	returnStorageGroup(w, sgID, remote)
}

func returnStorageGroup(w http.ResponseWriter, sgID string, remote bool) {
	if sgID != "" {
		if InducedErrors.GetSGOnRemote && remote {
			sg := Data.StorageGroupIDToStorageGroup["CSI-Test-Fake-Remote-SG"]
			fmt.Printf("Fake remote SG: %#v\n", sg)
			writeJSON(w, sg)
			return
		}
		if InducedErrors.GetSGWithVolOnRemote && remote {
			sg := Data.StorageGroupIDToStorageGroup["CSI-Test-Fake-Remote-SG"]
			sg.NumOfVolumes = 1
			fmt.Printf("Fake remote SG: %#v\n", sg)
			writeJSON(w, sg)
			return
		}
		if _, ok := Data.StorageGroupIDToRDFStorageGroup[sgID]; remote && !ok {
			writeError(w, "StorageGroup not found", http.StatusNotFound)
			return
		}
		if sg, ok := Data.StorageGroupIDToStorageGroup[sgID]; ok {
			fmt.Printf("SG: %#v\n", sg)
			writeJSON(w, sg)
			return
		}
		writeError(w, "StorageGroup not found", http.StatusNotFound)
	} else {
		storageGroupIDs := keys(Data.StorageGroupIDToStorageGroup)
		storageGroupIDList := &types.StorageGroupIDList{
			StorageGroupIDs: storageGroupIDs,
		}
		writeJSON(w, storageGroupIDList)
	}
}

func returnMaskingView(w http.ResponseWriter, mvID string) {
	if mvID != "" {
		if mv, ok := Data.MaskingViewIDToMaskingView[mvID]; ok {
			writeJSON(w, mv)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		maskingViewIDs := make([]string, 0)
		for k := range Data.MaskingViewIDToMaskingView {
			maskingViewIDs = append(maskingViewIDs, k)
		}
		maskingViewIDList := &types.MaskingViewList{
			MaskingViewIDs: maskingViewIDs,
		}
		writeJSON(w, maskingViewIDList)
	}
}

func writeJSON(w http.ResponseWriter, val interface{}) {
	if InducedErrors.InvalidResponse {
		fmt.Println("Inducing error")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		fmt.Println("error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(jsonBytes)
	if err != nil {
		log.Printf("Couldn't write to ResponseWriter")
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
}

// AddOneVolumeToStorageGroup - Adds volume to a storage group in the mock cache
func AddOneVolumeToStorageGroup(volumeID, volumeIdentifier, sgID string, size int) error {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	return addOneVolumeToStorageGroup(volumeID, volumeIdentifier, sgID, size)
}

func addOneVolumeToStorageGroup(volumeID, volumeIdentifier, sgID string, size int) error {
	if _, ok := Data.StorageGroupIDToStorageGroup[sgID]; !ok {
		return errors.New("The requested storage group doesn't exist")
	}
	if _, ok := Data.VolumeIDToVolume[volumeID]; ok {
		// Found the volume in cache
		// We are adding it to another storage group
		if volumes, ok := Data.StorageGroupIDToVolumes[sgID]; ok {
			found := false
			for _, volume := range volumes {
				if strings.Contains(volume, volumeID) {
					found = true
					break
				}
			}
			if found {
				return errors.New("Volume is already a part of the SG")
			}
			// Update the volume cache
			currentStorageGroups := Data.VolumeIDToVolume[volumeID].StorageGroupIDList
			newStorageGroups := append(currentStorageGroups, sgID)
			Data.VolumeIDToVolume[volumeID].StorageGroupIDList = newStorageGroups
			// Update volume's replication details in case the storage-group is replicated
			if _, ok := Data.StorageGroupIDToRDFStorageGroup[sgID]; ok {
				Data.VolumeIDToVolume[volumeID].Type = "RDF1+TDEV"
				Data.VolumeIDToVolume[volumeID].RDFGroupIDList = []types.RDFGroupID{
					{RDFGroupNumber: Data.RDFGroup.RdfgNumber},
				}
			}

			// Update the Storage Group caches
			Data.StorageGroupIDToStorageGroup[sgID].NumOfVolumes++
			currentVolumes := Data.StorageGroupIDToVolumes[sgID]
			newVolumes := append(currentVolumes, volumeID)
			Data.StorageGroupIDToVolumes[sgID] = newVolumes

			// Check if the volume was added to a masking view
			if Data.StorageGroupIDToStorageGroup[sgID].NumOfMaskingViews > 0 {
				Data.VolumeIDToVolume[volumeID].NumberOfFrontEndPaths = 1
			}
			Data.VolumeIDToVolume[volumeID].NumberOfStorageGroups++
		}
	} else {
		// We are adding a new volume
		addNewVolume(volumeID, volumeIdentifier, size, sgID) // #nosec G20
	}
	return nil
}

// AddVolumeToStorageGroupTest - Adds volume to storage group and updates mock cache
func AddVolumeToStorageGroupTest(w http.ResponseWriter, name, size, sgID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	addVolumeToStorageGroupTest(w, name, size, sgID)
}

func addVolumeToStorageGroupTest(w http.ResponseWriter, name, size, sgID string) {
	if name == "" || size == "" {
		writeError(w, "null name or size", http.StatusBadRequest)
	}
	id := strconv.Itoa(time.Now().Nanosecond())
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		writeError(w, "unable to convert size string to integer", http.StatusBadRequest)
	}
	if InducedErrors.VolumeNotCreatedError == false {
		addOneVolumeToStorageGroup(id, name, sgID, sizeInt) // #nosec G20
	}
	// Make a job to return
	resourceLink := fmt.Sprintf("sloprovisioning/system/%s/storagegroup/%s", DefaultSymmetrixID, sgID)
	if InducedErrors.JobFailedError {
		newMockJob(id, types.JobStatusRunning, types.JobStatusFailed, resourceLink)
	} else {
		newMockJob(id, types.JobStatusRunning, types.JobStatusSucceeded, resourceLink)
	}
	returnJobByID(w, id)
}

// AddSpecificVolumeToStorageGroup - Add volume based on volumeids to storage group mock cache
func AddSpecificVolumeToStorageGroup(w http.ResponseWriter, volumeIDs []string, sgID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	addSpecificVolumeToStorageGroup(w, volumeIDs, sgID)
}

func addSpecificVolumeToStorageGroup(w http.ResponseWriter, volumeIDs []string, sgID string) {
	if len(volumeIDs) == 0 {
		writeError(w, "empty list", http.StatusBadRequest)
	}
	jobID := strconv.Itoa(time.Now().Nanosecond())
	if InducedErrors.VolumeNotAddedError {
		writeError(w, "Error adding volume to the SG", http.StatusRequestTimeout)
		return
	}
	for _, volumeID := range volumeIDs {
		addOneVolumeToStorageGroup(volumeID, "TestVol", sgID, 0) // #nosec G20
	}
	// Make a job to return
	resourceLink := fmt.Sprintf("sloprovisioning/system/%s/storagegroup/%s", DefaultSymmetrixID, sgID)
	if InducedErrors.JobFailedError {
		newMockJob(jobID, types.JobStatusRunning, types.JobStatusFailed, resourceLink)
	} else {
		newMockJob(jobID, types.JobStatusRunning, types.JobStatusSucceeded, resourceLink)
	}
	returnJobByID(w, jobID)
}

func removeOneVolumeFromStorageGroup(volumeID, storageGroupID string) error {
	if _, ok := Data.StorageGroupIDToStorageGroup[storageGroupID]; !ok {
		return errors.New("The requested storage group doesn't exist")
	}
	if _, ok := Data.StorageGroupIDToVolumes[storageGroupID]; !ok {
		return errors.New("Storage Group to volume mapping doesn't exist")
	}
	vol, ok := Data.VolumeIDToVolume[volumeID]
	if !ok {
		return errors.New("The requested volume doesn't exist")
	}
	// Remove SG from the volume's SG list
	currentSGList := vol.StorageGroupIDList
	newStorageGroupList := make([]string, 0)
	for _, sgID := range currentSGList {
		if sgID != storageGroupID {
			newStorageGroupList = append(newStorageGroupList, sgID)
		}
	}
	vol.StorageGroupIDList = newStorageGroupList
	vol.NumberOfStorageGroups--
	// Change Volume's replication properties if replicated
	removeReplicationProps := false
	_, removeReplicationProps = Data.StorageGroupIDToRDFStorageGroup[storageGroupID]
	for _, sgID := range vol.StorageGroupIDList {
		if _, ok := Data.StorageGroupIDToRDFStorageGroup[sgID]; ok {
			removeReplicationProps = false
			break
		}
	}
	if removeReplicationProps {
		vol.Type = "TDEV"
		vol.RDFGroupIDList = nil
	}
	// Remove volume from the SG's volume list
	currentVolumeIDs := Data.StorageGroupIDToVolumes[storageGroupID]
	newVolumeIDList := make([]string, 0)
	for _, volID := range currentVolumeIDs {
		if volID != volumeID {
			newVolumeIDList = append(newVolumeIDList, volID)
		}
	}
	Data.StorageGroupIDToVolumes[storageGroupID] = newVolumeIDList
	// Update the count of volumes in SG
	Data.StorageGroupIDToStorageGroup[storageGroupID].NumOfVolumes--
	// Check if we need to update the number of front end paths for this particular volume
	update := true
	for _, sgID := range vol.StorageGroupIDList {
		if sg, ok := Data.StorageGroupIDToStorageGroup[sgID]; ok {
			if sg.NumOfMaskingViews > 1 {
				update = false
				break
			}
		}
	}
	if update {
		vol.NumberOfFrontEndPaths = 0
	}
	return nil
}

// RemoveVolumeFromStorageGroup - Remove volumes from storage group mock cache
func RemoveVolumeFromStorageGroup(w http.ResponseWriter, volumeIDs []string, sgID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	removeVolumeFromStorageGroup(w, volumeIDs, sgID)
}

func removeVolumeFromStorageGroup(w http.ResponseWriter, volumeIDs []string, sgID string) {
	for _, volID := range volumeIDs {
		fmt.Println("Volume ID: " + volID)
		removeOneVolumeFromStorageGroup(volID, sgID) // #nosec G20
	}
	returnStorageGroup(w, sgID, false)
}

// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/portgroup/{id}
// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/portgroup
func handlePortGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pgID := vars["id"]
	switch r.Method {

	case http.MethodGet:
		if InducedErrors.GetPortGroupError {
			writeError(w, "Error retrieving Port Group(s): induced error", http.StatusRequestTimeout)
			return
		}
		ReturnPortGroup(w, pgID)

	case http.MethodPost:
		if InducedErrors.CreatePortGroupError {
			writeError(w, "Error creating Port Group: induced error", http.StatusRequestTimeout)
			return
		}
		decoder := json.NewDecoder(r.Body)
		createPortGroupParams := &types.CreatePortGroupParams{}
		err := decoder.Decode(createPortGroupParams)
		if err != nil {
			writeError(w, "InvalidJson", http.StatusBadRequest)
			return
		}
		AddPortGroupFromCreateParams(createPortGroupParams)
		ReturnPortGroup(w, createPortGroupParams.PortGroupID)
	case http.MethodPut:
		if InducedErrors.UpdatePortGroupError {
			writeError(w, "Error updating Port Group: induced error", http.StatusRequestTimeout)
			return
		}
		decoder := json.NewDecoder(r.Body)
		updatePortGroupParams := &types.EditPortGroup{}
		err := decoder.Decode(updatePortGroupParams)
		if err != nil {
			writeError(w, "InvalidJson", http.StatusBadRequest)
			return
		}
		UpdatePortGroupFromParams(pgID, updatePortGroupParams)
		ReturnPortGroup(w, pgID)
	case http.MethodDelete:
		if InducedErrors.DeletePortGroupError {
			writeError(w, "Error deleting Port Group: induced error", http.StatusRequestTimeout)
			return
		}
		DeletePortGroup(pgID) // #nosec G20
	default:
		writeError(w, "Invalid Method", http.StatusBadRequest)
	}
}

// /univmax/restapi/90/system/symmetrix/{symid}/director/{director}/port/{id}
// /univmax/restapi/90/system/symmetrix/{symid}/director/{director}/port
func handlePort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dID := vars["director"]
	pID := vars["id"]
	queryString := r.URL.Query()
	switch r.Method {

	case http.MethodGet:
		if InducedErrors.GetPortError {
			writeError(w, "Error retrieving Port(s): induced error", http.StatusRequestTimeout)
			return
		}
		if InducedErrors.GetPortGigEError {
			queryType, ok := queryString["type"]
			if ok {
				if queryType[0] == "Gige" { // The first ?type=<value>
					writeError(w, "Error retrieving GigE ports: induced error", http.StatusRequestTimeout)
					return
				}
			}
		}
		if InducedErrors.GetPortISCSITargetError {
			queryType, ok := queryString["iscsi_target"]
			if ok {
				if queryType[0] == "true" { // The first ?iscsi_target=<value>
					writeError(w, "Error retrieving ISCSI targets: induced error", http.StatusRequestTimeout)
					return
				}
			}
		}
		mockCacheMutex.Lock()
		defer mockCacheMutex.Unlock()
		// if we asked for a specific Port, return those details
		if pID != "" {
			if InducedErrors.GetSpecificPortError {
				writeError(w, "Error retrieving Specific Port: induced error", http.StatusRequestTimeout)
				return
			}
			// Specific ports can be modeleted
			portName := dID + ":" + pID
			if Data.PortIDToSymmetrixPortType[portName] != nil {
				port := Data.PortIDToSymmetrixPortType[portName]
				if port == nil || port.Type == "" {
					writeError(w, "port not found", http.StatusNotFound)
				} else {
					symPort := &types.Port{
						SymmetrixPort: *port,
					}
					encoder := json.NewEncoder(w)
					encoder.Encode(symPort) // #nosec G20
				}
				return
			}
			returnPort(w, dID, pID)
		}
		// return a list of Ports
		returnPortIDList(w, dID)
	default:
		writeError(w, "Invalid Method", http.StatusBadRequest)
	}
}

// AddPort adds a port entry. Port type can either be "FibreChannel" or "GigE", or "" for a non existent port.
func AddPort(id, identifier, portType string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	port := &types.SymmetrixPortType{
		Type:       portType,
		Identifier: identifier,
	}
	Data.PortIDToSymmetrixPortType[id] = port
}

func returnPort(w http.ResponseWriter, dID, pID string) {
	replacements := make(map[string]string)
	replacements["__PORT_ID__"] = pID
	replacements["__DIRECTOR_ID__"] = dID
	returnJSONFile(Data.JSONDir, "port_template.json", w, replacements)
}

func returnPortIDList(w http.ResponseWriter, dID string) {
	replacements := make(map[string]string)
	replacements["__DIRECTOR_ID__"] = dID
	returnJSONFile(Data.JSONDir, "portIDList.json", w, replacements)
}

// /univmax/restapi/90/system/symmetrix/{symid}/director/{{id}
// /univmax/restapi/90/system/symmetrix/{symid}/director
func handleDirector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dID := vars["id"]
	switch r.Method {

	case http.MethodGet:
		if InducedErrors.GetDirectorError {
			writeError(w, "Error retrieving Director(s): induced error", http.StatusRequestTimeout)
			return
		}
		// if we asked for a specific Director, return those details
		if dID != "" {
			returnDirector(w, dID)
		}
		// return a list of Directors
		returnDirectorIDList(w)

	default:
		writeError(w, "Invalid Method", http.StatusBadRequest)
	}
}

func returnDirector(w http.ResponseWriter, dID string) {
	replacements := make(map[string]string)
	replacements["__DIRECTOR_ID__"] = dID
	returnJSONFile(Data.JSONDir, "director_template.json", w, replacements)
}

func returnDirectorIDList(w http.ResponseWriter) {
	replacements := make(map[string]string)
	returnJSONFile(Data.JSONDir, "directorIDList.json", w, replacements)
}

// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/initiator/{id}
// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/initiator
func handleInitiator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	initID := vars["id"]
	switch r.Method {

	case http.MethodGet:
		if InducedErrors.GetInitiatorError {
			writeError(w, "Error retrieving Initiator(s): induced error", http.StatusRequestTimeout)
			return
		}
		if initID != "" {
			if InducedErrors.GetInitiatorByIDError {
				writeError(w, "Error retrieving Initiator By ID: induced error", http.StatusRequestTimeout)
				return
			}
		}
		ReturnInitiator(w, initID)

	default:
		writeError(w, "Invalid Method", http.StatusBadRequest)
	}
}

// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/host/{id}
// /univmax/restapi/90/sloprovisioning/symmetrix/{symid}/host
func handleHost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostID := vars["id"]
	switch r.Method {

	case http.MethodGet:
		if InducedErrors.GetHostError {
			writeError(w, "Error retrieving Host(s): induced error", http.StatusRequestTimeout)
			return
		}
		ReturnHost(w, hostID)

	case http.MethodPost:
		if InducedErrors.CreateHostError {
			writeError(w, "Error creating Host: induced error", http.StatusRequestTimeout)
			return
		}
		decoder := json.NewDecoder(r.Body)
		createHostParam := &types.CreateHostParam{}
		err := decoder.Decode(createHostParam)
		if err != nil {
			writeError(w, "InvalidJson", http.StatusBadRequest)
			return
		}
		// Scan the initiators to see if there are any non iqn ones; then assume
		// host type Fibre.
		isFibre := false
		for _, initiator := range createHostParam.InitiatorIDs {
			if !strings.HasPrefix(initiator, "iqn.") {
				isFibre = true
			}
		}
		if isFibre {
			// Might need to add the Port information here
			AddHost(createHostParam.HostID, "Fibre", createHostParam.InitiatorIDs) // #nosec G20
		} else {
			//initNode := make([]string, 0)
			//initNode = append(initNode, "iqn.1993-08.org.centos:01:5ae577b352a7")
			AddHost(createHostParam.HostID, "iSCSI", createHostParam.InitiatorIDs) // #nosec G20
		}
		ReturnHost(w, createHostParam.HostID)

	case http.MethodPut:
		if hasError(&InducedErrors.UpdateHostError) {
			// if InducedErrors.UpdateHostError {
			writeError(w, "Error updating Host: induced error", http.StatusRequestTimeout)
			return
		}
		decoder := json.NewDecoder(r.Body)
		updateHostParam := &types.UpdateHostParam{}
		err := decoder.Decode(updateHostParam)
		if err != nil {
			writeError(w, "InvalidJson", http.StatusBadRequest)
			return
		}
		ReturnHost(w, hostID)

	case http.MethodDelete:
		if InducedErrors.DeleteHostError {
			writeError(w, "Error deleting Host: induced error", http.StatusRequestTimeout)
			return
		}
		RemoveHost(hostID) // #nosec G20
	default:
		writeError(w, "Invalid Method", http.StatusBadRequest)
	}
}

// ReturnHost - Returns a host from cache
func ReturnHost(w http.ResponseWriter, hostID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	returnHost(w, hostID)
}

func returnHost(w http.ResponseWriter, hostID string) {
	if hostID != "" {
		if host, ok := Data.HostIDToHost[hostID]; ok {
			writeJSON(w, host)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		hostIDs := make([]string, 0)
		for k := range Data.HostIDToHost {
			hostIDs = append(hostIDs, k)
		}
		hostIDList := &types.HostList{
			HostIDs: hostIDs,
		}
		writeJSON(w, hostIDList)
	}
}

// ReturnPortGroup - Returns port group information from cache
func ReturnPortGroup(w http.ResponseWriter, portGroupID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	returnPortGroup(w, portGroupID)
}

func returnPortGroup(w http.ResponseWriter, portGroupID string) {
	if portGroupID != "" {
		if pg, ok := Data.PortGroupIDToPortGroup[portGroupID]; ok {
			fmt.Printf("\n%v\n", pg)
			writeJSON(w, pg)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		portGroupIDs := make([]string, 0)
		for k := range Data.PortGroupIDToPortGroup {
			portGroupIDs = append(portGroupIDs, k)
		}
		portGroupList := &types.PortGroupList{
			PortGroupIDs: portGroupIDs,
		}
		writeJSON(w, portGroupList)
	}
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	writeError(w, "URL not found: "+r.URL.String(), http.StatusNotFound)
}

// Write an error code to the response writer
func writeError(w http.ResponseWriter, message string, httpStatus int) {
	w.WriteHeader(httpStatus)
	resp := new(types.Error)
	resp.Message = message
	// The following aren't used by the hardware but could be used internally
	//resp.HTTPStatusCode = http.StatusNotFound
	//resp.ErrorCode = int(errorCode)
	encoder := json.NewEncoder(w)
	err := encoder.Encode(resp)
	if err != nil {
		log.Printf("error encoding json: %s", err.Error())
	}
}

// Return content from a JSON file. Arguments are:
//
//	 directory, filename  of the file
//	wrriter ResponseWriter where data is output
//
// An optional replacement map. If supplied every instance of a key in the JSON file will be replaced with the corresponding value.
func returnJSONFile(directory, filename string, w http.ResponseWriter, replacements map[string]string) (jsonBytes []byte) {
	jsonBytes, err := ioutil.ReadFile(filepath.Join(directory, filename)) // #nosec G20
	if err != nil {
		log.Printf("Couldn't read %s/%s", directory, filename)
		if w != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		return make([]byte, 0)
	}
	if replacements != nil {
		jsonString := string(jsonBytes)
		for key, value := range replacements {
			jsonString = strings.Replace(jsonString, key, value, -1)
		}
		if Debug {
			log.Printf("Edited payload:%s", jsonString)
		}
		jsonBytes = []byte(jsonString)
	}
	if Debug {
		log.Printf("jsonBytes:%s", jsonBytes)
	}
	if w != nil {
		_, err = w.Write(jsonBytes)
		if err != nil {
			log.Printf("Couldn't write to ResponseWriter")
			w.WriteHeader(http.StatusInternalServerError)
			return make([]byte, 0)
		}
	}
	return jsonBytes
}

// AddTempSnapshots adds marked for deletion snapshots into mock to help snapcleanup thread to be functional
func AddTempSnapshots() {
	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("%05d", i)
		size := 7
		volumeIdentifier := "Vol" + id
		AddNewVolume(id, volumeIdentifier, size, DefaultStorageGroup) // #nosec G20
		SnapID := fmt.Sprintf("%s-%s-%d", "DEL", "snapshot", i)
		AddNewSnapshot(id, SnapID)
	}

}

// univmax/restapi/private/APIVersion/replication/symmetrix/{symid}/snapshot/{SnapID}
func handleSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// volID := vars["volID"]
	SnapID := vars["SnapID"]
	switch r.Method {
	case http.MethodPost:
		if InducedErrors.CreateSnapshotError {
			writeError(w, "Failed to create snapshot: induced error", http.StatusBadRequest)
			return
		}
		if InducedErrors.MaxSnapSessionError {
			writeError(w, "The maximum number of sessions has been exceeded for the specified Source device", http.StatusBadRequest)
			return
		}
		decoder := json.NewDecoder(r.Body)
		createSnapParam := &types.CreateVolumesSnapshot{}
		err := decoder.Decode(createSnapParam)
		if err != nil {
			writeError(w, "problem decoding POST Snapshot payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		CreateSnapshot(w, r, vars["SnapID"], createSnapParam.ExecutionOption, createSnapParam.SourceVolumeList)
		return
	case http.MethodPut:
		if SnapID == "" {
			writeError(w, "Snapshot name must be supplied", http.StatusBadRequest)
			return
		}
		decoder := json.NewDecoder(r.Body)
		updateSnapParam := &types.ModifyVolumeSnapshot{}
		err := decoder.Decode(updateSnapParam)
		if err != nil {
			writeError(w, "problem decoding PUT Snapshot payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Printf("PUT Snapshot payload: %#v\n", updateSnapParam)
		executionOption := updateSnapParam.ExecutionOption

		if updateSnapParam.Action == "Rename" {
			if InducedErrors.RenameSnapshotError {
				writeError(w, "error renaming the snapshot: induced error", http.StatusBadRequest)
				return
			}
			RenameSnapshot(w, r, updateSnapParam.VolumeNameListSource, executionOption, SnapID, updateSnapParam.NewSnapshotName)
			return
		}
		if updateSnapParam.Action == "Link" {
			if InducedErrors.MaxSnapSessionError {
				writeError(w, "The maximum number of sessions has been exceeded for the specified Source device", http.StatusBadRequest)
				return
			}
			if InducedErrors.LinkSnapshotError {
				writeError(w, "error linking the snapshot: induced error", http.StatusBadRequest)
				return
			}
			LinkSnapshot(w, r, updateSnapParam.VolumeNameListSource, updateSnapParam.VolumeNameListTarget, executionOption, SnapID)
			return
		}
		if updateSnapParam.Action == "Unlink" {
			if InducedErrors.LinkSnapshotError {
				writeError(w, "error unlinking the snapshot: induced error", http.StatusBadRequest)
				return
			}
			UnlinkSnapshot(w, r, updateSnapParam.VolumeNameListSource, updateSnapParam.VolumeNameListTarget, executionOption, SnapID)
			return
		}
		if updateSnapParam.Action == "Restore" {
			// restoreSnapshot(w, r, updateSnapParam.VolumeNameListSource, updateSnapParam.VolumeNameListTarget, executionOption, SnapID)
			// return
		}
	case http.MethodDelete:
		decoder := json.NewDecoder(r.Body)
		deleteSnapParam := &types.DeleteVolumeSnapshot{}
		err := decoder.Decode(deleteSnapParam)
		if err != nil {
			writeError(w, "problem decoding Delete Snapshot payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		DeleteSnapshot(w, r, vars["SnapID"], deleteSnapParam.ExecutionOption, deleteSnapParam.DeviceNameListSource, deleteSnapParam.Generation)
		return
	}
}

// CreateSnapshot - Creates a snapshot and updates mock cache
func CreateSnapshot(w http.ResponseWriter, r *http.Request, SnapID, executionOption string, sourceVolumeList []types.VolumeList) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	createSnapshot(w, r, SnapID, executionOption, sourceVolumeList)
}

func createSnapshot(w http.ResponseWriter, r *http.Request, SnapID, executionOption string, sourceVolumeList []types.VolumeList) {
	if strings.Contains(SnapID, ":") {
		writeError(w, "error, invalid snapshot name", http.StatusBadRequest)
		return
	}
	if executionOption != types.ExecutionOptionSynchronous {
		writeError(w, "expected SYNCHRONOUS", http.StatusBadRequest)
		return
	}
	if fewVolumeUnavalaible(sourceVolumeList) {
		writeError(w, "few devices not available", http.StatusBadRequest)
		return
	}
	// Make a job to return
	resourceLink := fmt.Sprintf("/replication/symmetrix/%s/snapshot/%s", DefaultSymmetrixID, SnapID)
	jobID := fmt.Sprintf("SnapID-%d", time.Now().Nanosecond())
	if InducedErrors.JobFailedError {
		newMockJob(jobID, types.JobStatusRunning, types.JobStatusFailed, resourceLink)
		returnJobByID(w, jobID)
		return
	}
	for i := 0; i < len(sourceVolumeList); i++ {
		source := sourceVolumeList[i].Name
		if !duplicateSnapshotCreationRequest(source, SnapID) {
			//Snapshot with unique name
			addNewSnapshot(source, SnapID)
		}
		newMockJob(jobID, types.JobStatusRunning, types.JobStatusSucceeded, resourceLink)
	}
	returnJobByID(w, jobID)
}

// AddNewSnapshot adds a snapshot to the mock cache
func AddNewSnapshot(source, SnapID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	addNewSnapshot(source, SnapID)
}

func addNewSnapshot(source, SnapID string) {
	time := time.Now().Nanosecond()
	snapshot := &types.Snapshot{
		Name:       SnapID,
		Generation: 0,
		State:      "Established",
		Timestamp:  strconv.Itoa(time),
	}
	snapIDtoSnap := Data.VolIDToSnapshots[source]
	if snapIDtoSnap == nil {
		snapIDtoSnap = map[string]*types.Snapshot{}
	}
	snapIDtoSnap[SnapID] = snapshot
	Data.VolIDToSnapshots[source] = snapIDtoSnap
	Data.VolumeIDToVolume[source].SnapSource = true
	fmt.Printf("*****added** %v***", Data.VolIDToSnapshots[source][SnapID])
	fmt.Printf("****Total Snaps on %s are: %d****", source, len(Data.VolIDToSnapshots[source]))
}

// DeleteSnapshot - Deletes a snapshot and updates mock cache
func DeleteSnapshot(w http.ResponseWriter, r *http.Request, SnapID string, executionOption string, deviceNameListSource []types.VolumeList, genID int64) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	deleteSnapshot(w, r, SnapID, executionOption, deviceNameListSource, genID)
}

func deleteSnapshot(w http.ResponseWriter, r *http.Request, SnapID string, executionOption string, deviceNameListSource []types.VolumeList, genID int64) {
	if InducedErrors.DeleteSnapshotError {
		writeError(w, "error deleting the snapshot: induced error", http.StatusBadRequest)
		return
	}
	if deviceNameListSource[0].Name == "" {
		writeError(w, "no source volume names given to link the snapshot", http.StatusBadRequest)
		return
	}
	if fewVolumeUnavalaible(deviceNameListSource) {
		writeError(w, "few devices not available", http.StatusBadRequest)
		return
	}
	resourceLink := fmt.Sprintf("/replication/symmetrix/%s/snapshot/%s", DefaultSymmetrixID, SnapID)
	jobID := fmt.Sprintf("SnapID-%d", time.Now().Nanosecond())
	if InducedErrors.JobFailedError {
		newMockJob(jobID, types.JobStatusRunning, types.JobStatusFailed, resourceLink)
	} else {
		for i := 0; i < len(deviceNameListSource); i++ {
			source := deviceNameListSource[i].Name

			//volume exists, check for availability of snapshot on it i.e, check if snapshot is found in snapIDtoSnap map "SnapID": Snapshot
			snapIDtoSnap := Data.VolIDToSnapshots[source]
			if _, ok := snapIDtoSnap[SnapID]; !ok {
				// snapshot is not found
				writeError(w, "no snapshot information", http.StatusBadRequest)
				return
			}

			//snapshot exists, check if it is linked to any target device/volumes
			snapIDtoLinkedVolKey := SnapID + ":" + source
			linkedVolume := Data.SnapIDToLinkedVol[snapIDtoLinkedVolKey]
			if len(linkedVolume) > 0 {
				//snapshot is linked to some volumes, can not delete
				writeError(w, "delete cannot be attempted because the snapshot has a link", http.StatusBadRequest)
				return
			}

			//all checks done: volume exists, snapshot existing without links -> it can be deleted
			delete(snapIDtoSnap, SnapID)
			Data.VolumeIDToVolume[source].SnapSource = false
			newMockJob(jobID, types.JobStatusRunning, types.JobStatusSucceeded, resourceLink)
		}
	}
	returnJobByID(w, jobID)
}

// RenameSnapshot - Renames a snapshot and updates mock cache
func RenameSnapshot(w http.ResponseWriter, r *http.Request, sourceVolumeList []types.VolumeList, executionOption, oldSnapID, newSnapID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	renameSnapshot(w, r, sourceVolumeList, executionOption, oldSnapID, newSnapID)
}

func renameSnapshot(w http.ResponseWriter, r *http.Request, sourceVolumeList []types.VolumeList, executionOption, oldSnapID, newSnapID string) {
	if fewVolumeUnavalaible(sourceVolumeList) {
		writeError(w, "few devices not available", http.StatusBadRequest)
		return
	}
	// Make a job to return
	resourceLink := fmt.Sprintf("/replication/symmetrix/%s/snapshot/%s", DefaultSymmetrixID, oldSnapID)
	jobID := fmt.Sprintf("SnapID-%d", time.Now().Nanosecond())
	if InducedErrors.JobFailedError {
		newMockJob(jobID, types.JobStatusRunning, types.JobStatusFailed, resourceLink)
	} else {
		for _, volID := range sourceVolumeList {
			if Data.VolIDToSnapshots[volID.Name][oldSnapID] == nil {
				writeError(w, "no snapshot information, Snapshot cannot be found", http.StatusBadRequest)
				return
			}
			for _, snap := range Data.VolIDToSnapshots[volID.Name] {
				if snap.Name == oldSnapID {
					snap.Name = newSnapID
					Data.VolIDToSnapshots[volID.Name] = map[string]*types.Snapshot{newSnapID: snap}
					newMockJob(jobID, types.JobStatusRunning, types.JobStatusSucceeded, resourceLink)
					Data.VolIDToSnapshots[volID.Name] = map[string]*types.Snapshot{newSnapID: snap}
					newMockJob(jobID, types.JobStatusRunning, types.JobStatusSucceeded, resourceLink)
				}
			}
		}
		returnJobByID(w, jobID)
	}
}

// LinkSnapshot - Links a snapshot and updates a mock cache
func LinkSnapshot(w http.ResponseWriter, r *http.Request, sourceVolumeList []types.VolumeList, targetVolumeList []types.VolumeList, executionOption, SnapID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	linkSnapshot(w, r, sourceVolumeList, targetVolumeList, executionOption, SnapID)
}

func linkSnapshot(w http.ResponseWriter, r *http.Request, sourceVolumeList []types.VolumeList, targetVolumeList []types.VolumeList, executionOption, SnapID string) {
	if sourceVolumeList[0].Name == "" {
		writeError(w, "no source volume names given to link the snapshot", http.StatusBadRequest)
		return
	}
	if targetVolumeList[0].Name == "" {
		writeError(w, "no link volume names given to link the snapshot to", http.StatusBadRequest)
		return
	}
	if len(sourceVolumeList) != len(targetVolumeList) {
		writeError(w, "cannot link snapshot, the number of source and devices should be same", http.StatusBadRequest)
		return
	}
	if fewVolumeUnavalaible(sourceVolumeList) {
		writeError(w, "few source devices not available", http.StatusBadRequest)
		return
	}
	if fewVolumeUnavalaible(targetVolumeList) {
		writeError(w, "few target devices not available", http.StatusBadRequest)
		return
	}
	// Make a job to return
	resourceLink := fmt.Sprintf("/replication/symmetrix/%s/snapshot/%s", DefaultSymmetrixID, SnapID)
	jobID := fmt.Sprintf("SnapID-%d", time.Now().Nanosecond())

	if InducedErrors.JobFailedError {
		newMockJob(jobID, types.JobStatusRunning, types.JobStatusFailed, resourceLink)
	} else {
		for key, volID := range sourceVolumeList {
			snapIDtoSnap := Data.VolIDToSnapshots[volID.Name]
			targetVolID := targetVolumeList[key].Name
			if snapIDtoSnap[SnapID] == nil {
				writeError(w, "no snapshot information, snopshot cannot be found on this device", http.StatusBadRequest)
				return
			}
			//all devices exist, #source=#target, snapshot exist, check if target already linked
			snapIDtoLinkedVolKey := SnapID + ":" + volID.Name
			volIDToLinkedVols := Data.SnapIDToLinkedVol[snapIDtoLinkedVolKey]
			if volIDToLinkedVols == nil {
				//No Linked Volume, first link request for this SnapID
				volIDToLinkedVols = map[string]*types.LinkedVolumes{}
			} else {
				//snapshot is linked to few devices, check if target is already linked
				if !(volIDToLinkedVols[targetVolID] == nil) {
					//duplicate link request
					writeError(w, "devices already in desired state", http.StatusBadRequest)
					return
				}
			}
			//all devices exist, #source=#target, snapshot exist, target is not linked -> ideal for Linking
			time := time.Now().Nanosecond()
			linkedVolume := &types.LinkedVolumes{
				TargetDevice: targetVolID,
				Timestamp:    strconv.Itoa(time),
				State:        "Linked",
				Copy:         false,
				Restored:     false,
				Linked:       true,
				Defined:      true,
			}
			if InducedErrors.TargetNotDefinedError {
				linkedVolume.Defined = false
			}

			volIDToLinkedVols[targetVolID] = linkedVolume
			Data.SnapIDToLinkedVol[snapIDtoLinkedVolKey] = volIDToLinkedVols
			Data.VolumeIDToVolume[targetVolID].SnapTarget = true
			newMockJob(jobID, types.JobStatusRunning, types.JobStatusSucceeded, resourceLink)
		}
	}
	returnJobByID(w, jobID)
}

// UnlinkSnapshot - Unlinks a snapshot and updates mock cache
func UnlinkSnapshot(w http.ResponseWriter, r *http.Request, sourceVolumeList []types.VolumeList, targetVolumeList []types.VolumeList, executionOption, SnapID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	unlinkSnapshot(w, r, sourceVolumeList, targetVolumeList, executionOption, SnapID)
}

func unlinkSnapshot(w http.ResponseWriter, r *http.Request, sourceVolumeList []types.VolumeList, targetVolumeList []types.VolumeList, executionOption, SnapID string) {
	if sourceVolumeList[0].Name == "" {
		writeError(w, "no source volume names given to unlink the snapshot", http.StatusBadRequest)
		return
	}
	if targetVolumeList[0].Name == "" {
		writeError(w, "no target volume names given to unlink the snapshot to", http.StatusBadRequest)
		return
	}
	if len(sourceVolumeList) != len(targetVolumeList) {
		writeError(w, "cannot unlink snapshot, the number of source and devices should be same", http.StatusBadRequest)
		return
	}
	if fewVolumeUnavalaible(sourceVolumeList) {
		writeError(w, "few source devices not available", http.StatusBadRequest)
		return
	}
	if fewVolumeUnavalaible(targetVolumeList) {
		writeError(w, "few target devices not available", http.StatusBadRequest)
		return
	}
	// Make a job to return
	resourceLink := fmt.Sprintf("/replication/symmetrix/%s/snapshot/%s", DefaultSymmetrixID, SnapID)
	jobID := fmt.Sprintf("SnapID-%d", time.Now().Nanosecond())

	if InducedErrors.JobFailedError {
		newMockJob(jobID, types.JobStatusRunning, types.JobStatusFailed, resourceLink)
	} else {
		for key, volID := range sourceVolumeList {
			snapIDtoSnap := Data.VolIDToSnapshots[volID.Name]
			targetVolID := targetVolumeList[key].Name
			if snapIDtoSnap[SnapID] == nil {
				writeError(w, "no snapshot information, snopshot cannot be found on this device", http.StatusBadRequest)
				return
			}
			//all devices exist, #source=#target, snapshot exist, check if source is linked to target
			snapIDtoLinkedVolKey := SnapID + ":" + volID.Name
			volIDToLinkedVolumes := Data.SnapIDToLinkedVol[snapIDtoLinkedVolKey]
			if _, ok := volIDToLinkedVolumes[targetVolID]; ok {
				//source volume is linked to target, ideal for unlink
				delete(volIDToLinkedVolumes, targetVolID)
				volIDToLinkedVolumes = Data.SnapIDToLinkedVol[snapIDtoLinkedVolKey]
				Data.VolumeIDToVolume[targetVolID].SnapTarget = false
				newMockJob(jobID, types.JobStatusRunning, types.JobStatusSucceeded, resourceLink)
			} else {
				//already unlinked
				writeError(w, "devices already in desired state", http.StatusBadRequest)
				return
			}
		}
	}
	returnJobByID(w, jobID)
}

// check if all the devices exist in the Mock VolumeIDToVolume or check if any unvailable devices
func fewVolumeUnavalaible(sourceVolumeList []types.VolumeList) bool {
	for _, volID := range sourceVolumeList {
		if Data.VolumeIDToVolume[volID.Name] == nil {
			return true
		}
	}
	return false
}

// returns true for Snapshot Creation if a snpshot with same name already there, false otherwise
func duplicateSnapshotCreationRequest(source, SnapID string) bool {
	_, ok := Data.VolIDToSnapshots[source][SnapID]
	return ok
}

// GET univmax/restapi/private/APIVersion/replication/symmetrix/{symid}/volume
func handleSymVolumes(w http.ResponseWriter, r *http.Request) {
	if InducedErrors.GetSymVolumeError {
		writeError(w, "error fetching the list: induced error", http.StatusBadRequest)
		return
	}
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	queryParams := r.URL.Query()
	symVolumeList := new(types.SymVolumeList)
	if details := queryParams.Get("includeDetails"); details == "true" {
		for key, snapshots := range Data.VolIDToSnapshots {
			symVolumeList.Name = append(symVolumeList.Name, key)
			var snapList []types.Snapshot
			for _, snap := range snapshots {
				snapshotName := fmt.Sprintf("%s-SRC-%s-%d", symVolumeList.Name[0], snap.Name, snap.Generation)
				if InducedErrors.InvalidSnapshotName {
					snapshotName = "InvalidSnapshot"
				}
				snapshot := types.Snapshot{
					Name:       snapshotName,
					Generation: snap.Generation,
					Timestamp:  snap.Timestamp,
					State:      snap.State,
				}
				snapList = append(snapList, snapshot)
			}
			symDevice := types.SymDevice{
				SymmetrixID: DefaultSymmetrixID,
				Name:        key,
				Snapshot:    snapList,
			}
			symVolumeList.SymDevice = append(symVolumeList.SymDevice, symDevice)
		}
	} else {
		for key := range Data.VolIDToSnapshots {
			symVolumeList.Name = append(symVolumeList.Name, key)
		}
	}
	writeJSON(w, symVolumeList)
}

// GET univmax/restapi/private/APIVersion/replication/symmetrix/{symid}/volume/{volID}/snapshot/
// GET univmax/restapi/private/APIVersion/replication/symmetrix/{symid}/volume/{volID}/snapshot/{SnapID}
func handleVolSnaps(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volID := vars["volID"]
	SnapID := vars["SnapID"]
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	if InducedErrors.GetVolSnapsError {
		writeError(w, "error fetching the Snapshot Info: induced error", http.StatusBadRequest)
		return
	}
	if Data.VolumeIDToVolume[volID] == nil {
		writeError(w, "Volume cannot be found: "+volID, http.StatusNotFound)
		return
	}

	volumeSnapshotSource, _ := returnSnapshotObjectList(volID)
	volumeSnapshotLink := returnVolumeSnapshotLink(volID)

	if SnapID == "" {
		// Both Volume Snapshots exist
		// for /{symid}/volume/{volID}/snapshot/
		snaphotVolumeGeneration := new(types.SnapshotVolumeGeneration)
		snaphotVolumeGeneration.DeviceName = volID
		snaphotVolumeGeneration.VolumeSnapshotSource = volumeSnapshotSource
		snaphotVolumeGeneration.VolumeSnapshotLink = volumeSnapshotLink
		writeJSON(w, snaphotVolumeGeneration)
	} else {
		// Both Volume Snapshots exist
		// for /{symid}/volume/{volID}/snapshot/{SnapID}
		volumeSnapshot := new(types.VolumeSnapshot)
		volumeSnapshot.DeviceName = volID
		volumeSnapshot.SnapshotName = SnapID
		for _, snapSrc := range volumeSnapshotSource {
			if snapSrc.SnapshotName == SnapID {
				volumeSnapshot.VolumeSnapshotSource = append(volumeSnapshot.VolumeSnapshotSource, types.VolumeSnapshotSource{
					SnapshotName: snapSrc.SnapshotName,
					Generation:   snapSrc.Generation,
					TimeStamp:    snapSrc.TimeStamp,
					State:        snapSrc.State,
				})
			}
		}
		volumeSnapshot.VolumeSnapshotLink = volumeSnapshotLink
		writeJSON(w, volumeSnapshot)
	}
}

// returns the List of VolumesSnapshot objects derived based on existing mock Snapshot object
func returnSnapshotObjectList(volID string) ([]types.VolumeSnapshotSource, []int64) {
	var volumeSnapshotSrc []types.VolumeSnapshotSource
	var generations []int64
	for _, snap := range Data.VolIDToSnapshots[volID] {
		snapshotSrc := types.VolumeSnapshotSource{
			SnapshotName:  snap.Name,
			Generation:    snap.Generation,
			TimeStamp:     snap.Timestamp,
			State:         snap.State,
			LinkedVolumes: returnLinkedVolumes(snap.Name + ":" + volID),
		}
		if InducedErrors.SnapshotExpired {
			snapshotSrc.Expired = true
		}
		volumeSnapshotSrc = append(volumeSnapshotSrc, snapshotSrc)
		generations = append(generations, snap.Generation)
	}

	return volumeSnapshotSrc, generations
}

// returns the List of Linked Volumes to Snapshots of a volume
func returnLinkedVolumes(snapIDtoLinkedVolKey string) []types.LinkedVolumes {
	var linkedVolumes []types.LinkedVolumes
	for _, volume := range Data.SnapIDToLinkedVol[snapIDtoLinkedVolKey] {
		linkedVolumes = append(linkedVolumes, *volume)
	}
	return linkedVolumes
}

// returns the List of volumeSnapshotLink to a Snapshot
func returnVolumeSnapshotLink(targetVolID string) []types.VolumeSnapshotLink {
	var snapshotLnk []types.VolumeSnapshotLink
	for _, volume := range Data.SnapIDToLinkedVol {
		if target, ok := volume[targetVolID]; ok {
			snapshotLnk = append(snapshotLnk, types.VolumeSnapshotLink{
				TargetDevice:     target.TargetDevice,
				Timestamp:        target.Timestamp,
				State:            target.State,
				TrackSize:        target.TrackSize,
				Tracks:           target.Tracks,
				PercentageCopied: target.PercentageCopied,
				Linked:           target.Linked,
				Restored:         target.Restored,
				Defined:          target.Defined,
				Copy:             target.Copy,
				Destage:          target.Destage,
				Modified:         target.Modified,
			})
		}
	}
	return snapshotLnk
}

// GET univmax/restapi/private/APIVersion/replication/symmetrix/{symid}/volume/{volID}/snapshot/{SnapID}/generation
// GET univmax/restapi/private/APIVersion/replication/symmetrix/{symid}/volume/{volID}/snapshot/{SnapID}/generation/{genID}
func handleGenerations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	volID := vars["volID"]
	SnapID := vars["SnapID"]
	genID := vars["genID"]
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	if Data.VolumeIDToVolume[volID] == nil {
		writeError(w, "Volume cannot be found: "+volID, http.StatusNotFound)
		return
	}

	volumeSnapshotSource, generations := returnSnapshotObjectList(volID)
	volumeSnapshotLink := returnVolumeSnapshotLink(volID)

	if genID == "" {
		// Both Volume Snapshots exist
		// for /{symid}/volume/{volID}/snapshot/{SnapID}/generation/
		volumeSnapshotGenerations := new(types.VolumeSnapshotGenerations)
		volumeSnapshotGenerations.DeviceName = volID
		volumeSnapshotGenerations.Generation = generations
		volumeSnapshotGenerations.SnapshotName = SnapID
		volumeSnapshotGenerations.VolumeSnapshotSource = volumeSnapshotSource
		volumeSnapshotGenerations.VolumeSnapshotLink = volumeSnapshotLink
		writeJSON(w, volumeSnapshotGenerations)
		return
	}
	// Both Volume Snapshots exist
	// for /{symid}/volume/{volID}/snapshot/{SnapID}/generation/{genID}
	volumeSnapshotGeneration := new(types.VolumeSnapshotGeneration)
	volumeSnapshotGeneration.DeviceName = volID
	volumeSnapshotGeneration.SnapshotName = SnapID
	volumeSnapshotGeneration.VolumeSnapshotLink = volumeSnapshotLink
	// volumeSnapshotGeneration.VolumeSnapshotSource = returnSnapshotGenerationInfo(volID, SnapID, genID)
	gID, _ := strconv.ParseInt(genID, 10, 64)
	for _, snapSrc := range volumeSnapshotSource {
		if snapSrc.SnapshotName == SnapID && snapSrc.Generation == gID {
			volumeSnapshotGeneration.VolumeSnapshotSource = snapSrc
			volumeSnapshotGeneration.Generation = snapSrc.Generation
			break
		}
	}
	writeJSON(w, volumeSnapshotGeneration)
	return
}

func handleCapabilities(w http.ResponseWriter, r *http.Request) {
	var jsonBytes []byte
	if InducedErrors.SnapshotNotLicensed {
		jsonBytes = []byte("{\"symmetrixCapability\":[{\"symmetrixId\":\"000197900046\",\"snapVxCapable\":false,\"rdfCapable\":true,\"virtualWitnessCapable\":false}]}")
	} else if InducedErrors.InvalidResponse {
		writeError(w, "something went wrong: induced error", http.StatusBadRequest)
		return
	} else if InducedErrors.UnisphereMismatchError {
		jsonBytes = []byte("{\"symmetrixCapability\":[{\"symmetrixId\":\"000000000000\",\"snapVxCapable\":true,\"rdfCapable\":true,\"virtualWitnessCapable\":false}]}")
	} else {
		jsonBytes = []byte("{\"symmetrixCapability\":[{\"symmetrixId\":\"000197900046\",\"snapVxCapable\":true,\"rdfCapable\":true,\"virtualWitnessCapable\":false}]}")
	}
	_, err := w.Write(jsonBytes)
	if err != nil {
		log.Printf("Couldn't write to ResponseWriter")
		w.WriteHeader(http.StatusInternalServerError)
	}
	return
}

func handlePrivVolume(w http.ResponseWriter, r *http.Request) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	if InducedErrors.GetPrivVolumeByIDError {
		writeError(w, "error fetching the Volume structure: induced error", http.StatusBadRequest)
		return
	}
	queryParams := r.URL.Query()
	privateVolumeIterator := new(types.PrivVolumeIterator)
	if wwn := queryParams.Get("wwn"); wwn != "" {
		volID := wwn[27:]
		volume := Data.VolumeIDToVolume[volID]
		volumeHeader := parseVolumetoVolumeHeader(volume)
		timeFinderInfo := returnTimeFinderInfo(volID)
		var result []types.VolumeResultPrivate
		result = append(result, types.VolumeResultPrivate{
			VolumeHeader:   *volumeHeader,
			TimeFinderInfo: *timeFinderInfo,
		})
		privVolumeResultList := types.PrivVolumeResultList{
			PrivVolumeList: result,
			From:           1,
			To:             1,
		}
		privateVolumeIterator.ResultList = privVolumeResultList
		privateVolumeIterator.ID = "70e15d35-baaf-43d3-865a-bf3300684895_0"
		privateVolumeIterator.ExpirationTime = 1576137450163
		privateVolumeIterator.MaxPageSize = 1000
		privateVolumeIterator.Count = 1
	}
	writeJSON(w, privateVolumeIterator)
}

func parseVolumetoVolumeHeader(volume *types.Volume) *types.VolumeHeader {
	volumeHeader := &types.VolumeHeader{
		VolumeID:     volume.VolumeID,
		CapGB:        volume.CapacityGB,
		CapMB:        volume.FloatCapacityMB,
		Status:       volume.Status,
		SSID:         volume.SSID,
		EffectiveWWN: volume.WWN,
		Encapsulated: volume.Encapsulated,
	}

	return volumeHeader
}

func returnTimeFinderInfo(volID string) *types.TimeFinderInfo {
	timeFinder := new(types.TimeFinderInfo)
	if _, ok := Data.VolIDToSnapshots[volID]; ok {
		timeFinder.SnapVXSrc = ok
	}
	for _, volIDToLinkedVols := range Data.SnapIDToLinkedVol {
		if _, ok := volIDToLinkedVols[volID]; ok {
			timeFinder.SnapVXTgt = ok
			break
		}
	}
	if timeFinder.SnapVXSrc || timeFinder.SnapVXTgt {
		timeFinder.SnapVXSession = append(timeFinder.SnapVXSession, returnSnapVXSession(volID, timeFinder.SnapVXSrc, timeFinder.SnapVXTgt))
	}
	return timeFinder
}

func returnSnapVXSession(volID string, isSource, isTarget bool) types.SnapVXSession {
	var snapVXSession types.SnapVXSession
	if isSource {
		snapVXSession.SourceSnapshotGenInfo = returnSrcSnapshotGenInfo(volID)
	}

	if isTarget {
		for snapIDtoLinkedVolKey, volIDToLinkedVolumes := range Data.SnapIDToLinkedVol {
			sourceVolID := strings.Split(snapIDtoLinkedVolKey, ":")[1]
			SnapID := strings.Split(snapIDtoLinkedVolKey, ":")[0]
			if _, ok := volIDToLinkedVolumes[volID]; ok {
				snapVXSession.TargetSourceSnapshotGenInfo = &types.TargetSourceSnapshotGenInfo{
					TargetDevice: volID,
					SourceDevice: sourceVolID,
					SnapshotName: SnapID,
				}
			}
		}
	}
	return snapVXSession
}

func returnSrcSnapshotGenInfo(volID string) []types.SourceSnapshotGenInfo {
	var srcSnapGenInfo []types.SourceSnapshotGenInfo

	for _, snapIDtoSnap := range Data.VolIDToSnapshots[volID] {
		timestamp, _ := strconv.ParseInt(snapIDtoSnap.Timestamp, 10, 64)
		srcSnapGenInfo = append(srcSnapGenInfo, types.SourceSnapshotGenInfo{
			SnapshotHeader: types.SnapshotHeader{
				Device:       volID,
				SnapshotName: snapIDtoSnap.Name,
				Generation:   snapIDtoSnap.Generation,
				Timestamp:    timestamp,
			},
		})
	}

	return srcSnapGenInfo
}

// /univmax/restapi/100/sloprovisioning/symmetrix/{symid}/hostgroup/{id}
// /univmax/restapi/100/sloprovisioning/symmetrix/{symid}/hostgroup
func handleHostGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hostGroupID := vars["id"]
	switch r.Method {

	case http.MethodGet:
		if InducedErrors.GetHostGroupError {
			writeError(w, "Error retrieving HostGroup: induced error", http.StatusRequestTimeout)
			return
		}
		returnHostGroup(w, hostGroupID)

	case http.MethodPost:
		if InducedErrors.CreateHostGroupError {
			writeError(w, "Error creating HostGroup: induced error", http.StatusRequestTimeout)
			return
		}
		decoder := json.NewDecoder(r.Body)
		createHostGroupParam := &types.CreateHostGroupParam{}
		err := decoder.Decode(createHostGroupParam)
		if err != nil {
			writeError(w, "InvalidJson", http.StatusBadRequest)
			return
		}
		AddHostGroup(createHostGroupParam.HostGroupID, createHostGroupParam.HostIDs, createHostGroupParam.HostFlags)
		ReturnHostGroup(w, createHostGroupParam.HostGroupID)

	case http.MethodPut:
		if hasError(&InducedErrors.UpdateHostGroupError) {
			writeError(w, "Error updating HostGroup: induced error", http.StatusRequestTimeout)
			return
		}
		decoder := json.NewDecoder(r.Body)
		updateHostGroupParam := &types.UpdateHostGroupParam{}
		err := decoder.Decode(updateHostGroupParam)
		if err != nil {
			writeError(w, "InvalidJson", http.StatusBadRequest)
			return
		}
		UpdateHostGroupFromParams(hostGroupID, updateHostGroupParam)
		ReturnHostGroup(w, hostGroupID)

	case http.MethodDelete:
		if InducedErrors.DeleteHostGroupError {
			writeError(w, "Error deleting HostGroup: induced error", http.StatusRequestTimeout)
			return
		}
		RemoveHostGroup(hostGroupID)
	default:
		writeError(w, "Invalid Method", http.StatusBadRequest)
	}
}

// ReturnHostGroup - Returns a hostGroup from cache
func ReturnHostGroup(w http.ResponseWriter, hostGroupID string) {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	returnHostGroup(w, hostGroupID)
}

func returnHostGroup(w http.ResponseWriter, hostGroupID string) {
	if hostGroupID != "" {
		if hostGroup, ok := Data.HostGroupIDToHostGroup[hostGroupID]; ok {
			writeJSON(w, hostGroup)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}
}

// AddHostGroup - Adds a host group to the mock data cache
func AddHostGroup(hostGroupID string, hostIDs []string, hostFlags *types.HostFlags) (*types.HostGroup, error) {
	if _, ok := Data.HostGroupIDToHostGroup[hostGroupID]; ok {
		return nil, errors.New("error! Host Group already exists")
	}
	newHostGroup(hostGroupID, hostIDs, hostFlags)
	return Data.HostGroupIDToHostGroup[hostGroupID], nil
}

func newHostGroup(hostGroupID string, hostIDs []string, hostFlags *types.HostFlags) {
	hostSummaries := []types.HostSummary{}

	for _, hostID := range hostIDs {
		Host := types.HostSummary{
			HostID: hostID,
		}
		hostSummaries = append(hostSummaries, Host)
	}

	hostGroup := &types.HostGroup{
		HostGroupID: hostGroupID,
		Hosts:       hostSummaries,
	}

	if hostFlags != nil {
		handleFlags(hostGroup, hostFlags)
	}

	Data.HostGroupIDToHostGroup[hostGroupID] = hostGroup
}

// RemoveHostGroup - Removes hostGroup from mock cache
func RemoveHostGroup(hostGroupID string) error {
	mockCacheMutex.Lock()
	defer mockCacheMutex.Unlock()
	return removeHostGroup(hostGroupID)
}

// removeHostGroup - Remove a hostGroup from the mock data cache
func removeHostGroup(hostGroupID string) error {
	_, ok := Data.HostGroupIDToHostGroup[hostGroupID]
	if !ok {
		return errors.New("error! Host doesn't exist")
	}
	Data.HostGroupIDToHostGroup[hostGroupID] = nil
	return nil
}

// UpdateHostGroupFromParams - Updates HostGroup given an UpdateHostGroupParam payload
func UpdateHostGroupFromParams(hostGroupID string, updateParams *types.UpdateHostGroupParam) {
	updateHostGroup(hostGroupID, updateParams.EditHostGroupAction) // #nosec G20
}

// updateHostGroup - Update HostGroup
func updateHostGroup(hostGroupID string, editPayload *types.EditHostGroupActionParams) (*types.HostGroup, error) {
	hostGroup, ok := Data.HostGroupIDToHostGroup[hostGroupID]
	if !ok {
		return nil, fmt.Errorf("error! HostGroup %s does not exist", hostGroupID)
	}

	if editPayload.RemoveHostParam != nil {
		hostSummaries := []types.HostSummary{}
		for _, host := range hostGroup.Hosts {
			if !stringInSlice(host.HostID, editPayload.RemoveHostParam.Host) {
				hostSummary := types.HostSummary{
					HostID: host.HostID,
				}
				hostSummaries = append(hostSummaries, hostSummary)
			}
		}
		hostGroup.Hosts = hostSummaries
	}

	if editPayload.AddHostParam != nil {
		for _, host := range editPayload.AddHostParam.Host {
			hostSummary := types.HostSummary{
				HostID: host,
			}
			hostGroup.Hosts = append(hostGroup.Hosts, hostSummary)
		}
	}

	if editPayload.SetHostGroupFlags != nil {
		handleFlags(hostGroup, editPayload.SetHostGroupFlags.HostFlags)
	}

	if editPayload.RenameHostGroupParam != nil {
		hostGroupID = editPayload.RenameHostGroupParam.NewHostGroupName
		hostGroup.HostGroupID = hostGroupID
	}

	// Update the HostGroup mapping with the update HostGroup
	Data.HostGroupIDToHostGroup[hostGroupID] = hostGroup
	return hostGroup, nil
}

func handleFlags(hostGroup *types.HostGroup, flagPayload *types.HostFlags) {
	var enabledFlags, disabledFlags []string
	if flagPayload.VolumeSetAddressing.Override {
		if flagPayload.VolumeSetAddressing.Enabled {
			enabledFlags = append(enabledFlags, "Volume_Set_Addressing(V)")
		}
		disabledFlags = append(disabledFlags, "Volume_Set_Addressing(V)")
	}

	if flagPayload.AvoidResetBroadcast.Override {
		if flagPayload.AvoidResetBroadcast.Enabled {
			enabledFlags = append(enabledFlags, "Avoid_Reset_Broadcast(ARB)")
		}
		disabledFlags = append(disabledFlags, "Avoid_Reset_Broadcast(ARB)")
	}

	if flagPayload.DisableQResetOnUA.Override {
		if flagPayload.DisableQResetOnUA.Enabled {
			enabledFlags = append(enabledFlags, "Disable_Q_Reset_on_UA(D)")
		}
		disabledFlags = append(disabledFlags, "Disable_Q_Reset_on_UA(D)")
	}

	if flagPayload.EnvironSet.Override {
		if flagPayload.EnvironSet.Enabled {
			enabledFlags = append(enabledFlags, "Environ_Set(E)")
		}
		disabledFlags = append(disabledFlags, "Environ_Set(E)")
	}

	if flagPayload.OpenVMS.Override {
		if flagPayload.OpenVMS.Enabled {
			enabledFlags = append(enabledFlags, "OpenVMS(OVMS)")
		}
		disabledFlags = append(disabledFlags, "OpenVMS(OVMS)")
	}

	if flagPayload.SCSI3.Override {
		if flagPayload.SCSI3.Enabled {
			enabledFlags = append(enabledFlags, "SCSI_3(SC3)")
		}
		disabledFlags = append(disabledFlags, "SCSI_3(SC3)")
	}

	if flagPayload.SCSISupport1.Override {
		if flagPayload.SCSISupport1.Enabled {
			enabledFlags = append(enabledFlags, "SCSI_Support1(OS2007)")
		}
		disabledFlags = append(disabledFlags, "SCSI_Support1(OS2007)")
	}

	if flagPayload.Spc2ProtocolVersion.Override {
		if flagPayload.Spc2ProtocolVersion.Enabled {
			enabledFlags = append(enabledFlags, "SPC2_Protocol_Version(SPC2)")
		}
		disabledFlags = append(disabledFlags, "SPC2_Protocol_Version(SPC2)")
	}

	enabledFlag := strings.Join(enabledFlags, ",")
	disabledFlag := strings.Join(disabledFlags, ",")

	hostGroup.EnabledFlags = enabledFlag
	hostGroup.DisabledFlags = disabledFlag

	if flagPayload.ConsistentLUN {
		hostGroup.ConsistentLun = true
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
