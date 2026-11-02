#!/bin/bash

#
# Here give the name of the test serie
#

TEST_NAME=obioligotag
CMD=obioligotag

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



#
# Every result is checked on its properties (size, alphabet, constraints),
# not against a reference file, because several equivalent families exist.
#

printf 'acgtac\nacgtat\nttttgg\nggcaac\ncaccat\nacgtac\n' > "${TMPDIR}/oligos.txt"
printf '>a\nACGTAC\n>b\nACGTAT\n>c\nTTTTGG\n>d\nGGCAAC\n>e\nCACCAT\n' > "${TMPDIR}/oligos.fasta"

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
if $CMD -s 5 -f 24 -d 3 -g 3 -p 2 > "${TMPDIR}/tags.fasta" 2> /dev/null \
   && [ "$(grep -vc '>' "${TMPDIR}/tags.fasta")" -ge 24 ] \
   && ! grep -v '>' "${TMPDIR}/tags.fasta" | grep -qE '(a{3}|c{3}|g{3}|t{3})'
then
    log "$MCMD: generating a family with homopolymer constraint OK" 
    ((success++))
else
    log "$MCMD: generating a family with homopolymer constraint failed" 
    ((failed++))
fi

((ntest++))
if $CMD -s 7 -f 60 -d 3 --alphabet ACT > "${TMPDIR}/noG.fasta" 2> /dev/null \
   && [ "$(grep -vc '>' "${TMPDIR}/noG.fasta")" -ge 60 ] \
   && ! grep -v '>' "${TMPDIR}/noG.fasta" | grep -q 'g'
then
    log "$MCMD: restricted alphabet OK" 
    ((success++))
else
    log "$MCMD: restricted alphabet failed" 
    ((failed++))
fi

((ntest++))
if $CMD -s 6 -f 20 -d 3 --distance-type lcs > "${TMPDIR}/lcs.fasta" 2> /dev/null \
   && [ "$(grep -vc '>' "${TMPDIR}/lcs.fasta")" -ge 20 ]
then
    log "$MCMD: LCS distance OK" 
    ((success++))
else
    log "$MCMD: LCS distance failed" 
    ((failed++))
fi

((ntest++))
if $CMD -s 7 -f 40 -d 3 -r cnnnnnn > "${TMPDIR}/rejected.fasta" 2> /dev/null \
   && [ "$(grep -vc '>' "${TMPDIR}/rejected.fasta")" -ge 40 ] \
   && ! grep -v '>' "${TMPDIR}/rejected.fasta" | grep -q '^c'
then
    log "$MCMD: rejected pattern OK" 
    ((success++))
else
    log "$MCMD: rejected pattern failed" 
    ((failed++))
fi

((ntest++))
if $CMD -L "${TMPDIR}/oligos.txt" -f 3 -d 3 > "${TMPDIR}/from_text.fasta" 2> /dev/null \
   && [ "$(grep -vc '>' "${TMPDIR}/from_text.fasta")" -ge 3 ]
then
    log "$MCMD: selecting oligonucleotides from a text file OK" 
    ((success++))
else
    log "$MCMD: selecting oligonucleotides from a text file failed" 
    ((failed++))
fi

((ntest++))
if $CMD -L "${TMPDIR}/oligos.fasta" -f 3 -d 3 > "${TMPDIR}/from_fasta.fasta" 2> /dev/null \
   && [ "$(grep -vc '>' "${TMPDIR}/from_fasta.fasta")" -ge 3 ]
then
    log "$MCMD: selecting oligonucleotides from a fasta file OK" 
    ((success++))
else
    log "$MCMD: selecting oligonucleotides from a fasta file failed" 
    ((failed++))
fi

((ntest++))
if ! $CMD -s 5 -f 100 -d 3 > /dev/null 2>&1
then
    log "$MCMD: family size beyond the Hamming bound refused OK" 
    ((success++))
else
    log "$MCMD: family size beyond the Hamming bound not refused" 
    ((failed++))
fi

((ntest++))
if ! $CMD -s 5 -L "${TMPDIR}/oligos.txt" -f 3 > /dev/null 2>&1
then
    log "$MCMD: --oligo-size with --oligo-list refused OK" 
    ((success++))
else
    log "$MCMD: --oligo-size with --oligo-list not refused" 
    ((failed++))
fi

#########################################
#
# At the end of the tests
# the cleanup function is called
#
#########################################

cleanup
