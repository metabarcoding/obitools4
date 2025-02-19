#!/bin/bash -e

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
}

log() {
    echo "[$TEST_NAME @ $(date)] $*" 1>&2
}

trap cleanup EXIT ERR SIGINT SIGTERM

echo "Testing $TEST_NAME..." 1>&2


#################################################
####
#### Below are the tests
####
#### Before each test :
####  - increment the variable ntest
####
#### Run the command as the condition of a if / then /else
####
#### then clause is executed on success of the command
####  - Write a success message using the log function
####  - increment the variable success
####
#### else clause is executed on failure of the command
####  - Write a failure message using the log function
####  - increment the variable failed
####
#################################################

((ntest++))
if obicount "${TEST_DIR}/wolf_F.fasta.gz" \
    > "${TMPDIR}/wolf_F.fasta_count.csv" ; then
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
