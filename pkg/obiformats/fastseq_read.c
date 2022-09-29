#include "fastseq_read.h"


static fast_kseq_t* _open_fast_sek(gzFile fp, int shift) {
    fast_kseq_t* iterator;

    iterator = (fast_kseq_t*)malloc(sizeof(fast_kseq_t));


    if (iterator == NULL)
        return NULL;

    iterator->filez = fp;
    iterator->finished = false;
    iterator->shift = shift;

    if (fp != Z_NULL) {
        iterator->seq = kseq_init(fp);

        if (iterator->seq == NULL) {
            free(iterator);
            iterator=NULL;
        }
    }
    else {
        free(iterator);
        iterator=NULL;
    }

    return iterator;
}

/**
 * @brief open a FastA or FastQ file gizzed or not 
 * 
 * @param filename a const char* indicating the path of the 
 *        fast* file
 * @return kseq_t* a pointer to a kseq_t structure or NULL on
 *         failing
 */
fast_kseq_t* open_fast_sek_file(const char* filename, int shift) {
    gzFile fp;  

    fp = gzopen(filename, "r");  
    return _open_fast_sek(fp, shift);
}

fast_kseq_p open_fast_sek_fd(int fd, bool keep_open, int shift) {
    gzFile fp;  

    if (keep_open)
        fd = dup(fd);

    fp = gzdopen(fd, "r"); 

    return _open_fast_sek(fp, shift);
}

fast_kseq_p open_fast_sek_stdin(int shift) {
    return open_fast_sek_fd(fileno(stdin), true, shift);
}


int64_t next_fast_sek(fast_kseq_t* iterator) {
    int64_t l;

    if (iterator == NULL || iterator->seq == NULL)
        return -3;

    l = kseq_read(iterator->seq);
    if (l < 0) l = 0;

    iterator->finished = l==0;
    if (l>0) l = gzoffset(iterator->filez);

    return l;
}

int rewind_fast_sek(fast_kseq_t* iterator) {
    if (iterator == NULL || iterator->seq == NULL)
        return -3;

    kseq_rewind(iterator->seq);
    return 0;
}

int close_fast_sek(fast_kseq_t* iterator) {
    gzFile fp;  
    kseq_t *seq;
    int rep = -3;

    if (iterator == NULL)
        return rep;

    fp  = iterator->filez;
    seq = iterator->seq;

    free(iterator);

    if (seq != NULL)
        kseq_destroy(iterator->seq);

    if (fp != Z_NULL)
        rep = gzclose(fp);

    return rep;
}
    
