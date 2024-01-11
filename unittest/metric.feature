Feature: PMAX metrics test

  Scenario Outline: Test GetStorageGroupMetrics
  Given a valid connection
  And I have an allowed list of <arrays>
  And I induce error <induced>
  When I call GetStorageGroupMetrics
  Then the error message contains <errormsg>
  And I get StorageGroupMetrics

  Examples:
  | arrays           | induced                        | errormsg                          |
  | "000000000000"   | "none"                         | "ignored as it is not managed"    |
  | "000197900046"   | "GetStorageGroupMetricsError"  | "induced error"                   |
  | "000197900046"   | "none"                         | "none"                            |

  Scenario Outline: Test GetVolumesMetrics
  Given a valid connection
  And I have an allowed list of <arrays>
  And I induce error <induced>
  When I call GetVolumesMetrics
  Then the error message contains <errormsg>
  And I get VolumesMetrics

  Examples:
  | arrays           | induced                        | errormsg                          |
  | "000000000000"   | "none"                         | "ignored as it is not managed"    |
  | "000197900046"   | "GetVolumesMetricsError"       | "induced error"                   |
  | "000197900046"   | "none"                         | "none"                            |

  Scenario Outline: Test GetStorageGroupPerfKeys
  Given a valid connection
  And I have an allowed list of <arrays>
  And I induce error <induced>
  When I call GetStorageGroupPerfKeys
  Then the error message contains <errormsg>
  And I get StorageGroupPerfKeys

  Examples:
  | arrays           | induced                        | errormsg                          |
  | "000000000000"   | "none"                         | "ignored as it is not managed"    |
  | "000197900046"   | "GetStorageGroupPerfKeyError"  | "induced error"                   |
  | "000197900046"   | "none"                         | "none"                            |


  Scenario Outline: Test GetArrayPerfKeys
  Given a valid connection
  And I induce error <induced>
  When I call GetArrayPerfKeys
  Then the error message contains <errormsg>
  And I get ArrayPerfKeys

  Examples:
  | arrays           | induced                        | errormsg                          |
  | "000197900046"   | "GetArrayPerfKeyError"         | "induced error"                   |
  | "000197900046"   | "none"                         | "none"                            |

  @this
  Scenario Outline: Test GetVolumesMetricsByID
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call GetVolumesMetricsByID for "vol1"
    Then the error message contains <errormsg>
    And I get VolumesMetrics

    Examples:
      | arrays           | induced                        | errormsg                          |
      | "000000000000"   | "none"                         | "ignored as it is not managed"    |
      | "000197900046"   | "GetVolumesMetricsError"         | "induced error"                   |
      | "000197900046"   | "none"                         | "none"                            |

  @this
  Scenario Outline: Test GetFileSystemMetricsByID
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call GetFileSystemMetricsByID for "file1"
    Then the error message contains <errormsg>
    And I get FileMetrics

    Examples:
      | arrays           | induced                        | errormsg                          |
      | "000000000000"   | "none"                         | "ignored as it is not managed"    |
      | "000197900046"   | "GetFileSysMetricsError"         | "induced error"                   |
      | "000197900046"   | "none"                         | "none"                            |