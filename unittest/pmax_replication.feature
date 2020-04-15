Feature: PMAX replication test

  Scenario: Snapshot licence
    Given a valid connection
    When I excute the capabilities on the symmetrix array
    Then the error message contains "none"

  Scenario Outline: Create a snapshot on a source volume
    Given a valid connection
    And I have 5 volumes 
    When I call CreateSnapshot with <volIDs> and snapshot <snapID> on it
    Then the error message contains <errormsg>
    And I get a valid Snapshot object if no error

    Examples:
    | volIDs                 |  snapID      | errormsg          |
    | "00001,00002,00003"    | "snapshot1"  | "none"            |
    | "00001,00001"          | "snapshot1"  | "none"            |
    | "00001,00002,00003"    | "snap:shot"  | "error"           |
    | "00001,00007"          | "snapshot1"  | "not available"   |

  Scenario Outline: List all volumes with snapshots
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I have 2 volumes
    And I induce error <induced>
    And I call CreateSnapshot with "00001,00002" and snapshot "snapshot1" on it
    When I call GetSnapVolumeList with <queryKey> and <queryValue>
    Then the error message contains <errormsg>
    And I should get a list of volumes having snapshots if no error
  
    Examples:
      | queryKey         | queryValue |  errormsg                    |  induced            | whitelist |
      | ""               |  ""        |  "none"                      | "none"              |   ""      |
      | "includeDetails" | "true"     |  "none"                      | "none"              |   ""      | 
      | ""               |  ""        |  "induced error"             | "GetSymVolumeError" |   ""      |
      | ""               |  ""        |  "ignored via a whitelist"   | "none"              | "ignored" |


  Scenario Outline: List all Snapshot for a volume
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I have 3 volumes
    And I induce error <induced>
    And I call CreateSnapshot with <volIDs> and snapshot <snapID> on it
    When I call GetVolumeSnapInfo with volume <volID>
    Then the error message contains <errormsg>
    And I should get a list of snapshots if no error

    Examples:
      | volIDs                 |  snapID                 |   volID        | errormsg                  | whitelist | induced            |
      | "00001,00001"          | "snapshot1"             |  "00001"       | "none"                    |   ""      | "none"             |
      | "00001,00002,00003"    | "snapshot1"             |  "00002"       | "none"                    |   ""      | "none"             |
      | "00001"                | "snapshot1"             |  "00002"       | "none"                    |   ""      | "none"             |
      | "00001"                | "snapshot1"             |  "00004"       | "cannot be found"         |   ""      | "none"             |
      | "00001"                | "snapshot1"             |  "00002"       | "ignored via a whitelist" | "ignored" | "none"             |
      | "00001"                | "snapshot1"             |  "00002"       | "induced error"           |   ""      | "GetVolSnapsError" |

  Scenario Outline: Get the Snapshot for linked or Unlinked volumes
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I have 5 volumes
    And I induce error <induced>
    And I call CreateSnapshot with "00001,00002" and snapshot <snapID> on it
    And I call ModifySnapshot with "00002,00002", "00004,00005", <snapID>, "", 0 and "Link"
    When I call GetSnapshotInfo with <volID> and snapshot <snapID> on it
    Then the error message contains <errormsg>
    And I should get the snapshot details if no error

    Examples:
      | volID         |    snapID   |  errormsg                  | whitelist | induced            |
      | "00001"       | "snapshot1" |  "none"                    |    ""     | "none"             |
      | "00004"       | "snapshot1" |  "none"                    |    ""     | "none"             |
      | "00002"       | "snapshot1" |  "none"                    |    ""     | "none"             |
      | "00005"       | "snapshot1" |  "none"                    |    ""     | "none"             |
      | "00003"       | "snapshot1" |  "none"                    |    ""     | "none"             |
      | "00007"       | "snapshot1" |  "cannot be found"         |    ""     | "none"             |
      | "00007"       | "snapshot1" |  "ignored via a whitelist" | "ignored" | "none"             |
      | "00007"       | "snapshot1" |  "induced error"           |    ""     | "GetVolSnapsError" |


  Scenario Outline: Get a list Generation for given Snapshot
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I have 3 volumes
    And I call CreateSnapshot with <volIDs> and snapshot <snapID> on it
    When I call GetSnapshotGenerations with <volID> and snapshot <snapID> on it
    Then the error message contains <errormsg> 
    And I should get the generation list if no error

    Examples:
      | volIDs                 |  snapID                 | volID          | errormsg                    | whitelist |
      | "00001,00001"          | "snapshot1"             |  "00001"       | "none"                      |    ""     |
      | "00001"                | "snapshot1"             |  "00002"       | "none"                      |    ""     |
      | "00001"                | "snapshot1"             |  "00007"       | "cannot be found"           |    ""     |
      | "00001"                | "snapshot1"             |  "00007"       | "ignored via a whitelist"   | "ignored" |

  Scenario Outline: Get a Generation Info for given Snapshot
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I have 3 volumes
    And I call CreateSnapshot with <volIDs> and snapshot <snapID> on it
    And I call GetSnapshotGeneration with <volID>, snapshot <snapID> and <genID> on it
    Then the error message contains <errormsg> 
    And I should get a generation Info if no error

    Examples:
      | volIDs                 |  snapID                 | volID          | genID  | errormsg                    | whitelist |
      | "00001,00001"          | "snapshot1"             |  "00001"       |   1    | "none"                      |    ""     |
      | "00001"                | "snapshot1"             |  "00002"       |   0    | "none"                      |    ""     |
      | "00001"                | "snapshot1"             |  "00007"       |   0    | "cannot be found"           |    ""     |
      | "00001"                | "snapshot1"             |  "00007"       |   0    | "ignored via a whitelist"   | "ignored" |

  Scenario Outline: Renaming a snapshot
    Given a valid connection
    And I induce error <induced>
    And I have 3 volumes
    And I call CreateSnapshot with <volIDs> and snapshot <snapID> on it
    When I call ModifySnapshot with <source>, <target>, <snapID>, <newSnapID>, <genID> and <action>
    Then the error message contains <errormsg>
    And I should get a valid response if no error

    Examples:
    |   volIDs    |    source   | target |  snapID     |  newSnapID      | genID |   action   |  errormsg                   | induced           |
    |   "00001"   |    "00001"  |   ""   | "snapshot1" | "snapshot_csi"  |    0  |  "Rename"  |  "none"                     | "none"            |
    |   "00001"   |    "00002"  |   ""   | "snapshot1" | "snapshot_csi"  |    0  |  "Rename"  |  "no snapshot information"  | "none"            |
    |   "00001"   |    "00001"  |   ""   | "snapshot1" | "snapshot_csi"  |    0  |  "Rename"  |  "Not Found"                | "JobFailedError"  |

  Scenario Outline: Linking a snapshot
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I have 4 volumes
    And I call CreateSnapshot with <volIDs> and snapshot <snapID> on it
    When I call ModifySnapshot with <source>, <target>, <snapID>, "", 0 and <action>
    Then the error message contains <errormsg>
    And I should get a valid response if no error

      Examples:
    |   volIDs         |    source         |   target        |   snapID    |  action  |  errormsg                  | whitelist |
    |   "00001"        |    "00001"        |   "00002"       | "snapshot1" |  "Link"  |  "none"                    |    ""     |
    |   "00001,00002"  |    "00001"        |   "00002"       | "snapshot1" |  "Link"  |  "none"                    |    ""     |
    |   "00001,00002"  |    "00001,00002"  |   "00003,00004" | "snapshot1" |  "Link"  |  "none"                    |    ""     |
    |   "00001,00002"  |    "00001,00001"  |   "00003,00004" | "snapshot1" |  "Link"  |  "none"                    |    ""     |
    |   "00001,00002"  |    "00001,00001"  |   "00003,00004" | "snapshot1" |  "Link"  |  "ignored via a whitelist" | "ignored" |
    |   "00001,00002"  |    "00001,00001"  |   "00002,00002" | "snapshot1" |  "Link"  |  "already in desired state"|    ""     |
    |   "00001,00002"  |    "00001,00002"  |   "00002"       | "snapshot1" |  "Link"  |  "cannot link snapshot"    |    ""     |
    |   "00001"        |    "00002"        |   "00004"       | "snapshot1" |  "Link"  |  "no snapshot information" |    ""     |
    |   "00001"        |    "00005"        |   "00004"       | "snapshot1" |  "Link"  |  "devices not available"   |    ""     |
    |   "00001"        |    "00004"        |   "00005"       | "snapshot1" |  "Link"  |  "devices not available"   |    ""     |
    |   "00001"        |      ""           |   "00002"       | "snapshot1" |  "Link"  |  "no source volume"        |    ""     |
    |   "00001"        |    "00001"        |     ""          | "snapshot1" |  "Link"  |  "no link volume"          |    ""     |
    |   "00001"        |    "00001"        |   "00002"       | "snapshot1" |    ""    |  "not a supported action"  |    ""     |

  Scenario Outline: Unlinking a snapshot
    Given a valid connection
    And I have 5 volumes
    And I call CreateSnapshot with "00001,00003,00005" and snapshot <snapID> on it
    And I call ModifySnapshot with "00001,00003", "00002,00004", <snapID>, "", 0 and "Link"
    When I call ModifySnapshot with <source>, <target>, <snapID>, "", 0 and <action>
    Then the error message contains <errormsg>
    And I should get a valid response if no error 

    Examples:
    |    source        |   target        |   snapID    |   action   |  errormsg                   |
    |    "00001"       |   "00002"       | "snapshot1" |  "Unlink"  |  "none"                     |
    |    "00001,00003" |   "00002,00004" | "snapshot1" |  "Unlink"  |  "none"                     |
    |    "00001"       |   "00008"       | "snapshot1" |  "Unlink"  |  "devices not available"    |
    |    "00005"       |   "00008"       | "snapshot1" |  "Unlink"  |  "devices not available"    |
    |    "00001"       |   "00002"       | "snapshot1" |    ""      |  "not a supported action"   |
    |    "00001,00003" |   "00002"       | "snapshot1" |  "Unlink"  |  "cannot unlink snapshot"   |
    |      ""          |   "00002"       | "snapshot1" |  "Unlink"  |  "no source volume"         |
    |    "00001"       |     ""          | "snapshot1" |  "Unlink"  |  "no target volume"         |
    |    "00002"       |   "00001"       | "snapshot1" |  "Unlink"  |  "no snapshot information"  |
    |    "00002,00004" |   "00001,00003" | "snapshot1" |  "Unlink"  |  "no snapshot information"  |
    |    "00001"       |   "00005"       | "snapshot1" |  "Unlink"  |  "already in desired state" |
    |    "00001"       |   "00003"       | "snapshot1" |  "Unlink"  |  "already in desired state" |
    |    "00001,00003" |   "00004,00004" | "snapshot1" |  "Unlink"  |  "already in desired state" |
 
  Scenario Outline: Delete a snapshot
    Given a valid connection
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    And I have 4 volumes
    And I call CreateSnapshot with "00001,00002,00003" and snapshot <snapID> on it
    And I call CreateSnapshot with "00001,00002,00003" and snapshot "snapshot2" on it
    And I call ModifySnapshot with "00002", "00004", <snapID>, "", 0 and "Link"
    When I call DeleteSnapshot with <volID>, snapshot <snapID> and 0  on it
    Then the error message contains <errormsg>
    And I should get a valid response if no error

    Examples:
      | volID         |    snapID   |  errormsg                  | whitelist | induced          |
      | "00001"       | "snapshot1" |  "none"                    |    ""     | "none"           |
      | "00001,00003" | "snapshot1" |  "none"                    |    ""     | "none"           |
      | "00002"       | "snapshot1" |  "snapshot has a link"     |    ""     | "none"           |
      | "00007"       | "snapshot1" |  "devices not available"   |    ""     | "none"           |
      | "00004"       | "snapshot1" |  "no snapshot information" |    ""     | "none"           |
      |  ""           | "snapshot1" |  "no source volume"        |    ""     | "none"           |
      |  "00001"      | "snapshot1" |  "ignored via a whitelist" | "ignored" | "none"           |
      |  "00001"      | "snapshot1" | "Job status not successful"|    ""     | "JobFailedError" |
 
  Scenario Outline: Testing GetPrivVolumeByID
    Given a valid connection
    And I have 4 volumes
    And I have a whitelist of <whitelist>
    And I induce error <induced>
    And I call CreateSnapshot with "00001,00002" and snapshot "snapshot1" on it
    And I call CreateSnapshot with "00001" and snapshot "snapshot2" on it
    And I call ModifySnapshot with "00001,00001", "00002,00003", "snapshot1", "", 0 and "Link"
    When I call GetPrivVolumeByID with <volID>
    Then the error message contains <errormsg>
    And I should get a private volume information if no error

    Examples:
    | volID   | errormsg                  | whitelist | induced                  |
    | "00001" | "none"                    |   ""      | "none"                   |
    | "00002" | "none"                    |   ""      | "none"                   |
    | "00003" | "none"                    |   ""      | "none"                   |
    | "00004" | "none"                    |   ""      | "none"                   |
    | "00007" | "cannot be found"         |   ""      | "none"                   |
    | "00001" | "ignored via a whitelist" | "ignored" | "none"                   |
    | "00001" | "induced error"           |   ""      | "GetPrivVolumeByIDError" |