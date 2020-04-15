# GO powermax REST library
This directory contains a lighweight Go wrapper around the Unisphere REST API

## Unit Tests
Unit Tests exist for the wrapper. These tests do not modify the array.

#### Running Unit Tests
To run these tests, from this directory, run:
```
make unit-test
```

#### Running Unit Tests With Debugging enabled
To run the tests and be able to attach debugger to the tests, run:
```
make unit-test-debug
```

Or this can be run in steps. Build the debug exec:

```
make unit-test-debug-build
```

Then to start with debugging:

```
make dlv-unit-test
```

The process will listen on port 55555 for a debugger to attach. Once the debugger is attached, the tests will start executing.

## Integration Tests
Integration Tests exist for the wrapper as well. These tests WILL MODIFY the array.

#### Pre-requisites
Before running integration tests, do the following:

* Modify the Unisphere endpoint and Symmetrix ID in the Makefile. 

* Examine the inttest/pmax_integration_test.go file of
the repository. Within that file, two variables are defined:
    * username
    * password
 
   Either change those variables to match an existing user in Unisphere, or create
   a new user in Unisphere matching those credentials.
* The integration test expects certain storage objects to be present on the PowerMax array you are using for integration tests. Examine the file inttest/pmax-integration_test.go and modify the declared variables with appropriate names from the PowerMax array in use.
For e.g. - Set `defaultStorageGroup` to an existing storage group from the array. 


#### Running Integration Tests
To run these tests, from the this directory, run:

For full tests:
```
make int-test
```

For an abbreviated set of tests:
```
make short-int-test
```

