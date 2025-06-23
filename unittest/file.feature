Feature: PMAX file test

  @v2.4.0
  Scenario Outline: TestCases for GetFileSystemList
    Given a valid connection
    And I induce error <induced>
    When I call GetFileSystemList
    Then the error message contains <errormsg>
    And I get a valid FileSystem ID List if no error

    Examples:
      | induced                    | errormsg               |
      | "none"                     | "none"                 |
      | "GetFileSystemListError"   | "induced error"        |
      | "InvalidResponse"          | "EOF"                  |

  @v2.4.0
  Scenario: TestCases for GetFileSystemList with Param
    Given a valid connection
    When I call GetFileSystemListWithParam
    Then the error message contains "none"
    And I get a valid FileSystem ID List if no error


  @v2.4.0
  Scenario Outline: TestCases for GetNASServerList
    Given a valid connection
    And I induce error <induced>
    When I call GetNASServerList
    Then the error message contains <errormsg>
    And I get a valid NAS Server ID List if no error

    Examples:
      | induced                    | errormsg               |
      | "none"                     | "none"                 |
      | "GetNASServerListError"    | "induced error"        |
      | "InvalidResponse"          | "EOF"                  |

  @v2.4.0
  Scenario: TestCases for GetNASServerList with Param
    Given a valid connection
    When I call GetNASServerListWithParam
    Then the error message contains "none"
    And I get a valid NAS Server ID List if no error

  @v2.4.0
  Scenario Outline: TestCases for GetNFSExportList
    Given a valid connection
    And I induce error <induced>
    When I call GetNFSExportList
    Then the error message contains <errormsg>
    And I get a valid NFS Export ID List if no error

    Examples:
      | induced                    | errormsg               |
      | "none"                     | "none"                 |
      | "GetNFSExportListError"    | "induced error"        |
      | "InvalidResponse"          | "EOF"                  |

  @v2.4.0
  Scenario: TestCases for GetNFSExportList with Param
    Given a valid connection
    When I call GetNFSExportListWithParam
    Then the error message contains "none"
    And I get a valid NFS Export ID List if no error

  @v2.4.0
  Scenario Outline: Test cases for GetFileSystemByID
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call GetFileSystemByID <id>
    Then the error message contains <errormsg>
    And I get a valid fileSystem Object if no error

    Examples:
      | id            | induced                  | errormsg                      | arrays    |
      | "id1"         | "none"                   | "none"                        | ""        |
      | "id2"         | "none"                   | "cannot be found"             | ""        |
      | "id1"         | "GetFileSystemError"     | "induced error"               | ""        |
      | "id1"         | "InvalidResponse"        | "EOF"                         | ""        |
      | "id1"         | "httpStatus500"          | "Internal Error"              | ""        |
      | "id1"         | "InvalidJSON"            | "invalid character"           | ""        |
      | "id1"         | "none"                   | "ignored as it is not managed"| "ignored" |

  @v2.4.0
  Scenario Outline: Test cases for CreateFileSystem
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call CreateFileSystem <name>
    Then the error message contains <errormsg>
    And I get a valid fileSystem Object if no error

    Examples:
      | name          | induced                 | errormsg                      | arrays    |
      | "fs-1"        | "none"                  | "none"                        | ""        |
      | "fs-1"        | "none"                  | "none"                        | ""        |
      | "fs-1"        | "httpStatus500"         | "Internal Error"              | ""        |
      | "fs-1"        | "InvalidJSON"           | "invalid character"           | ""        |
      | "fs-1"        | "none"                  | "ignored as it is not managed"| "ignored" |
      | "fs-1"        | "CreateFileSystemError" | "induced error"               | ""        |

  @v2.4.0
  Scenario Outline: Test cases for ModifyFileSystem
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call ModifyFileSystem on <name>
    Then the error message contains <errormsg>
    And I get a valid fileSystem Object if no error

    Examples:
      | name         | induced                 | errormsg                      | arrays    |
      | "id1"        | "none"                  | "none"                        | ""        |
      | "id1"        | "none"                  | "none"                        | ""        |
      | "id1"        | "httpStatus500"         | "Internal Error"              | ""        |
      | "id1"        | "InvalidJSON"           | "invalid character"           | ""        |
      | "id1"        | "none"                  | "ignored as it is not managed"| "ignored" |
      | "id1"        | "UpdateFileSystemError" | "induced error"               | ""        |

  @v2.4.0
  Scenario Outline: Test cases for DeleteFileSystem
    Given a valid connection
    And I have an allowed list of <arrays>
    And I call CreateFileSystem <name>
    And I get a valid fileSystem Object if no error
    And I induce error <induced>
    Then I call DeleteFileSystem
    Then the error message contains <errormsg>

    Examples:
      | name         | induced                 | errormsg                      | arrays    |
      | "fs-del"     | "none"                  | "none"                        | ""        |
      | "fs-del"     | "DeleteFileSystemError" | "induced error"               | ""        |

  @v2.4.0
  Scenario Outline: Test cases for DeleteFileSystem with no allowed array
    Given a valid connection
    And I have an allowed list of <arrays>
    Then I call DeleteFileSystem
    Then the error message contains <errormsg>
    Examples:
      | errormsg                       | arrays    |
      | "ignored as it is not managed" | "ignored" |

  @v2.4.0
  Scenario Outline: Test cases for GetNASServerByID
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call GetNASServerByID <id>
    Then the error message contains <errormsg>
    And I get a valid nasServer Object if no error

    Examples:
      | id            | induced                  | errormsg                      | arrays    |
      | "id1"         | "none"                   | "none"                        | ""        |
      | "id3"         | "none"                   | "cannot be found"             | ""        |
      | "id1"         | "GetNASServerError"      | "induced error"               | ""        |
      | "id1"         | "InvalidResponse"        | "EOF"                         | ""        |
      | "id1"         | "httpStatus500"          | "Internal Error"              | ""        |
      | "id1"         | "InvalidJSON"            | "invalid character"           | ""        |
      | "id1"         | "none"                   | "ignored as it is not managed"| "ignored" |

  @v2.4.0
  Scenario Outline: Test cases for ModifyNASServer
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call ModifyNASServer on <id>
    Then the error message contains <errormsg>
    And I get a valid nasServer Object if no error

    Examples:
      | id           | induced                 | errormsg                      | arrays    |
      | "id1"        | "none"                  | "none"                        | ""        |
      | "id1"        | "none"                  | "none"                        | ""        |
      | "id1"        | "httpStatus500"         | "Internal Error"              | ""        |
      | "id1"        | "InvalidJSON"           | "invalid character"           | ""        |
      | "id1"        | "none"                  | "ignored as it is not managed"| "ignored" |
      | "id1"        | "UpdateNASServerError"  | "induced error"               | ""        |

  @v2.4.0
  Scenario Outline: Test cases for DeleteNASServer
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    Then I call DeleteNASServer "id2"
    Then the error message contains <errormsg>

    Examples:
     | induced                 | errormsg                       | arrays    |
     | "none"                  | "none"                         | ""        |
     | "DeleteNASServerError"  | "induced error"                | ""        |
     | "none"                  | "ignored as it is not managed" | "ignored" |

  @v2.4.0
  Scenario Outline: Test cases for GetNFSExportByID
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call GetNFSExportByID <id>
    Then the error message contains <errormsg>
    And I get a valid NFSExport object if no error

    Examples:
      | id            | induced                  | errormsg                      | arrays    |
      | "id1"         | "none"                   | "none"                        | ""        |
      | "id3"         | "none"                   | "cannot be found"             | ""        |
      | "id1"         | "GetNFSExportError"      | "induced error"               | ""        |
      | "id1"         | "InvalidResponse"        | "EOF"                         | ""        |
      | "id1"         | "httpStatus500"          | "Internal Error"              | ""        |
      | "id1"         | "InvalidJSON"            | "invalid character"           | ""        |
      | "id1"         | "none"                   | "ignored as it is not managed"| "ignored" |

  @v2.4.0
  Scenario Outline: Test cases for CreateNFSExport
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call CreateNFSExport <name>
    Then the error message contains <errormsg>
    And I get a valid NFSExport object if no error

    Examples:
      | name          | induced                 | errormsg                      | arrays    |
      | "nfs-1"        | "none"                  | "none"                        | ""        |
      | "nfs-1"        | "none"                  | "none"                        | ""        |
      | "nfs-1"        | "httpStatus500"         | "Internal Error"              | ""        |
      | "nfs-1"        | "InvalidJSON"           | "invalid character"           | ""        |
      | "nfs-1"        | "none"                  | "ignored as it is not managed"| "ignored" |
      | "nfs-1"        | "CreateNFSExportError"  | "induced error"               | ""        |


  @v2.4.0
  Scenario Outline: Test cases for ModifyNFSExport
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call ModifyNFSExport on <name>
    Then the error message contains <errormsg>
    And I get a valid NFSExport object if no error

    Examples:
      | name         | induced                 | errormsg                      | arrays    |
      | "id1"        | "none"                  | "none"                        | ""        |
      | "id1"        | "none"                  | "none"                        | ""        |
      | "id1"        | "httpStatus500"         | "Internal Error"              | ""        |
      | "id1"        | "InvalidJSON"           | "invalid character"           | ""        |
      | "id1"        | "none"                  | "ignored as it is not managed"| "ignored" |
      | "id1"        | "UpdateNFSExportError"  | "induced error"               | ""        |

  @v2.4.0
  Scenario Outline: Test cases for DeleteNFSExport
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    Then I call DeleteNFSExport "id2"
    Then the error message contains <errormsg>

    Examples:
      | induced                 | errormsg                       | arrays    |
      | "none"                  | "none"                         | ""        |
      | "DeleteNFSExportError"  | "induced error"                | ""        |
      | "none"                  | "ignored as it is not managed" | "ignored" |

  @v2.4.0
  Scenario Outline: Test cases for GetFileInterfaceByID
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call GetFileInterfaceByID <id>
    Then the error message contains <errormsg>
    And I get a valid fileInterface Object if no error

    Examples:
      | id            | induced                  | errormsg                      | arrays    |
      | "id1"         | "none"                   | "none"                        | ""        |
      | "id3"         | "none"                   | "Could not find"              | ""        |
      | "id1"         | "GetFileInterfaceError"  | "induced error"               | ""        |
      | "id1"         | "InvalidResponse"        | "EOF"                         | ""        |
      | "id1"         | "httpStatus500"          | "Internal Error"              | ""        |
      | "id1"         | "InvalidJSON"            | "invalid character"           | ""        |
      | "id1"         | "none"                   | "ignored as it is not managed"| "ignored" |

  @v2.4.0
  Scenario Outline: TestCases for GetNFSServerList
    Given a valid connection
    And I induce error <induced>
    When I call GetNFSServerList
    Then the error message contains <errormsg>
    And I get a valid NFS Server ID List if no error

    Examples:
      | induced                    | errormsg               |
      | "none"                     | "none"                 |
      | "GetNFSServerListError"    | "induced error"        |
      | "InvalidResponse"          | "EOF"                  |

  @v2.4.0
  Scenario Outline: Test cases for GetNFSServerByID
    Given a valid connection
    And I have an allowed list of <arrays>
    And I induce error <induced>
    When I call GetNFSServerByID <id>
    Then the error message contains <errormsg>
    And I get a valid nfsServer Object if no error

    Examples:
      | id            | induced                  | errormsg                      | arrays    |
      | "id1"         | "none"                   | "none"                        | ""        |
      | "id3"         | "none"                   | "cannot be found"             | ""        |
      | "id1"         | "GetNFSServerError"      | "induced error"               | ""        |
      | "id1"         | "InvalidResponse"        | "EOF"                         | ""        |
      | "id1"         | "httpStatus500"          | "Internal Error"              | ""        |
      | "id1"         | "InvalidJSON"            | "invalid character"           | ""        |
      | "id1"         | "none"                   | "ignored as it is not managed"| "ignored" |

