#!/bin/bash
# This will run coverage analysis using the integration testing.
# The user.env has placeholders for required values and must point to a valid Unisphere deployment
# This will make real calls to  Unisphere

integrationfiles="inttest/pmax_integration_test.go inttest/pmax_replication_integration_test.go"
# Usage information
function usage {
   echo
   echo "`basename ${0}`"
   echo "    --no-cleanup   - Disable the default cleaning of SGs and Hosts created during test"
   echo "    --short        - To run an abbreviated set of tests"
   exit 1
}

# Default values
QUAL="false"
ENV=user.env
NOCLEANUP="no"
SHORT="no"

while getopts ":h-:" opt; do
    case "${opt}" in
    -)
        case "${OPTARG}" in
        env )
            ENV=${!OPTIND}
            QUAL="true"
            OPTIND=$((OPTIND + 1))
            ;;
        no-cleanup )
            NOCLEANUP="yes"
            ;;
        short )
            SHORT="yes"
            ;;
        *)
            echo "Invalid option"
            usage
            exit 1
        esac
        ;;
    h)
        usage 
        exit 0
        ;;
    *)
        echo "Invalid option"
        usage
        exit 1
        ;;
    esac
done

if [ ! -f "inttest/user.env" ]; then
    echo "missing user.env"
    exit 1
fi
echo "applying user configurations"
source inttest/user.env

if [ ${QUAL} = "true" ]; then
    if [ ! -f "testenvs/${ENV}" ]; then
    echo "missing env file"
    exit 1
    fi
echo "overwriting user configurations"
echo "using ${ENV} configurations"
source testenvs/${ENV}
fi

if [ ${NOCLEANUP} = "yes" ]; then
    echo "disbling the default storage group and host clean up after the tests have run"
    export  Cleanup="false"
fi

if [ ${SHORT} = "yes" ]; then
    echo "running an abbreviated set of tests"
    go test -v -short -timeout 90m -coverprofile=c.out -coverpkg github.com/dell/gopowermax/v2 $integrationfiles
else
    go test -v -timeout 90m -coverprofile=c.out -coverpkg github.com/dell/gopowermax/v2 $integrationfiles
fi
