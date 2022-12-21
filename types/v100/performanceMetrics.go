package v100

// StorageGroupMetricsParam parameters for query
type StorageGroupMetricsParam struct {
	SymmetrixID    string   `json:"symmetrixId"`
	StartDate      int64    `json:"startDate"`
	EndDate        int64    `json:"endDate"`
	DataFormat     string   `json:"dataFormat"`
	StorageGroupID string   `json:"storageGroupId"`
	Metrics        []string `json:"metrics"`
}

// StorageGroupMetricsIterator contains the result of query
type StorageGroupMetricsIterator struct {
	ResultList     StorageGroupMetricsResultList `json:"resultList"`
	ID             string                        `json:"id"`
	Count          int                           `json:"count"`
	ExpirationTime int64                         `json:"expirationTime"`
	MaxPageSize    int                           `json:"maxPageSize"`
}

// StorageGroupMetricsResultList contains the list of storage group metrics
type StorageGroupMetricsResultList struct {
	Result []StorageGroupMetric `json:"result"`
	From   int                  `json:"from"`
	To     int                  `json:"to"`
}

// StorageGroupMetric is the struct of metric
type StorageGroupMetric struct {
	HostReads         float64 `json:"HostReads"`
	HostWrites        float64 `json:"HostWrites"`
	HostMBReads       float64 `json:"HostMBReads"`
	HostMBWritten     float64 `json:"HostMBWritten"`
	ReadResponseTime  float64 `json:"ReadResponseTime"`
	WriteResponseTime float64 `json:"WriteResponseTime"`
	AllocatedCapacity float64 `json:"AllocatedCapacity"`
	Timestamp         int64   `json:"timestamp"`
}

// VolumeMetricsParam parameters for query
type VolumeMetricsParam struct {
	SystemID                       string   `json:"systemId"`
	StartDate                      int64    `json:"startDate"`
	EndDate                        int64    `json:"endDate"`
	DataFormat                     string   `json:"dataFormat"`
	CommaSeparatedStorageGroupList string   `json:"commaSeparatedStorageGroupList"`
	Metrics                        []string `json:"metrics"`
}

// VolumeMetricsIterator contains the result of query
type VolumeMetricsIterator struct {
	ResultList     VolumeMetricsResultList `json:"resultList"`
	ID             string                  `json:"id"`
	Count          int                     `json:"count"`
	ExpirationTime int64                   `json:"expirationTime"`
	MaxPageSize    int                     `json:"maxPageSize"`
}

// VolumeMetricsResultList contains the list of volume result
type VolumeMetricsResultList struct {
	Result []VolumeResult `json:"result"`
	From   int            `json:"from"`
	To     int            `json:"to"`
}

// VolumeResult contains the list of volume metrics and ID of volume
type VolumeResult struct {
	VolumeResult  []VolumeMetric `json:"volumeResult"`
	VolumeID      string         `json:"volumeId"`
	StorageGroups string         `json:"storageGroups"`
}

// VolumeMetric is the struct of metric
type VolumeMetric struct {
	MBRead            float64 `json:"MBRead"`
	MBWritten         float64 `json:"MBWritten"`
	Reads             float64 `json:"Reads"`
	Writes            float64 `json:"Writes"`
	ReadResponseTime  float64 `json:"ReadResponseTime"`
	WriteResponseTime float64 `json:"WriteResponseTime"`
	Timestamp         int64   `json:"timestamp"`
}
