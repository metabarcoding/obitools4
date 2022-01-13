/* ==================================================== */
/*      Copyright (c) Atelier de BioInformatique        */
/*      Dec. 94                                         */
/*      File: apat_search.c                             */
/*      Purpose: recherche du pattern                   */
/*               algorithme de Baeza-Yates/Gonnet       */
/*                             Manber (agrep)           */
/*      History:                                        */
/*      07/12/94 : <MFS>   first version                */
/*      28/12/94 : <Gloup> revised version              */
/*      14/05/99 : <Gloup> last revision                */
/*      07/12/21 : <Zafacs> last some cleaning for 2020 */
/* ==================================================== */

#include <stdlib.h>
#include <string.h>

#include "libstki.h"
#include "apat.h"

#define POP             PopiOut
#define PUSH            PushiIn
#define TOPCURS         CursiToTop
#define DOWNREAD        ReadiDown

#define KRONECK(x, msk) ((~x & msk) ? 0 : 1)
#define MIN(x, y)       ((x) < (y)  ? (x) : (y))

/* -------------------------------------------- */
/* Construction de la matrice S                 */
/* -------------------------------------------- */

int CreateS(Pattern *ppat, int32_t lalpha)
{
        int32_t     indx, pindx, i, j; 
        patword_t   amask, omask, *smat;

        ppat->ok = false;

        omask = 0x0L;

        // if (! (smat = NEWN(uint32_t, lalpha)))
        //    return 0;
        smat = ppat->smat;

        for (i = 0 ; i < lalpha ; i++)
           smat[i] = 0x0;

        for (i = ppat->patlen - 1, amask = 0x1L ; i >= 0 ; i--, amask <<= 1) {

            indx = ppat->patcode[i];

            if (ppat->patcode[i] & OBLIBIT)
                omask |= amask;

            for (j = 0, pindx = 0x1L ; j < lalpha ; j++, pindx <<= 1)
                if (indx & pindx)
                    smat[j] |= amask;
        }

        ppat->smat = smat;

        ppat->omask = omask;

        ppat->ok = true;

        return 1;

}

/* -------------------------------------------- */
/* Baeza-Yates/Manber algorithm                 */
/* NoError                                      */
/* -------------------------------------------- */
int32_t ManberNoErr(Seq *pseq, Pattern *ppat, int patnum,int begin,int length)
{
        int32_t     pos;
        patword_t    smask, r;
        uint8_t     *data;
        StackiPtr *stkpos, *stkerr;
        int32_t    end;
        
        end = begin + length;
        end = (end <= (size_t)(pseq->seqlen+pseq->circular)) ? end:(size_t)(pseq->seqlen+pseq->circular);
        

                                        /* create local masks   */

        smask = r = 0x1L << ppat->patlen;

                                        /* init. scan           */
        data   = pseq->data + begin;
        stkpos = pseq->hitpos + patnum;
        EmptyStacki(stkpos[0]);
        stkerr = pseq->hiterr + patnum;
        EmptyStacki(stkerr[0]);

                                        /* loop on text data    */
                                        
        for (pos = begin ; pos < end ; pos++) {

            r = (r >> 1) & ppat->smat[*data++];

            if (r & 0x1L) {
                PUSH(stkpos, pos - ppat->patlen + 1);
                PUSH(stkerr, 0);
            }

            r |= smask;
        }
        
        return (*stkpos)->top;  /* aka # of hits        */
}

/* -------------------------------------------- */
/* Baeza-Yates/Manber algorithm                 */
/* Substitution only                            */
/*                                              */
/* Note : r array is stored as :                */
/*    0 0 r(0,j) r(0,j+1) r(1,j) r(1,j+1) ...   */
/*                                              */
/* -------------------------------------------- */
int32_t ManberSub(Seq *pseq, Pattern *ppat, int patnum,int begin,int length)
{
        int         e, emax, found;
        uint32_t    pos;
        uint32_t    smask, cmask, sindx;
        uint32_t    *pr, r[2 * MAX_PAT_ERR + 2];
        uint8_t     *data;
        StackiPtr   *stkpos, *stkerr;
        uint32_t    end;
        
        end = begin + length;
        end = (end <= (size_t)(pseq->seqlen+pseq->circular)) ? end:(size_t)(pseq->seqlen+pseq->circular);

                                        /* create local masks   */
        emax = ppat->maxerr;

        r[0] = r[1] = 0x0;

        cmask = smask = 0x1L << ppat->patlen;

        for (e = 0, pr = r + 3 ; e <= emax ; e++, pr += 2)
                *pr = cmask;

        cmask = ~ ppat->omask;

                                        /* init. scan           */
        data   = pseq->data + begin;
        stkpos = pseq->hitpos + patnum;
        EmptyStacki(stkpos[0]);
        stkerr = pseq->hiterr + patnum;
        EmptyStacki(stkerr[0]);

                                        /* loop on text data    */
                                        
        for (pos = begin ; pos < end ; pos++) {

            sindx  = ppat->smat[*data++];

            for (e = found = 0, pr = r ; e <= emax ; e++, pr += 2) {
                
                pr[2]  = pr[3] | smask;

                pr[3]  =   ((pr[0] >> 1) & cmask)       /* sub   */
                         | ((pr[2] >> 1) & sindx);      /* ident */

                if (pr[3] & 0x1L) {                     /* found */
                    if (! found)  {
                       PUSH(stkpos, pos - ppat->patlen + 1);
                       PUSH(stkerr, e);
                    }
                    found++;
                }
            }
        }

        return (*stkpos)->top;  /* aka # of hits        */
}

