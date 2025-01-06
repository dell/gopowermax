Feature: PMAX SRDF test

  @srdf
  Scenario Outline: Create a storage-group with volumes and protect it mode: ASYNC
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have 5 volumes
    When I call CreateSGReplica with "ASYNC"
    Then the error message contains <errormsg>
    And then SG should be replicated
    And the volumes should "" be replicated

    Examples:
    |     induced     |            errormsg               |  arrays     |
    |     "none"      |              "none"               |      ""     |
    |  "InvalidJSON"  |       "invalid character"         |      ""     |
    |     "none"      |    "ignored as it is not managed" |  "ignored"  |
    | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Create a storage-group with volumes and protect it mode: METRO
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have 5 volumes
    When I call CreateSGReplica with "METRO"
    Then the error message contains <errormsg>
    And then SG should be replicated
    And the volumes should "" be replicated

    Examples:
      |     induced     |            errormsg               |  arrays     |
      |     "none"      |              "none"               |      ""     |
      |  "InvalidJSON"  |       "invalid character"         |      ""     |
      |     "none"      |    "ignored as it is not managed" |  "ignored"  |
      | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Get SRDF info about a storage group
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have 5 volumes
    When I call CreateSGReplica with "ASYNC"
    And I call GetStorageGroupRDFInfo
    Then the error message contains <errormsg>

    Examples:
    |     induced     |            errormsg               |  arrays     |
    |     "none"      |              "none"               |      ""     |
    |  "InvalidJSON"  |       "invalid character"         |      ""     |
    |     "none"      |    "ignored as it is not managed" |  "ignored"  |
    | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Get RDF device-pair-info on a volume
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have 1 volumes
    When I call CreateSGReplica with "ASYNC"
    And I call GetRDFDevicePairInfo
    Then the error message contains <errormsg>

    Examples:
    |     induced     |            errormsg               |  arrays     |
    |     "none"      |              "none"               |      ""     |
    |  "InvalidJSON"  |       "invalid character"         |      ""     |
    |     "none"      |    "ignored as it is not managed" |  "ignored"  |
    | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Get Protected StorageGroup
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have 1 volumes
    When I call GetProtectedStorageGroup
    Then the error message contains <errormsg>

    Examples:
    |     induced     |            errormsg               |  arrays     |
    |     "none"      |              "none"               |      ""     |
    |  "InvalidJSON"  |       "invalid character"         |      ""     |
    |     "none"      |    "ignored as it is not managed" |  "ignored"  |
    | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Get RDFGroup info
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call GetRDFGroup
    Then the error message contains <errormsg>

    Examples:
    |     induced     |            errormsg               |  arrays     |
    |     "none"      |              "none"               |      ""     |
    |  "InvalidJSON"  |       "invalid character"         |      ""     |
    |     "none"      |    "ignored as it is not managed" |  "ignored"  |
    | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Add volumes to protected ASYNC storage-group
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have <vol> volumes
    When I call AddVolumesToProtectedStorageGroup with "ASYNC"
    Then the error message contains <errormsg>
    And the volumes should "" be replicated

    Examples:
    | vol |     induced     |                   errormsg                     |  arrays     |
    |  5  |     "none"      |                    "none"                      |      ""     |
    |  5  |     "none"      |         "ignored as it is not managed"         |  "ignored"  |
    |  5  | "httpStatus500" |               "Internal Error"                 |      ""     |
    |  0  |     "none"      |  "at least one volume id has to be specified"  |      ""     |

  @srdf
  Scenario Outline: Add volumes to protected METRO storage-group
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have <vol> volumes
    When I call AddVolumesToProtectedStorageGroup with "METRO"
    Then the error message contains <errormsg>
    And the volumes should "" be replicated

    Examples:
      | vol |     induced     |                   errormsg                     |  arrays     |
      |  5  |     "none"      |                    "none"                      |      ""     |
      |  5  |     "none"      |         "ignored as it is not managed"         |  "ignored"  |
      |  5  | "httpStatus500" |               "Internal Error"                 |      ""     |
      |  0  |     "none"      |  "at least one volume id has to be specified"  |      ""     |

  @srdf
  Scenario Outline: Test cases for Synchronous CreateVolumeInProtectedStorageGroupS
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call CreateVolumeInProtectedStorageGroupS with name <volname> and size <size>
    Then the error message contains <errormsg>
    And I get a valid Volume with name <volname> if no error

    Examples:
      | volname                                                                         | size | induced                   | errormsg                                               | arrays    |
      | "IntgA"                                                                         | 1    | "none"                    | "none"                                                 | ""        |
      | "IntgB"                                                                         | 5    | "none"                    | "none"                                                 | ""        |
      | "IntgG"                                                                         | 1    | "GetVolumeError"          | "failed to find newly created volume with name: IntgG" | ""        |
      | "IntgH"                                                                         | 1    | "VolumeNotCreatedError"   | "failed to find newly created volume with name: IntgH" | ""        |
      | "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxy" | 1    | "none"                    | "Length of volumeName exceeds max limit"               | ""        |
      | "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk"               | 1    | "none"                    | "none"                                                 | ""        |
      | "IntgA"                                                                         | 1    | "none"                    | "ignored as it is not managed"                         | "ignored" |
      | "IntgA"                                                                         | 1    | "UpdateStorageGroupError" | "induced error"                                        | ""        |

  @srdf
  Scenario Outline: Remove volumes from storage-group
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have <vol> volumes
    And I call CreateSGReplica with "ASYNC"
    When I call RemoveVolumesFromProtectedStorageGroup
    Then the error message contains <errormsg>
    And the volumes should "not" be replicated

    Examples:
    | vol |     induced     |                   errormsg                     |  arrays     |
    |  5  |     "none"      |                    "none"                      |      ""     |
    |  5  |  "InvalidJSON"  |             "invalid character"                |      ""     |
    |  5  |     "none"      |         "ignored as it is not managed"         |  "ignored"  |
    |  5  | "httpStatus500" |               "Internal Error"                 |      ""     |
    |  0  |     "none"      |  "at least one volume id has to be specified"  |      ""     |

  @srdf
  Scenario Outline: Create an SRDF Pair
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have 1 volumes
    When I call CreateRDFPair with "ASYNC"
    Then the error message contains <errormsg>

  Examples:
  |     induced     |            errormsg               |  arrays     |
  |     "none"      |              "none"               |      ""     |
  |  "InvalidJSON"  |       "invalid character"         |      ""     |
  |     "none"      |    "ignored as it is not managed" |  "ignored"  |
  | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Create an SRDF Pair
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have 1 volumes
    When I call CreateRDFPair with "METRO"
    Then the error message contains <errormsg>

    Examples:
      |     induced     |            errormsg               |  arrays     |
      |     "none"      |              "none"               |      ""     |
      |  "InvalidJSON"  |       "invalid character"         |      ""     |
      |     "none"      |    "ignored as it is not managed" |  "ignored"  |
      | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Execute Action
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I have 1 volumes
    When I call ExecuteAction <action>
    Then the error message contains <errormsg>

  Examples:
    | induced              | errormsg                       | action      | arrays    |
    | "none"               | "none"                         | "Suspend"   | ""        |
    | "none"               | "none"                         | "Resume"    | ""        |
    | "none"               | "none"                         | "Failback"  | ""        |
    | "none"               | "none"                         | "Failover"  | ""        |
    | "none"               | "none"                         | "Establish" | ""        |
    | "none"               | "none"                         | "Swap"      | ""        |
    | "none"               | "not a supported action"       | "Dance"     | ""        |
    | "none"               | "ignored as it is not managed" | "Suspend"   | "ignored" |
    | "httpStatus500"      | "Internal Error"               | "Suspend"   | ""        |
    | "ExecuteActionError" | "induced error"                | "Resume"    | ""        |

  @autosrdf
  Scenario Outline: GetFreeLocalAndRemoteRDFg - Create a SRDF Pair with auto SRDF group creation
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I call GetFreeLocalAndRemoteRDFg
    Then the error message contains <errormsg>
    Examples:
      | induced            | errormsg                       | arrays    |
      | "none"             | "none"                         | ""        |
      | "GetFreeRDFGError" | "induced error"                | ""        |
      | "InvalidJSON"      | "invalid character"            | ""        |
      | "httpStatus500"    | "Internal Error"               | ""        |
      | "none"             | "ignored as it is not managed" | "ignored" |

  @autosrdf
  Scenario Outline: GetFreeLocalAndRemoteRDFg - Create a SRDF Pair with auto SRDF group creation
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I call GetFreeLocalAndRemoteRDFg
    Then the error message contains <errormsg>
    Examples:
      | induced            | errormsg                       | arrays    |
      | "none"             | "none"                         | ""        |
      | "GetFreeRDFGError" | "induced error"                | ""        |
      | "InvalidJSON"      | "invalid character"            | ""        |
      | "httpStatus500"    | "Internal Error"               | ""        |
      | "none"             | "ignored as it is not managed" | "ignored" |

  @autosrdf
  Scenario Outline: GetLocalOnlineRDFDirs - Create a SRDF Pair with auto SRDF group creation
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I call GetLocalOnlineRDFDirs
    Then the error message contains <errormsg>
    Examples:
      | induced                      | errormsg                       | arrays    |
      | "none"                       | "none"                         | ""        |
      | "GetLocalOnlineRDFDirsError" | "induced error"                | ""        |
      | "InvalidJSON"                | "invalid character"            | ""        |
      | "httpStatus500"              | "Internal Error"               | ""        |
      | "none"                       | "ignored as it is not managed" | "ignored" |

  @autosrdf
  Scenario Outline: GetLocalOnlineRDFPorts - Create a SRDF Pair with auto SRDF group creation
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I call GetLocalOnlineRDFPorts
    Then the error message contains <errormsg>
    Examples:
      | induced                       | errormsg                       | arrays    |
      | "none"                        | "none"                         | ""        |
      | "GetLocalOnlineRDFPortsError" | "induced error"                | ""        |
      | "InvalidJSON"                 | "invalid character"            | ""        |
      | "httpStatus500"               | "Internal Error"               | ""        |
      | "none"                        | "ignored as it is not managed" | "ignored" |

  @autosrdf
  Scenario Outline: GetRemoteRDFPortOnSAN - Create a SRDF Pair with auto SRDF group creation
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I call GetRemoteRDFPortOnSAN
    Then the error message contains <errormsg>
    Examples:
      | induced                      | errormsg                       | arrays    |
      | "none"                       | "none"                         | ""        |
      | "GetRemoteRDFPortOnSANError" | "induced error"                | ""        |
      | "InvalidJSON"                | "invalid character"            | ""        |
      | "httpStatus500"              | "Internal Error"               | ""        |
      | "none"                       | "ignored as it is not managed" | "ignored" |

  @autosrdf
  Scenario Outline: GetLocalRDFPortDetails - Create a SRDF Pair with auto SRDF group creation
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I call GetLocalRDFPortDetails
    Then the error message contains <errormsg>
    Examples:
      | induced                       | errormsg                       | arrays    |
      | "none"                        | "none"                         | ""        |
      | "GetLocalRDFPortDetailsError" | "induced error"                | ""        |
      | "InvalidJSON"                 | "invalid character"            | ""        |
      | "httpStatus500"               | "Internal Error"               | ""        |
      | "none"                        | "ignored as it is not managed" | "ignored" |

  @autosrdf
  Scenario Outline: GetRDFGroupList - Create a SRDF Pair with auto SRDF group creation
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I call GetRDFGroupList with query <query>
    Then the error message contains <errormsg>
    Examples:
      | induced            | errormsg                       | query                 | arrays    |
      | "none"             | "none"                         | ""                    | ""        |
      | "none"             | "none"                         | "remote_symmetrix_id" | ""        |
      | "none"             | "none"                         | "volume_count"        | ""        |
      | "InvalidJSON"      | "invalid character"            | ""                    | ""        |
      | "httpStatus500"    | "Internal Error"               | ""                    | ""        |
      | "none"             | "ignored as it is not managed" | ""                    | "ignored" |
      | "GetRDFGroupError" | "induced error"                | ""                    | ""        |

  @autosrdf
  Scenario Outline: ExecuteCreateRDFGroup - Create a SRDF Pair with auto SRDF group creation
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    And I call ExecuteCreateRDFGroup
    Then the error message contains <errormsg>
    Examples:
      | induced               | errormsg                       | arrays    |
      | "none"                | "none"                         | ""        |
      | "CreateRDFGroupError" | "induced error"                | ""        |
      | "httpStatus500"       | "Internal Error"               | ""        |
      | "none"                | "ignored as it is not managed" | "ignored" |
    