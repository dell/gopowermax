Feature: PMAX SRDF test

  @srdf
  Scenario Outline: Create a storage-group with volumes and protect it
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    And I have 5 volumes
    When I call CreateSGReplica
    Then the error message contains <errormsg>
    And then SG should be replicated
    And the volumes should "" be replicated

    Examples:
    |     induced     |            errormsg               |  whitelist  |
    |     "none"      |              "none"               |      ""     |
    |  "InvalidJSON"  |       "invalid character"         |      ""     |
    |     "none"      |     "ignored via a whitelist"     |  "ignored"  |
    | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Get SRDF info about a storage group
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    And I have 5 volumes
    When I call CreateSGReplica
    And I call GetStorageGroupRDFInfo
    Then the error message contains <errormsg>

    Examples:
    |     induced     |            errormsg               |  whitelist  |
    |     "none"      |              "none"               |      ""     |
    |  "InvalidJSON"  |       "invalid character"         |      ""     |
    |     "none"      |     "ignored via a whitelist"     |  "ignored"  |
    | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Get RDF device-pair-info on a volume
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    And I have 1 volumes
    When I call CreateSGReplica
    And I call GetRDFDevicePairInfo
    Then the error message contains <errormsg>

    Examples:
    |     induced     |            errormsg               |  whitelist  |
    |     "none"      |              "none"               |      ""     |
    |  "InvalidJSON"  |       "invalid character"         |      ""     |
    |     "none"      |     "ignored via a whitelist"     |  "ignored"  |
    | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Get Protected StorageGroup
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    And I have 1 volumes
    When I call GetProtectedStorageGroup
    Then the error message contains <errormsg>

    Examples:
    |     induced     |            errormsg               |  whitelist  |
    |     "none"      |              "none"               |      ""     |
    |  "InvalidJSON"  |       "invalid character"         |      ""     |
    |     "none"      |     "ignored via a whitelist"     |  "ignored"  |
    | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Get RDFGroup info
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    When I call GetRDFGroup
    Then the error message contains <errormsg>

    Examples:
    |     induced     |            errormsg               |  whitelist  |
    |     "none"      |              "none"               |      ""     |
    |  "InvalidJSON"  |       "invalid character"         |      ""     |
    |     "none"      |     "ignored via a whitelist"     |  "ignored"  |
    | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Add volumes to protected storage-group
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    And I have <vol> volumes
    When I call AddVolumesToProtectedStorageGroup
    Then the error message contains <errormsg>
    And the volumes should "" be replicated

    Examples:
    | vol |     induced     |                   errormsg                     |  whitelist  |
    |  5  |     "none"      |                    "none"                      |      ""     |
    |  5  |     "none"      |          "ignored via a whitelist"             |  "ignored"  |
    |  5  | "httpStatus500" |               "Internal Error"                 |      ""     |
    |  0  |     "none"      |  "at least one volume id has to be specified"  |      ""     |

  @srdf
  Scenario Outline: Remove volumes from storage-group
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    And I have <vol> volumes
    And I call CreateSGReplica
    When I call RemoveVolumesFromProtectedStorageGroup
    Then the error message contains <errormsg>
    And the volumes should "not" be replicated

    Examples:
    | vol |     induced     |                   errormsg                     |  whitelist  |
    |  5  |     "none"      |                    "none"                      |      ""     |
    |  5  |  "InvalidJSON"  |             "invalid character"                |      ""     |
    |  5  |     "none"      |          "ignored via a whitelist"             |  "ignored"  |
    |  5  | "httpStatus500" |               "Internal Error"                 |      ""     |
    |  0  |     "none"      |  "at least one volume id has to be specified"  |      ""     |

  @srdf
  Scenario Outline: Create an SRDF Pair
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    And I have 1 volumes
    When I call CreateRDFPair
    Then the error message contains <errormsg>

  Examples:
  |     induced     |            errormsg               |  whitelist  |
  |     "none"      |              "none"               |      ""     |
  |  "InvalidJSON"  |       "invalid character"         |      ""     |
  |     "none"      |     "ignored via a whitelist"     |  "ignored"  |
  | "httpStatus500" |          "Internal Error"         |      ""     |

  @srdf
  Scenario Outline: Create an SRDF Pair
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    And I have 1 volumes
    When I call ExecuteAction <action>
    Then the error message contains <errormsg>

  Examples:
  |     induced     |            errormsg               |  whitelist  |   action    |
  |     "none"      |              "none"               |      ""     |  "Suspend"  |
  |     "none"      |              "none"               |      ""     |  "Resume"   |
  |     "none"      |     "not a supported action"      |      ""     |   "Dance"   |
  |     "none"      |     "ignored via a whitelist"     |  "ignored"  |  "Suspend"  |
  | "httpStatus500" |          "Internal Error"         |      ""     |  "Suspend"  |
