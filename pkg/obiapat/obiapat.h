#ifndef __obiapat_h__
#define __obiapat_h__

#include <stdio.h>
#include <stdint.h>

#include "apat.h"

/*****************************************************
 * 
 *  Data type declarations
 * 
 *****************************************************/

/*
 * 
 *  Sequence types
 * 
 */

typedef struct {
	
	int32_t  taxid;
	char     AC[20];
	int32_t  DE_length;
	int32_t  SQ_length;
	int32_t  CSQ_length;
	
	char     data[1];
	
} ecoseqformat_t;

typedef struct {
	int32_t taxid;
	int32_t SQ_length;
	char    *AC;
	char    *DE;
	char    *SQ;
} ecoseq_t;



/*****************************************************
 * 
 *  Function declarations
 * 
 *****************************************************/

void* ecoError(int error,
               const char* message,
               const char * filename,
               int linenumber,
			   int *errno,
			   char **error_msg);

#define ECOERROR(code,message,errno,errmsg) \
    { return ecoError((code),(message),__FILE__,__LINE__,errno,errmsg); }

#define ECO_IO_ERROR       (1)
#define ECO_MEM_ERROR      (2)
#define ECO_ASSERT_ERROR   (3)
#define ECO_NOTFOUND_ERROR (4)


/*
 * 
 * Low level system functions
 * 
 */

int32_t is_big_endian();
int32_t swap_int32_t(int32_t);

void   *eco_malloc(int32_t chunksize,
                   const char *error_message,
                   const char *filename,
                   int32_t    line, 
				   int *errno, char **errmsg);
                   
                   
void   *eco_realloc(void *chunk,
                    int32_t chunksize,
                    const char *error_message,
                    const char *filename,
                    int32_t    line, 
				   int *errno, char **errmsg);
                    
void    eco_free(void *chunk,
                 const char *error_message,
                 const char *filename,
                 int32_t    line, 
				 int *errno, char **errmsg);
                 
void    eco_trace_memory_allocation();
void    eco_untrace_memory_allocation();

#define ECOMALLOC(size,error_message,errno,errmsg) \
	    eco_malloc((size),(error_message),__FILE__,__LINE__,errno,errmsg)
	   
#define ECOREALLOC(chunk,size,error_message,errno,errmsg) \
        eco_realloc((chunk),(size),(error_message),__FILE__,__LINE__,errno,errmsg)
        
#define ECOFREE(chunk,error_message,errno,errmsg) \
        eco_free((chunk),(error_message),__FILE__,__LINE__,errno,errmsg)
        



ecoseq_t *new_ecoseq();
int32_t   delete_ecoseq(ecoseq_t *);
ecoseq_t *new_ecoseq_with_data( char *AC,
								char *DE,
								char *SQ,
								int32_t   taxid
								);



int32_t  delete_apatseq(Seq *pseq, 
				   int *errno, char **errmsg);
Pattern *buildPattern(const char *pat, int32_t error_max, uint8_t hasIndel, int *errno, char **errmsg);
Pattern *complementPattern(Pattern *pat, int *errno, char **errmsg);

Seq *new_apatseq(const char *in,int32_t circular, int32_t seqlen,
                    Seq *out, 
					int *errno, char **errmsg);
					
char *ecoComplementPattern(char *nucAcSeq);
char *ecoComplementSequence(char *nucAcSeq);
char *getSubSequence(char* nucAcSeq,int32_t begin,int32_t end, 
				   int *errno, char **errmsg);


#endif /* __obiapat_h__ */
