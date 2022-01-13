#ifndef _READ_H
#define _READ_H

#include <zlib.h>
#include <stdio.h>
#include <stdint.h>
#include <stdbool.h> 

#include "kseq/kseq.h"

KSEQ_INIT(gzFile, gzread)

typedef struct {
  kseq_t  *seq;
  bool    finished;
  int16_t shift;
  gzFile filez;
} fast_kseq_t, *fast_kseq_p;
 

fast_kseq_t* open_fast_sek_file(const char* filename, int shift);
fast_kseq_t* open_fast_sek_fd(int fd, bool keep_open, int shift);
fast_kseq_t* open_fast_sek_stdin(int shift);

/**
 * @brief read the next sequence on the fast* stream
 * 
 * @param seq a kseq_t* created using function open_fast_sek
 * @return int if greater than 0 represents the length of the
 *             sequence, otherwise indicates an error
 *              - -1 : no more sequence in the stream
 *              - -2 : too short quality sequence
 *              - -3 : called with NULL pointer
 */ 
int64_t next_fast_sek(fast_kseq_t* iterator);


int close_fast_sek(fast_kseq_t* iterator);
int rewind_fast_sek(fast_kseq_t* iterator);

#endif