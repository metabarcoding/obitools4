#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include "libstki.h"
#include "apat.h"

#include "obiapat.h"

static void EncodeSequence(SeqPtr seq);
static void UpperSequence(char *seq);

/*
 * print the message given as argument and exit the program
 * @param error		error number	
 * @param message 	the text explaining what's going on
 * @param filename	the file source where the program failed
 * @param linenumber	the line where it has failed
 * filename and linenumber are written at pre-processing 
 * time by a macro
 */
void* ecoError(int error,
               const char* message,
               const char * filename,
               int linenumber,
			   int *errno,
			   char **error_msg)
{
	*error_msg = malloc(1001);
	snprintf(*error_msg,1000, 
	         "Error %d in file %s line %d : %s", 
	               error,
	               filename,
	               linenumber,
	               message);

	*errno = error;
	return NULL;
}



/*
 * @doc: DNA alphabet (IUPAC)
 */
#define LX_BIO_DNA_ALPHA   "ABCDEFGHIJKLMNOPQRSTUVWXYZ#![]"

/*
 * @doc: complementary DNA alphabet (IUPAC)
 */
#define LX_BIO_CDNA_ALPHA  "TVGHEFCDIJMLKNOPQYSAABWXRZ#!]["


static char sNuc[]     = LX_BIO_DNA_ALPHA;
static char sAnuc[]    = LX_BIO_CDNA_ALPHA;

static char LXBioBaseComplement(char nucAc);
static char *LXBioSeqComplement(char *nucAcSeq);
static char *reverseSequence(char *str,char isPattern);

 
/* ---------------------------- */

char LXBioBaseComplement(char nucAc)
{
    char *c;

    if ((c = strchr(sNuc, nucAc)))
        return sAnuc[(c - sNuc)];
    else
        return nucAc;
}

/* ---------------------------- */

char *LXBioSeqComplement(char *nucAcSeq)
{
    char *s;

    for (s = nucAcSeq ; *s ; s++)
        *s = LXBioBaseComplement(*s);

    return nucAcSeq;
}


char *reverseSequence(char *str,char isPattern)
{
        char *sb, *se, c;

        if (! str)
            return str;
            
        sb = str;
        se = str + strlen(str) - 1;

        while(sb <= se) {
           c    = *sb;
          *sb++ = *se;
          *se-- = c;
        }

		sb = str;
		se = str + strlen(str) - 1;
		
		if (isPattern)
			for (;sb <= se; sb++)
			{
				if (*sb=='#')
				{
					if (*(sb+1) == '[') {
						while(*sb !=']') {
							*sb = *(sb+1);
							sb++;
						}
						*sb='#';
					} else {
					if (((se - sb) > 2) && (*(sb+2)=='!'))
					{
						*sb='!';
						sb+=2;
						*sb='#';
					}
					else
					{
						*sb=*(sb+1);
						sb++;
						*sb='#';
					}}
				}
				else if (*sb=='!')
					{
						*sb=*(sb-1);
						*(sb-1)='!';
					}
			}

        return str;
}

char *ecoComplementPattern(char *nucAcSeq)
{
    return reverseSequence(LXBioSeqComplement(nucAcSeq),1);
}

char *ecoComplementSequence(char *nucAcSeq)
{
    return reverseSequence(LXBioSeqComplement(nucAcSeq),0);
}


char *getSubSequence(char* nucAcSeq,int32_t begin,int32_t end, 
					int *errno, char **errmsg)
