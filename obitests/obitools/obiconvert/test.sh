#!/bin/bash

#
# Here give the name of the test serie
#

TEST_NAME=obiconvert
CMD=obiconvert

######
#
# Some variable and function definitions: please don't change them
#
######
TEST_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"

if [ -z "$TEST_DIR" ] ; then
   TEST_DIR="."
fi

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
if obiconvert -Z "${TEST_DIR}/gbpln1088.4Mb.fasta.gz" \
                 > "${TMPDIR}/xxx.fasta.gz" && \
   zdiff "${TEST_DIR}/gbpln1088.4Mb.fasta.gz" \
                 "${TMPDIR}/xxx.fasta.gz" 
then
    log "$MCMD: converting large fasta file to fasta OK" 
    ((success++))
else
    log "$MCMD: converting large fasta file to fasta failed" 
    ((failed++))
fi
 
((ntest++))
if obiconvert -Z --fastq-output \
              "${TEST_DIR}/gbpln1088.4Mb.fasta.gz" \
                 > "${TMPDIR}/xxx.fastq.gz" && \
   obiconvert -Z --fasta-output \
              "${TMPDIR}/xxx.fastq.gz" \
              > "${TMPDIR}/yyy.fasta.gz" && \
   zdiff "${TEST_DIR}/gbpln1088.4Mb.fasta.gz" \
                 "${TMPDIR}/yyy.fasta.gz" 
then
    log "$MCMD: converting large file between fasta and fastq OK" 
    ((success++))
else
    log "$MCMD: converting large file between fasta and fastq failed" 
    ((failed++))
fi

#########################################
#
# At the end of the tests
# the cleanup function is called
#
#########################################

cleanup
