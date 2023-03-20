/* ==================================================== */
/*      Copyright (c) Atelier de BioInformatique        */
/*      Dec. 94                                         */
/*      File: apat.h                                    */
/*      Purpose: pattern scan                           */
/*      History:                                        */
/*      28/12/94 : <Gloup> ascan first version          */
/*      14/05/99 : <Gloup> last revision                */
/*      07/12/21 : <Zafacs> last some cleaning for 2020 */
/* ==================================================== */


#ifndef H_apat
#define H_apat

#include <stdio.h>
#include "libstki.h" 


/* ----------------------------------------------- */
/* constantes                                      */
/* ----------------------------------------------- */

#ifndef BUFSIZ
#define BUFSIZ          1024    /* io buffer size               */
#endif

#define MAX_NAME_LEN    BUFSIZ  /* max length of sequence name  */

#define ALPHA_LEN        26     /* alphabet length              */
                                /* *DO NOT* modify              */

#define MAX_PATTERN       1     /* max # of patterns            */
                                /* *DO NOT* modify              */

#define MAX_PAT_LEN      64     /* max pattern length           */
                                /* *DO NOT* modify              */

#define MAX_PAT_ERR      64     /* max # of errors              */
                                /* *DO NOT* modify              */

#define PATMASK 0x3ffffff       /* mask for 26 symbols          */
                                /* *DO NOT* modify              */

#define OBLIBIT 0x4000000       /* bit 27 to 1 -> oblig. pos    */
                                /* *DO NOT* modify              */

                                /* mask for position            */
#define ONEMASK   0x8000000000000000      /* mask for highest position    */

                                /* masks for Levenhstein edit   */
#define OPER_IDT  0x0000000000000000    /* identity                     */
#define OPER_INS  0x4000000000000000    /* insertion                    */
#define OPER_DEL  0x8000000000000000    /* deletion                     */
#define OPER_SUB  0xc000000000000000    /* substitution                 */

#define OPER_SHFT 30            /* <unused> shift               */

                                /* Levenhstein Opcodes          */
#define SOPER_IDT 0x0           /* identity                     */
#define SOPER_INS 0x1           /* insertion                    */
#define SOPER_DEL 0x2           /* deletion                     */
#define SOPER_SUB 0x3           /* substitution                 */

                                /* Levenhstein Opcodes masks    */
#define OPERMASK  0xc000000000000000    /* mask for Opcodes      /!\    */
#define NOPERMASK 0x3fffffffffffffff    /* negate of previous    /!\    */

                                /* special chars in pattern     */
#define PATCHARS  "[]!#"

                                /* 26 letter alphabet           */
                                /* in alphabetical order        */

#define ORD_ALPHA "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

                                /* protein alphabet             */

#define PROT_ALPHA "ACDEFGHIKLMNPQRSTVWY"

                                /* dna/rna alphabet             */

#define DNA_ALPHA "ABCDGHKMNRSTUVWXY"


/* ----------------------------------------------- */
/* data structures                                 */
/* ----------------------------------------------- */

typedef uint64_t patword_t;

                                        /* -------------------- */
typedef enum {                          /* data encoding        */
                                        /* -------------------- */
        alpha = 0,                      /* [A-Z]                */
        dna,                            /* IUPAC DNA            */
        protein                         /* IUPAC proteins       */
} CodType;

                                        /* -------------------- */
typedef struct {                        /* sequence             */
                                        /* -------------------- */
    char      *name;                    /* sequence name        */
    int32_t   seqlen;                   /* sequence length      */
    int32_t   seqsiz;                   /* sequence buffer size */
    int32_t   datsiz;                   /* data buffer size     */
    int32_t   circular;
    uint8_t   *data;                    /* data buffer          */
    char      *cseq;                    /* sequence buffer      */
    StackiPtr hitpos[MAX_PATTERN];      /* stack of hit pos.    */
    StackiPtr hiterr[MAX_PATTERN];      /* stack of errors      */
} Seq, *SeqPtr;

                                        /* -------------------- */
typedef struct {                        /* pattern              */
                                        /* -------------------- */
    int32_t    patlen;                  /* pattern length       */
    int32_t    maxerr;                  /* max # of errors      */
    char       *cpat;                   /* pattern string       */
    uint32_t   *patcode;                /* encoded pattern      */
    patword_t  *smat;                   /* S matrix             */
    patword_t  omask;                   /* oblig. bits mask     */
    bool       hasIndel;                /* are indels allowed   */
    bool       ok;                      /* is pattern ok        */
} Pattern, *PatternPtr;


/* ----------------------------------------------- */
/* prototypes                                      */
/* ----------------------------------------------- */

                                     /* apat_seq.c */

SeqPtr  FreeSequence     (SeqPtr pseq);
SeqPtr  NewSequence      (void);
int32_t ReadNextSequence (SeqPtr pseq);
int32_t WriteSequence    (FILE *filou , SeqPtr pseq);

                                   /* apat_parse.c      */
uint32_t *GetCode          (CodType ctype);
int32_t  CheckPattern      (Pattern *ppat);
int32_t  EncodePattern     (Pattern *ppat, CodType ctype);
int32_t  ReadPattern       (Pattern *ppat);
void     PrintDebugPattern (Pattern *ppat);
int      lenPattern        (const char *pat);

                                /* apat_search.c        */

int32_t CreateS           (Pattern *ppat, int32_t lalpha);
int32_t ManberNoErr       (Seq *pseq , Pattern *ppat, int32_t patnum,int32_t begin,int32_t length);
int32_t ManberSub         (Seq *pseq , Pattern *ppat, int32_t patnum,int32_t begin,int32_t length);
int32_t ManberIndel       (Seq *pseq , Pattern *ppat, int32_t patnum,int32_t begin,int32_t length);
int32_t ManberAll         (Seq *pseq , Pattern *ppat, int32_t patnum,int32_t begin,int32_t length);
int32_t NwsPatAlign       (Seq *pseq , Pattern *ppat, int32_t nerr, int32_t begin, int32_t *reslen, int32_t *reserr);                      /* apat_sys.c   */

float   UserCpuTime     (int32_t reset);
float   SysCpuTime      (int32_t reset);
char    *StrCpuTime     (int32_t reset);
void    Erreur          (char *msg , int32_t stat);
int32_t AccessFile      (char *path, char *mode);

#endif /* H_apat */