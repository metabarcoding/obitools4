#!/bin/bash

#
# Here give the name of the test serie
#
TEST_NAME=obicount

######
#
# Some variable and function definitions: please don't change them
#
######
TEST_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
OBITOOLS_DIR="${TEST_DIR/obitest*/}build"
export PATH="${OBITOOLS_DIR}:${PATH}"


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
       exit 1
    fi

    exit 0
}

log() {
    echo "[$TEST_NAME @ $(date)] $*" 1>&2
}

log "Testing $TEST_NAME..." 
log "Test directory is $TEST_DIR" 
log "obitools directory is $OBITOOLS_DIR" 
log "Temporary directory is $TMPDIR" 

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
if obicount "${TEST_DIR}/wolf_F.fasta.gz" \
    > "${TMPDIR}/wolf_F.fasta_count.csv" 
then
    log "OBICount: fasta reading OK" 
    ((success++))
else
    log "OBICount: fasta reading failed" 
    ((failed++))
fi

((ntest++))
if obicount "${TEST_DIR}/wolf_F.fastq.gz" \
    > "${TMPDIR}/wolf_F.fastq_count.csv"
then
    log "OBICount: fastq reading OK"
    ((success++))
else
    log "OBICount: fastq reading failed" 
    ((failed++))
fi

((ntest++))
if obicount "${TEST_DIR}/wolf_F.csv.gz" \
    > "${TMPDIR}/wolf_F.csv_count.csv"
then
    log "OBICount: csv reading OK" 
    ((success++))
else
    log "OBICount: csv reading failed"
    ((failed++))
fi

((ntest++))
if diff "${TMPDIR}/wolf_F.fasta_count.csv" \
        "${TMPDIR}/wolf_F.fastq_count.csv"  > /dev/null
then
    log "OBICount: counting on fasta and fastq are identical OK"
    ((success++))
else
    log "OBICount: counting on fasta and fastq are different failed"
    ((failed++))
fi

((ntest++))
if diff "${TMPDIR}/wolf_F.fasta_count.csv" \
        "${TMPDIR}/wolf_F.csv_count.csv" > /dev/null
then
    log "OBICount: counting on fasta and csv are identical OK"
    ((success++))
else
    log "OBICount: counting on fasta and csv are different failed"
    ((failed++))
fi

#########################################
#
# At the end of the tests
# the cleanup function is called
#
#########################################

cleanup
