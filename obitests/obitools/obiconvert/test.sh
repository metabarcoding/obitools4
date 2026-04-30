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


# ------------------------------------------------------------------
# --raw-taxid tests (no taxonomy loaded)
# ------------------------------------------------------------------

# Running test
((ntest++))
if obiconvert --raw-taxid "${TEST_DIR}/out_ecotag.fasta" \
              > "${TMPDIR}/raw_taxid.fasta" 2>/dev/null
then
    log "$MCMD --raw-taxid: running OK"
    ((success++))
else
    log "$MCMD --raw-taxid: running failed"
    ((failed++))
fi

# Taxids must be bare numbers — no full-format "taxon:ID [Name]@rank" strings
((ntest++))
if grep '"taxid"' "${TMPDIR}/raw_taxid.fasta" | grep -qv '"taxid":"[0-9][0-9]*"'
then
    log "$MCMD --raw-taxid: taxid format check failed (full-format taxid found)"
    ((failed++))
else
    log "$MCMD --raw-taxid: taxid format OK (all taxids are bare numbers)"
    ((success++))
fi

# --raw-taxid is idempotent: piping through a second obiconvert --raw-taxid must
# produce bit-for-bit identical output.
((ntest++))
if obiconvert --raw-taxid "${TMPDIR}/raw_taxid.fasta" \
              > "${TMPDIR}/raw_taxid2.fasta" 2>/dev/null
then
    log "$MCMD --raw-taxid piped: running OK"
    ((success++))
else
    log "$MCMD --raw-taxid piped: running failed"
    ((failed++))
fi

((ntest++))
if diff "${TMPDIR}/raw_taxid.fasta" \
        "${TMPDIR}/raw_taxid2.fasta" > /dev/null
then
    log "$MCMD --raw-taxid piped: idempotency OK"
    ((success++))
else
    log "$MCMD --raw-taxid piped: idempotency failed (outputs differ)"
    ((failed++))
fi


# ------------------------------------------------------------------
# --taxonomy tests (full-format taxid, no --raw-taxid)
# ------------------------------------------------------------------

# Running test
((ntest++))
if obiconvert --taxonomy "${TEST_DIR}/taxonomy.csv" \
              "${TEST_DIR}/out_ecotag.fasta" \
              > "${TMPDIR}/taxo.fasta" 2>/dev/null
then
    log "$MCMD --taxonomy: running OK"
    ((success++))
else
    log "$MCMD --taxonomy: running failed"
    ((failed++))
fi

# Taxids must be in full "taxon:ID [Name]@rank" format
((ntest++))
if grep '"taxid"' "${TMPDIR}/taxo.fasta" | grep -q '"taxid":"taxon:[0-9]'
then
    log "$MCMD --taxonomy: taxid format OK (full-format taxids present)"
    ((success++))
else
    log "$MCMD --taxonomy: taxid format check failed (no full-format taxid found)"
    ((failed++))
fi


# ------------------------------------------------------------------
# --raw-taxid --taxonomy tests
# ------------------------------------------------------------------

# Running test
((ntest++))
if obiconvert --raw-taxid --taxonomy "${TEST_DIR}/taxonomy.csv" \
              "${TEST_DIR}/out_ecotag.fasta" \
              > "${TMPDIR}/raw_taxid_taxo.fasta" 2>/dev/null
then
    log "$MCMD --raw-taxid --taxonomy: running OK"
    ((success++))
else
    log "$MCMD --raw-taxid --taxonomy: running failed"
    ((failed++))
fi

# Taxids must be bare numbers even when taxonomy is loaded
((ntest++))
if grep '"taxid"' "${TMPDIR}/raw_taxid_taxo.fasta" | grep -qv '"taxid":"[0-9][0-9]*"'
then
    log "$MCMD --raw-taxid --taxonomy: taxid format check failed (full-format taxid found)"
    ((failed++))
else
    log "$MCMD --raw-taxid --taxonomy: taxid format OK (all taxids are bare numbers)"
    ((success++))
fi

# --raw-taxid with or without taxonomy must yield identical taxid values
((ntest++))
if diff <(grep '"taxid"' "${TMPDIR}/raw_taxid.fasta" | grep -o '"taxid":"[^"]*"' | sort) \
        <(grep '"taxid"' "${TMPDIR}/raw_taxid_taxo.fasta" | grep -o '"taxid":"[^"]*"' | sort) \
        > /dev/null
then
    log "$MCMD --raw-taxid vs --raw-taxid --taxonomy: taxid values match OK"
    ((success++))
else
    log "$MCMD --raw-taxid vs --raw-taxid --taxonomy: taxid values differ (unexpected)"
    ((failed++))
fi


#########################################
#
# At the end of the tests
# the cleanup function is called
#
#########################################

cleanup