/* -------------------------------------------- */
/* Baeza-Yates/Manber algorithm                 */
/* Substitution + Indels                        */
/*                                              */
/* Note : r array is stored as :                */
/*    0 0 r(0,j) r(0,j+1) r(1,j) r(1,j+1) ...   */
/*                                              */
/* Warning: may return shifted pos.             */
/*                                              */
/* -------------------------------------------- */
int32_t ManberIndel(Seq *pseq, Pattern *ppat, int patnum,int begin,int length)
{
        int       e, emax, found;
        uint32_t     pos;
        uint32_t    smask, cmask, sindx;
        uint32_t    *pr, r[2 * MAX_PAT_ERR + 2];
        uint8_t     *data;
        StackiPtr *stkpos, *stkerr;
        uint32_t    end;
        
        end = begin + length;
        end = (end <= (size_t)(pseq->seqlen+pseq->circular)) ? end:(size_t)(pseq->seqlen+pseq->circular);

                                        /* create local masks   */
        emax = ppat->maxerr;

        r[0] = r[1] = 0x0;

        cmask = smask = 0x1L << ppat->patlen;

        for (e = 0, pr = r + 3 ; e <= emax ; e++, pr += 2) {
                *pr = cmask;
                cmask = (cmask >> 1) | smask;
        }

        cmask = ~ ppat->omask;

                                        /* init. scan           */
        data   = pseq->data + begin;
        stkpos = pseq->hitpos + patnum;
        stkerr = pseq->hiterr + patnum;

                                        /* loop on text data    */
                                        
        for (pos = begin ; pos < end ; pos++) {

            sindx  = ppat->smat[*data++];

            for (e = found = 0, pr = r ; e <= emax ; e++, pr += 2) {
                
                pr[2]  = pr[3] | smask;

                pr[3]   =  ((     pr[0]                 /* ins   */
                              |  (pr[0] >> 1)           /* sub   */
                              |  (pr[1] >> 1))          /* del   */
                            & cmask)
                          | ((pr[2] >> 1) & sindx);     /* ident */

                if (pr[3] & 0x1L) {                     /* found */
                    if (! found) {
                        PUSH(stkpos, pos - ppat->patlen + 1);
                        PUSH(stkerr, e);
                    }
                    found++;
                }

            }
        }

        return (*stkpos)->top;  /* aka # of hits        */
}

/* -------------------------------------------- */
/* Baeza-Yates/Manber algorithm                 */
/* API call to previous functions               */
/* -------------------------------------------- */
int32_t ManberAll(Seq *pseq, Pattern *ppat, int patnum,int begin,int length)
{
        if (ppat->maxerr == 0)
           return ManberNoErr(pseq, ppat, patnum, begin, length);
        else if (ppat->hasIndel)
           return ManberIndel(pseq, ppat, patnum, begin, length);
        else
           return ManberSub(pseq, ppat, patnum, begin, length);
}


/* -------------------------------------------- */
/* Alignement NWS                               */
/* pour edition des hits                        */
/* (avec substitution obligatoire aux bords)    */
/* -------------------------------------------- */

int32_t NwsPatAlign(pseq, ppat, nerr, reslen, reserr)
        Seq     *pseq;
        Pattern *ppat;
        int32_t   nerr, *reslen, *reserr;
{
        uint8_t  *sseq, *px;
        int32_t  i, j, lseq, lpat, npos, dindel, dsub,
               *pc, *pi, *pd, *ps;
        uint32_t amask;

        static int32_t sTab[(MAX_PAT_LEN+MAX_PAT_ERR+1) * (MAX_PAT_LEN+1)];

        lseq = pseq->seqlen;

        pc = sTab;              /*  |----|----| --> i   */
        pi = pc - 1;            /*  | ps | pd | |       */
        pd = pi - lseq;         /*  |----|----| |       */
        ps = pd - 1;            /*  | pi | pc | v j     */
                                /*  |---------|         */

        lseq = pseq->seqlen;
        lpat = ppat->patlen;

        sseq = pseq->data - 1;

        amask = ONEMASK >> lpat;

        for (j = 0 ; j <= lpat ; j++) {

           for (i = 0 , px = sseq ; i <= lseq ; i++, px++) {

              if (i && j) {
                  dindel = MIN(*pi, *pd) + 1;
                  dsub   = *ps + KRONECK(ppat->smat[*px], amask);
                  *pc    = MIN(dindel, dsub);
              }
              else if (i)               /* j == 0       */
                  *pc = *pi + 1;
              else if (j)               /* i == 0       */
                  *pc = *pd + 1;
              else                      /* root         */
                   *pc = 0;

              pc++;
              pi++;
              pd++;
              ps++;
           }
        
           amask <<= 1;
        }

        pc--;

        for (i = lseq, npos = 0 ; i >= 0 ; i--, pc--) {
            if (*pc <= nerr) {
                *reslen++ =  i;
                *reserr++ = *pc;
                npos++;
             }
        }

        return npos;
}
