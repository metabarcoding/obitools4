#!/bin/bash

#
# Here give the name of the test serie
#

TEST_NAME=obiclust
CMD=obiclust

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


#
# Dataset used by the following tests: two families with a rare variant each,
# a duplicate, and a sequence with an ambiguous nucleotide that is discarded.
#
cat > "${TMPDIR}/data.fasta" <<DATA
>a1 count=20;
acgtacgtacgtaacctagcctagcctagggatcgatcgatcgatcgaatcg
>a2 count=3;
acgtacgtacgtaacctagcctagcctagggatcgatcgatcgttcgaatcg
>a3 count=2;
acgtacgtacgtaacctagcctagcctagggatcgatcgatcgatcgaatcg
>b1 count=15;
ttgcatgcatgctagctagctagcatgcatcgatcgatgctagctagtagct
>b2 count=1;
ttgcatgcatgctagctagctagcatgcatcgatcgatgctagctagtaact
>x1 count=4;
ttgcatgcatgctangctagctagcatgcatcgatcgatgctagctagtagc
DATA

((ntest++))
if $CMD --max-cpu 2 -O -t 0.95 "${TMPDIR}/data.fasta" > "${TMPDIR}/clusters.fasta" 2> "${TMPDIR}/clusters.log"
then
    log "$MCMD: clustering runs OK"
    ((success++))
else
    log "$MCMD: clustering failed"
    ((failed++))
fi

((ntest++))
if [ "$(grep -c '^>' "${TMPDIR}/clusters.fasta")" = "5" ]
then
    log "$MCMD: the sequence with an ambiguous nucleotide is discarded OK"
    ((success++))
else
    log "$MCMD: wrong number of sequences in the output"
    ((failed++))
fi

((ntest++))
if [ "$(grep '^>a[23] ' "${TMPDIR}/clusters.fasta" | grep -c 'cluster=a1;')" = "2" ] &&
   [ "$(grep '^>b2 ' "${TMPDIR}/clusters.fasta" | grep -c 'cluster=b1;')" = "1" ] &&
   grep '^>b1 ' "${TMPDIR}/clusters.fasta" | grep -q 'cluster_center=true;' &&
   grep '^>a1 ' "${TMPDIR}/clusters.fasta" | grep -q 'cluster_weight=25;'
then
    log "$MCMD: sequences are attached to the right center OK"
    ((success++))
else
    log "$MCMD: sequences are not attached to the right center"
    ((failed++))
fi

((ntest++))
if $CMD --max-cpu 2 -t 0.95 --otu-map "${TMPDIR}/otu.csv" --no-sequence-output "${TMPDIR}/data.fasta" > "${TMPDIR}/none.fasta" 2> /dev/null &&
   [ ! -s "${TMPDIR}/none.fasta" ] &&
   [ "$(head -n 1 "${TMPDIR}/otu.csv")" = "id,cluster,cluster_center,cluster_score,count" ] &&
   [ "$(sed -n 2,4p "${TMPDIR}/otu.csv" | cut -d, -f1,2 | tr '\n' ' ')" = "a1,a1 a3,a1 a2,a1 " ] &&
   [ "$(grep -c ',true,' "${TMPDIR}/otu.csv")" = "2" ]
then
    log "$MCMD: CSV table of the clusters OK"
    ((success++))
else
    log "$MCMD: CSV table of the clusters failed"
    ((failed++))
fi

((ntest++))
if $CMD --max-cpu 1 -O -t 0.95 "${TMPDIR}/data.fasta" 2> /dev/null | grep '^>' | sed 's/ .*//' > "${TMPDIR}/order1.txt" &&
   $CMD --max-cpu 8 -O -t 0.95 "${TMPDIR}/data.fasta" 2> /dev/null | grep '^>' | sed 's/ .*//' > "${TMPDIR}/order8.txt" &&
   cmp -s "${TMPDIR}/order1.txt" "${TMPDIR}/order8.txt"
then
    log "$MCMD: the result does not depend on the number of threads OK"
    ((success++))
else
    log "$MCMD: the result depends on the number of threads"
    ((failed++))
fi


#########################################
#
# At the end of the tests
# the cleanup function is called
#
#########################################

cleanup
