#!/bin/bash

#
# Here give the name of the test serie
#

TEST_NAME=obisuperkmer
CMD=obisuperkmer

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

# Test 1: Basic super k-mer extraction with default parameters
((ntest++))
if obisuperkmer "${TEST_DIR}/test_sequences.fasta" \
    > "${TMPDIR}/output_default.fasta" 2>&1
then
    log "$MCMD: basic extraction with default parameters OK"
    ((success++))
else
    log "$MCMD: basic extraction with default parameters failed"
    ((failed++))
fi

# Test 2: Verify output is not empty
((ntest++))
if [ -s "${TMPDIR}/output_default.fasta" ]
then
    log "$MCMD: output file is not empty OK"
    ((success++))
else
    log "$MCMD: output file is empty - failed"
    ((failed++))
fi

# Test 3: Count number of super k-mers extracted (should be > 0)
((ntest++))
num_sequences=$(grep -c "^>" "${TMPDIR}/output_default.fasta")
if [ "$num_sequences" -gt 0 ]
then
    log "$MCMD: extracted $num_sequences super k-mers OK"
    ((success++))
else
    log "$MCMD: no super k-mers extracted - failed"
    ((failed++))
fi

# Test 4: Verify super k-mers have required metadata attributes
((ntest++))
if grep -q "minimizer_value" "${TMPDIR}/output_default.fasta" && \
   grep -q "minimizer_seq" "${TMPDIR}/output_default.fasta" && \
   grep -q "parent_id" "${TMPDIR}/output_default.fasta"
then
    log "$MCMD: super k-mers contain required metadata OK"
    ((success++))
else
    log "$MCMD: super k-mers missing metadata - failed"
    ((failed++))
fi

# Test 5: Extract super k-mers with custom k and m parameters
((ntest++))
if obisuperkmer -k 15 -m 7 "${TEST_DIR}/test_sequences.fasta" \
    > "${TMPDIR}/output_k15_m7.fasta" 2>&1
then
    log "$MCMD: extraction with custom k=15, m=7 OK"
    ((success++))
else
    log "$MCMD: extraction with custom k=15, m=7 failed"
    ((failed++))
fi

# Test 6: Verify custom parameters in output metadata
((ntest++))
if grep -q '"k":15' "${TMPDIR}/output_k15_m7.fasta" && \
   grep -q '"m":7' "${TMPDIR}/output_k15_m7.fasta"
then
    log "$MCMD: custom parameters correctly set in metadata OK"
    ((success++))
else
    log "$MCMD: custom parameters not in metadata - failed"
    ((failed++))
fi

# Test 7: Test with different output format (FASTA output explicitly)
((ntest++))
if obisuperkmer --fasta-output -k 21 -m 11 \
    "${TEST_DIR}/test_sequences.fasta" \
    > "${TMPDIR}/output_fasta.fasta" 2>&1
then
    log "$MCMD: FASTA output format OK"
    ((success++))
else
    log "$MCMD: FASTA output format failed"
    ((failed++))
fi

# Test 8: Verify all super k-mers have superkmer in their ID
((ntest++))
if grep "^>" "${TMPDIR}/output_default.fasta" | grep -q "superkmer"
then
    log "$MCMD: super k-mer IDs contain 'superkmer' OK"
    ((success++))
else
    log "$MCMD: super k-mer IDs missing 'superkmer' - failed"
    ((failed++))
fi

# Test 9: Verify parent sequence IDs are preserved
((ntest++))
if grep -q "seq1" "${TMPDIR}/output_default.fasta" && \
   grep -q "seq2" "${TMPDIR}/output_default.fasta" && \
   grep -q "seq3" "${TMPDIR}/output_default.fasta"
then
    log "$MCMD: parent sequence IDs preserved OK"
    ((success++))
else
    log "$MCMD: parent sequence IDs not preserved - failed"
    ((failed++))
fi

# Test 10: Test with output file option
((ntest++))
if obisuperkmer -o "${TMPDIR}/output_file.fasta" \
    "${TEST_DIR}/test_sequences.fasta" 2>&1
then
    log "$MCMD: output to file with -o option OK"
    ((success++))
else
    log "$MCMD: output to file with -o option failed"
    ((failed++))
fi

# Test 11: Verify output file was created with -o option
((ntest++))
if [ -s "${TMPDIR}/output_file.fasta" ]
then
    log "$MCMD: output file created with -o option OK"
    ((success++))
else
    log "$MCMD: output file not created with -o option - failed"
    ((failed++))
fi

# Test 12: Verify super k-mers are shorter than or equal to parent sequences
((ntest++))
# Count nucleotides in input sequences (excluding headers)
input_bases=$(grep -v "^>" "${TEST_DIR}/test_sequences.fasta" | tr -d '\n' | wc -c)
# Count nucleotides in output sequences (excluding headers)
output_bases=$(grep -v "^>" "${TMPDIR}/output_default.fasta" | tr -d '\n' | wc -c)

if [ "$output_bases" -le "$input_bases" ]
then
    log "$MCMD: super k-mer total length <= input length OK"
    ((success++))
else
    log "$MCMD: super k-mer total length > input length - failed"
    ((failed++))
fi

#########################################
#
# At the end of the tests
# the cleanup function is called
#
#########################################

cleanup