/*
   extract subsequence from nucAcSeq [begin,end[
*/
{
	static char *buffer  = NULL;
	static int32_t buffSize= 0;
	int32_t length;
	
	if (begin < end)
	{
		length = end - begin;
		
		if (length >= buffSize)
		{
			buffSize = length+1;
			if (buffer)
				buffer=ECOREALLOC(buffer,buffSize,
						   	      "Error in reallocating sub sequence buffer",errno,errmsg);
			else
				buffer=ECOMALLOC(buffSize,
				          		 "Error in allocating sub sequence buffer",errno,errmsg);
				
		}
		
		strncpy(buffer,nucAcSeq + begin,length);
		buffer[length]=0;
	}
	else
	{
		length = end + strlen(nucAcSeq) - begin;
		
		if (length >= buffSize)
		{
			buffSize = length+1;
			if (buffer)
				buffer=ECOREALLOC(buffer,buffSize,
						   	      "Error in reallocating sub sequence buffer",errno,errmsg);
			else
				buffer=ECOMALLOC(buffSize,
				          		 "Error in allocating sub sequence buffer",errno,errmsg);
				
		}
		strncpy(buffer,nucAcSeq+begin,length - end);
		strncpy(buffer+(length-end),nucAcSeq ,end);
		buffer[length]=0;
	}
	
	return buffer;
}


/* -------------------------------------------- */
/* uppercase sequence                           */
/* -------------------------------------------- */

#define IS_LOWER(c) (((c) >= 'a') && ((c) <= 'z'))
#define TO_UPPER(c) ((c) - 'a' + 'A')

void UpperSequence(char *seq)
{
        char *cseq;

        for (cseq = seq ; *cseq ; cseq++) 
            if (IS_LOWER(*cseq))
                *cseq = TO_UPPER(*cseq);
}
 
#undef IS_LOWER
#undef TO_UPPER




/* -------------------------------------------- */
/* encode sequence                              */
/* IS_UPPER is slightly faster than isupper     */
/* -------------------------------------------- */

#define IS_UPPER(c) (((c) >= 'A') && ((c) <= 'Z'))



void EncodeSequence(SeqPtr seq)
{
        int   i;
        uint8_t *data;
        char  *cseq;
		char nuc;

        data = seq->data;
        cseq = seq->cseq;

        while (*cseq) {
			nuc = *cseq & (~32);
            *data = (IS_UPPER(nuc) ? nuc - 'A' : 0x0);
            data++;
            cseq++;
        }
        
        for (i=0,cseq=seq->cseq;i < seq->circular; i++,cseq++,data++) {
			nuc = *cseq & (~32);
            *data = (IS_UPPER(nuc) ? nuc - 'A' : 0x0);
		}

        for (i = 0 ; i < MAX_PATTERN ; i++)
            seq->hitpos[i]->top = seq->hiterr[i]->top = 0;

}

#undef IS_UPPER


SeqPtr new_apatseq(const char *in,int32_t circular, int32_t seqlen,
                    SeqPtr out, 
					int *errno, char **errmsg)
{
        int    i;

		if (circular != 0) circular=MAX_PAT_LEN;

		if (!out)
		{
			out = ECOMALLOC(sizeof(Seq),
			                "Error in Allocation of a new Seq structure",errno,errmsg);
			          
	        for (i  = 0 ; i < MAX_PATTERN ; i++) 
	        {
			   
	           if (! (out->hitpos[i] = NewStacki(kMinStackiSize))) 
	             	ECOERROR(ECO_MEM_ERROR,"Error in hit stack Allocation",errno,errmsg);
	
	           if (! (out->hiterr[i] = NewStacki(kMinStackiSize)))
	            	ECOERROR(ECO_MEM_ERROR,"Error in error stack Allocation",errno,errmsg);
	        }
		}

		
		out->seqsiz = out->seqlen = seqlen;
		out->circular = circular;
		
		if (!out->data)
		{
			out->data = ECOMALLOC((out->seqlen+circular) *sizeof(uint8_t),
		    	     			  "Error in Allocation of a new Seq data member",
								   errno,errmsg);    
		   	out->datsiz=  out->seqlen+circular;
		}
		else if ((out->seqlen +circular) >= out->datsiz)
		{
			out->data = ECOREALLOC(out->data,(out->seqlen+circular) *sizeof(uint8_t),
			                      "Error during Seq data buffer realloc",
								  errno,errmsg);
		   	out->datsiz=  out->seqlen+circular;			                      
		}

		out->cseq = (char *)in;
		
		EncodeSequence(out);

        return out;
}

