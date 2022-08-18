package v100

//MigrationEnv related data types
type MigrationEnv struct {
	ArrayId               string `json:"arrayId"`
	StorageGroupCount     int    `json:"storageGroupCount"`
	MigrationSessionCount int    `json:"migrationSessionCount"`
	Local                 bool   `json:"local"`
}

//MigrationSession contains information about device pairs, source and target masking views
type MigrationSession struct {
	SourceArray       string                 `json:"sourceArray"`
	TargetArray       string                 `json:"targetArray"`
	StorageGroup      string                 `json:"storageGroup"`
	State             string                 `json:"state"`
	TotalCapacity     float64                `json:"totalCapacity"`
	RemainingCapacity float64                `json:"remainingCapacity"`
	DevicePairs       []MigrationDevicePairs `json:"devicePairs"`
	SourceMaskingView []SourceMaskingView    `json:"sourceMaskingView"`
	TargetMaskingView []TargetMaskingView    `json:"targetMaskingView"`
	Offline           bool                   `json:"offline"`
	Type              string                 `json:"type"`
}

type ModifyMigrationSessionRequest struct {
	Action          string `json:"action"`
	ExecutionOption string `json:"executionOption"`
}

type CreateMigrationEnv struct {
	OtherArrayId    string `json:"otherArrayId"`
	ExecutionOption string `json:"executionOption"`
}

type MigrationDevicePairs struct {
	SrcVolumeName string `json:"srcVolumeName"`
	InvalidSrc    bool   `json:"invalidSrc"`
	MissingSrc    bool   `json:"missingSrc"`
	TgtVolumeName string `json:"tgtVolumeName"`
	InvalidTgt    bool   `json:"invalidTgt"`
	MissingTgt    bool   `json:"missingTgt"`
}

type SourceMaskingView struct {
	Name           string           `json:"name"`
	Invalid        bool             `json:"invalid"`
	childInvalid   bool             `json:"childInvalid"`
	Missing        bool             `json:"missing"`
	InitiatorGroup []InitiatorGroup `json:"initiatorGroup"`
	PortGroup      []PortGroups     `json:"portGroup"`
}

type TargetMaskingView struct {
	Name           string           `json:"name"`
	Invalid        bool             `json:"invalid"`
	childInvalid   bool             `json:"childInvalid"`
	Missing        bool             `json:"missing"`
	InitiatorGroup []InitiatorGroup `json:"initiatorGroup"`
	PortGroup      []PortGroups     `json:"portGroup"`
}

type InitiatorGroup struct {
	Name         string       `json:"name"`
	Invalid      bool         `json:"invalid"`
	childInvalid bool         `json:"childInvalid"`
	Missing      bool         `json:"missing"`
	Initiator    []Initiators `json:"initiator"`
}

type Initiators struct {
	Name    string `json:"name"`
	WWN     string `json:"wwn"`
	Invalid bool   `json:"invalid"`
}

type PortGroups struct {
	Name         string  `json:"name"`
	Invalid      bool    `json:"invalid"`
	childInvalid bool    `json:"childInvalid"`
	Missing      bool    `json:"missing"`
	Ports        []Ports `json:"ports"`
}

type Ports struct {
	Name    string `json:"name"`
	Invalid bool   `json:"invalid"`
}
