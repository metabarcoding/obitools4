#!/bin/bash

#
# Here give the name of the test serie
#

TEST_NAME=obipairing
CMD=obipairing

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
if obipairing -F "${TEST_DIR}/wolf_F.fastq.gz" \
              -R "${TEST_DIR}/wolf_R.fastq.gz" \
    | obidistribute -Z -c mode \
                    -p "${TMPDIR}/wolf_paired_%s.fastq.gz" 
then
    log "OBIPairing: sequence pairing OK" 
    ((success++))
else
    log "OBIPairing: sequence pairing failed" 
    ((failed++))
fi

((ntest++))
if obicsv -Z -s -i \
          -k ali_dir -k ali_length -k pairing_fast_count \
          -k pairing_fast_overlap -k pairing_fast_score \
          -k score -k score_norm -k seq_a_single \
          -k seq_b_single -k seq_ab_match \
          "${TMPDIR}/wolf_paired_alignment.fastq.gz" \
    > "${TMPDIR}/wolf_paired_alignment.csv.gz" \
    && zdiff -c "${TEST_DIR}/wolf_paired_alignment.csv.gz" \
                "${TMPDIR}/wolf_paired_alignment.csv.gz" 
then
    log "OBIPairing: check aligned sequences OK" 
    ((success++))
else
    log "OBIPairing: check aligned sequences failed" 
    ((failed++))
fi

((ntest++))
if obicsv -Z -s -i \
          "${TMPDIR}/wolf_paired_join.fastq.gz" \
    > "${TMPDIR}/wolf_paired_join.csv.gz" \
    && zdiff -c "${TEST_DIR}/wolf_paired_join.csv.gz" \
                "${TMPDIR}/wolf_paired_join.csv.gz"
then
    log "OBIPairing: check joined sequences OK" 
    ((success++))
else
    log "OBIPairing: check joined sequences failed" 
    ((failed++))
fi

#########################################
#
# At the end of the tests
# the cleanup function is called
#
#########################################

cleanup
