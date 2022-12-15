package v100

type StorageGroupMetricsParam struct {
	SymmetrixId    string   `json:"symmetrixId"`
	StartDate      int64    `json:"startDate"`
	EndDate        int64    `json:"endDate"`
	DataFormat     string   `json:"dataFormat"`
	StorageGroupId string   `json:"storageGroupId"`
	Metrics        []string `json:"metrics"`
}

type StorageGroupMetricsIterator struct {
	ResultList     StorageGroupMetricsResultList `json:"resultList"`
	Id             string                        `json:"id"`
	Count          int                           `json:"count"`
	ExpirationTime int64                         `json:"expirationTime"`
	MaxPageSize    int                           `json:"maxPageSize"`
}

type StorageGroupMetricsResultList struct {
	Result []StorageGroupMetric `json:"result"`
	From   int                  `json:"from"`
	To     int                  `json:"to"`
}

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

type VolumeMetricsParam struct {
	SystemId                       string   `json:"systemId"`
	StartDate                      int64    `json:"startDate"`
	EndDate                        int64    `json:"endDate"`
	DataFormat                     string   `json:"dataFormat"`
	CommaSeparatedStorageGroupList string   `json:"commaSeparatedStorageGroupList"`
	Metrics                        []string `json:"metrics"`
}

type VolumeMetricsIterator struct {
	ResultList     VolumeMetricsResultList `json:"resultList"`
	Id             string                  `json:"id"`
	Count          int                     `json:"count"`
	ExpirationTime int64                   `json:"expirationTime"`
	MaxPageSize    int                     `json:"maxPageSize"`
}

type VolumeMetricsResultList struct {
	Result []VolumeResult `json:"result"`
	From   int            `json:"from"`
	To     int            `json:"to"`
}

type VolumeResult struct {
	VolumeResult  []VolumeMetric `json:"volumeResult"`
	VolumeId      string         `json:"volumeId"`
	StorageGroups string         `json:"storageGroups"`
}

type VolumeMetric struct {
	MBRead            float64 `json:"MBRead"`
	MBWritten         float64 `json:"MBWritten"`
	Reads             float64 `json:"Reads"`
	Writes            float64 `json:"Writes"`
	ReadResponseTime  float64 `json:"ReadResponseTime"`
	WriteResponseTime float64 `json:"WriteResponseTime"`
	Timestamp         int64   `json:"timestamp"`
}
