#include "obiapat.h"
#include <stdlib.h>

static int eco_log_malloc = 0;

void    eco_trace_memory_allocation()
{
	eco_log_malloc=1;
}

void    eco_untrace_memory_allocation()
{
	eco_log_malloc=0;
}


void   *eco_malloc(int32_t chunksize,
                   const char *error_message,
                   const char *filename,
                   int32_t    line, 
				   int *errno, char **errmsg)
{
	void * chunk;
	
	chunk = calloc(1,chunksize);
	
	if (!chunk)
		ecoError(ECO_MEM_ERROR,error_message,filename,line,errno,errmsg);
		
	if (eco_log_malloc)
		fprintf(stderr,
			    "Memory segment located at %p of size %d is allocated (file : %s [%d])",
			    chunk,
			    chunksize,
			    filename,
			    line);
		
	return chunk;
}

void   *eco_realloc(void *chunk,
                    int32_t newsize,
                    const char *error_message,
                    const char *filename,
                    int32_t    line, 
				   int *errno, char **errmsg)
{
	void *newchunk;
	
	newchunk = realloc(chunk,newsize);
	
	if (!newchunk)
		ecoError(ECO_MEM_ERROR,error_message,filename,line,errno,errmsg);

	if (eco_log_malloc)
		fprintf(stderr,
			    "Old memory segment %p is reallocated at %p with a size of %d (file : %s [%d])",
			    chunk,
			    newchunk,
			    newsize,
			    filename,
			    line);
		
	return newchunk;	
}

void    eco_free(void *chunk,
                 const char *error_message,
                 const char *filename,
                 int32_t    line, 
				   int *errno, char **errmsg)
{
	free(chunk);
	
	if (eco_log_malloc)
		fprintf(stderr,
			    "Memory segment %p is released => %s (file : %s [%d])",
			    chunk,
			    error_message,
			    filename,
			    line);
}
