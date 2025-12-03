#!/bin/bash

#
# Here give the name of the test serie
#

TEST_NAME=obiuniq
CMD=obiuniq

######
#
# Some variable and function definitions: please don't change them
#
######
TEST_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
OBITOOLS_DIR="${TEST_DIR/obitest*/}build"
export PATH="${OBITOOLS_DIR}:${PATH}"

MCMD="$(echo "${CMD:0:4}" | tr '[:lower:]' '[:upper:]')$(echo "${CMD:4}" | tr '[:upper:]' '[:lower:]')"

TMPDIR="$(mktemp -d)"
ntest=0
success=0
failed=0

cleanup() {
    echo "========================================" 1>&2
    echo "## Results of the $TEST_NAME tests:" 1>&2

    echo 1>&2
    echo "- $ntest tests run" 1>&2
    echo "- $success successfully completed" 1>&2
    echo "- $failed failed tests" 1>&2
    echo 1>&2
    echo "Cleaning up the temporary directory..." 1>&2
    echo 1>&2
    echo "========================================" 1>&2

    rm -rf "$TMPDIR"  # Suppress the temporary directory

    if [ $failed -gt 0 ]; then
       log "$TEST_NAME tests failed" 
        log
        log
       exit 1
    fi

    log
    log

    exit 0
}

log() {
    echo -e "[$TEST_NAME @ $(date)] $*" 1>&2
}

log "Testing $TEST_NAME..." 
log "Test directory is $TEST_DIR" 
log "obitools directory is $OBITOOLS_DIR" 
log "Temporary directory is $TMPDIR" 
log "files: $(find $TEST_DIR | awk -F'/' '{print $NF}' | tail -n +2)"

######################################################################
####
#### Below are the tests
####
#### Before each test :
####  - increment the variable ntest
####
#### Run the command as the condition of an if / then /else
####  - The command must return 0 on success
####  - The command must return an exit code different from 0 on failure
####  - The datafiles are stored in the same directory than the test script
####  - The test script directory is stored in the TEST_DIR variable
####  - If result files have to be produced they must be stored
####    in the temporary directory (TMPDIR variable)
####
#### then clause is executed on success of the command
####  - Write a success message using the log function
####  - increment the variable success
####
#### else clause is executed on failure of the command
####  - Write a failure message using the log function
####  - increment the variable failed
####
######################################################################



((ntest++))
if $CMD -h > "${TMPDIR}/help.txt" 2>&1 
then
    log "$MCMD: printing help OK" 
    ((success++))
else
    log "$MCMD: printing help failed" 
    ((failed++))
fi

((ntest++))
if obiuniq "${TEST_DIR}/touniq.fasta" \
    > "${TMPDIR}/touniq_u.fasta"
then
    log "OBIUniq simple: running OK" 
    ((success++))
else
    log "OBIUniq simple: running failed"
    ((failed++))
fi

obicsv -s --auto ${TEST_DIR}/touniq_u.fasta \
| tail -n +2 \
| sort \
> "${TMPDIR}/touniq_u_ref.csv"

obicsv -s --auto ${TMPDIR}/touniq_u.fasta \
| tail -n +2 \
| sort \
> "${TMPDIR}/touniq_u.csv"

((ntest++))
if diff "${TMPDIR}/touniq_u_ref.csv" \
        "${TMPDIR}/touniq_u.csv"  > /dev/null
then
    log "OBIUniq simple: result OK"
    ((success++))
else
    log "OBIUniq simple: result failed"
    ((failed++))
fi

((ntest++))
if obiuniq -c a "${TEST_DIR}/touniq.fasta" \
    > "${TMPDIR}/touniq_u_a.fasta"
then
    log "OBIUniq one category: running OK" 
    ((success++))
else
    log "OBIUniq one category: running failed"
    ((failed++))
fi

obicsv -s --auto ${TEST_DIR}/touniq_u_a.fasta \
| tail -n +2 \
| sort \
> "${TMPDIR}/touniq_u_a_ref.csv"

obicsv -s --auto ${TMPDIR}/touniq_u_a.fasta \
| tail -n +2 \
| sort \
> "${TMPDIR}/touniq_u_a.csv"


((ntest++))
if diff "${TMPDIR}/touniq_u_a_ref.csv" \
        "${TMPDIR}/touniq_u_a.csv"  > /dev/null
then
    log "OBIUniq one category: result OK"
    ((success++))
else
    log "OBIUniq one category: result failed"
    ((failed++))
fi

((ntest++))
if obiuniq -c a -c b "${TEST_DIR}/touniq.fasta" \
    > "${TMPDIR}/touniq_u_a_b.fasta"
then
    log "OBIUniq two categories: running OK" 
    ((success++))
else
    log "OBIUniq two categories: running failed"
    ((failed++))
fi

obicsv -s --auto ${TEST_DIR}/touniq_u_a_b.fasta \
| tail -n +2 \
| sort \
> "${TMPDIR}/touniq_u_a_b_ref.csv"

obicsv -s --auto ${TMPDIR}/touniq_u_a_b.fasta \
| tail -n +2 \
| sort \
> "${TMPDIR}/touniq_u_a_b.csv"

((ntest++))
if diff "${TMPDIR}/touniq_u_a_b_ref.csv" \
        "${TMPDIR}/touniq_u_a_b.csv"  > /dev/null
then
    log "OBIUniq two categories: result OK"
    ((success++))
else
    log "OBIUniq two categories: result failed"
    ((failed++))
fi

#########################################
#
# At the end of the tests
# the cleanup function is called
#
#########################################

cleanup