int32_t delete_apatseq(SeqPtr pseq, 
					int *errno, char **errmsg)
{
         int i;

        if (pseq) {

            if (pseq->data) 
            	ECOFREE(pseq->data,"Freeing sequence data buffer",
				errno,errmsg);

            for (i = 0 ; i < MAX_PATTERN ; i++) {
                if (pseq->hitpos[i]) FreeStacki(pseq->hitpos[i]);
                if (pseq->hiterr[i]) FreeStacki(pseq->hiterr[i]);
            }

            ECOFREE(pseq,"Freeing apat sequence structure",
			errno,errmsg);
            
            return 0;
        }
        
        return 1;
}

PatternPtr buildPattern(const char *pat, int32_t error_max, 
						int *errno, char **errmsg)
{
	PatternPtr pattern;
	int32_t    patlen;
	int32_t    patlen2;

	patlen  = strlen(pat);
	patlen2 = lenPattern(pat);

	pattern = ECOMALLOC(sizeof(Pattern) +                   // Space for struct Pattern
							sizeof(char)*patlen+1 +         // Space for cpat
							sizeof(uint32_t) * patlen2 +    // Space for patcode
							sizeof(patword_t) * ALPHA_LEN , // Space for smat
						"Error in pattern allocation",
						errno,errmsg);
						
	pattern->ok      = true;
	pattern->hasIndel= false;
	pattern->maxerr  = error_max;
	
	pattern->cpat    = (char*)pattern + sizeof(Pattern);
	pattern->patcode = (uint32_t*)(pattern->cpat + patlen + 1); 
	pattern->smat    = (patword_t*)(pattern->patcode + patlen2);
	                             
	strncpy(pattern->cpat,pat,patlen);
	pattern->cpat[patlen]=0;
	UpperSequence(pattern->cpat);
	
	if (!CheckPattern(pattern))
		ECOERROR(ECO_ASSERT_ERROR,"Error in pattern checking",errno,errmsg);
		
	if (! EncodePattern(pattern, dna))
		ECOERROR(ECO_ASSERT_ERROR,"Error in pattern encoding",errno,errmsg);

   	if (! CreateS(pattern, ALPHA_LEN))
		ECOERROR(ECO_ASSERT_ERROR,"Error in pattern compiling",errno,errmsg);
	
	return pattern;
		
}

PatternPtr complementPattern(PatternPtr pat, int *errno, 
								char **errmsg)
{
	PatternPtr pattern;
	
	pattern = ECOMALLOC(sizeof(Pattern) + 
							sizeof(char)      * strlen(pat->cpat) + 1 +
							sizeof(uint32_t)  * pat->patlen  +
							sizeof(patword_t) * ALPHA_LEN,
						"Error in pattern allocation",
						errno,errmsg);
						
	pattern->ok      = true;
	pattern->hasIndel= pat->hasIndel;
	pattern->maxerr  = pat->maxerr;
	pattern->patlen  = pat->patlen;

	pattern->cpat    = (char*)pattern + sizeof(Pattern);
	pattern->patcode = (uint32_t*)(pattern->cpat + strlen(pat->cpat) + 1); 
	pattern->smat    = (patword_t*)(pattern->patcode + pat->patlen);
	                             
	strcpy(pattern->cpat,pat->cpat);
	
	ecoComplementPattern(pattern->cpat);
	
	if (!CheckPattern(pattern))
		ECOERROR(ECO_ASSERT_ERROR,"Error in pattern checking",errno,errmsg);
		
	if (! EncodePattern(pattern, dna))
		ECOERROR(ECO_ASSERT_ERROR,"Error in pattern encoding",errno,errmsg);

   	if (! CreateS(pattern, ALPHA_LEN))
		ECOERROR(ECO_ASSERT_ERROR,"Error in pattern compiling",errno,errmsg);
	
	return pattern;
		
}
