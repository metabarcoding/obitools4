#!/bin/bash

#
# Here give the name of the test serie
#

TEST_NAME=obialign
CMD=obialign

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
# Dataset used by the following tests: two families of similar sequences, a
# sequence identical to another one, and a sequence with an ambiguous
# nucleotide that is discarded.
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

cat > "${TMPDIR}/ref.fasta" <<DATA
>r1
acgtacgtacgtaacctagcctagcctagggatcgatcgatcgatcgaatcg
>r2
ttgcatgcatgctagctagctagcatgcatcgatcgatgctagctagtagct
>r3
ggggccccggggccccggggccccggggccccggggccccggggcccc
DATA

((ntest++))
if $CMD --max-cpu 2 "${TMPDIR}/data.fasta" > "${TMPDIR}/all.csv" 2> /dev/null &&
   [ "$(head -n 1 "${TMPDIR}/all.csv")" = "id1,id2,score" ] &&
   [ "$(tail -n +2 "${TMPDIR}/all.csv" | wc -l | tr -d ' ')" = "10" ]
then
    log "$MCMD: without threshold all the pairs of the 5 usable sequences are written OK"
    ((success++))
else
    log "$MCMD: the table without threshold is wrong"
    ((failed++))
fi

((ntest++))
if $CMD --max-cpu 2 -t 0.95 "${TMPDIR}/data.fasta" > "${TMPDIR}/t95.csv" 2> /dev/null &&
   [ "$(tail -n +2 "${TMPDIR}/t95.csv" | cut -d, -f1,2 | tr '\n' ' ')" = "a1,a2 a1,a3 a2,a3 b1,b2 " ] &&
   grep -q '^a1,a3,1$' "${TMPDIR}/t95.csv"
then
    log "$MCMD: only the pairs reaching the threshold are written, in order OK"
    ((success++))
else
    log "$MCMD: the table with a threshold is wrong"
    ((failed++))
fi

((ntest++))
if $CMD --max-cpu 2 -t 0.95 -x "${TMPDIR}/data.fasta" 2> /dev/null | head -n 2 | tail -n 1 | grep -q '^a1,a2,[0-9.]*,20,3,52,52$'
then
    log "$MCMD: extra columns OK"
    ((success++))
else
    log "$MCMD: extra columns failed"
    ((failed++))
fi

((ntest++))
if $CMD --max-cpu 2 -g -t 0.9 "${TMPDIR}/data.fasta" 2> /dev/null | grep -q '^b1,x1,'
then
    log "$MCMD: a sequence with an IUPAC letter is kept with --accept-ambiguous OK"
    ((success++))
else
    log "$MCMD: --accept-ambiguous failed"
    ((failed++))
fi

((ntest++))
if $CMD --max-cpu 2 -R "${TMPDIR}/ref.fasta" -t 0.95 "${TMPDIR}/data.fasta" 2> /dev/null | tail -n +2 | cut -d, -f1,2 | tr '\n' ' ' > "${TMPDIR}/two.txt" &&
   [ "$(cat "${TMPDIR}/two.txt")" = "a1,r1 a2,r1 a3,r1 b1,r2 b2,r2 " ]
then
    log "$MCMD: comparison with a reference file OK"
    ((success++))
else
    log "$MCMD: comparison with a reference file failed"
    ((failed++))
fi

((ntest++))
if $CMD --max-cpu 1 "${TMPDIR}/data.fasta" 2> /dev/null > "${TMPDIR}/cpu1.csv" &&
   $CMD --max-cpu 8 "${TMPDIR}/data.fasta" 2> /dev/null > "${TMPDIR}/cpu8.csv" &&
   cmp -s "${TMPDIR}/cpu1.csv" "${TMPDIR}/cpu8.csv"
then
    log "$MCMD: the table does not depend on the number of threads OK"
    ((success++))
else
    log "$MCMD: the table depends on the number of threads"
    ((failed++))
fi


#########################################
#
# At the end of the tests
# the cleanup function is called
#
#########################################

cleanup
