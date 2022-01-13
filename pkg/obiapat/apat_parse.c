/* ==================================================== */
/*      Copyright (c) Atelier de BioInformatique        */
/*      Mar. 92                                         */
/*      File: apat_parse.c                              */
/*      Purpose: Codage du pattern                      */
/*      History:                                        */
/*      00/07/94 : <Gloup> first version (stanford)     */
/*      00/11/94 : <Gloup> revised for DNA/PROTEIN      */
/*      30/12/94 : <Gloup> modified EncodePattern       */
/*                         for manber search            */
/*      14/05/99 : <Gloup> indels added                 */
/*      07/12/21 : <Zafacs> some cleaning for 2020      */
/* ==================================================== */

#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <ctype.h>

#include "apat.h"
                                /* -------------------- */
                                /* default char         */
                                /* encodings            */
                                /* -------------------- */

static uint32_t sDftCode[]  =  {

        0x00000001 /* A */, 0x00000002 /* B */, 0x00000004 /* C */,
        0x00000008 /* D */, 0x00000010 /* E */, 0x00000020 /* F */,
        0x00000040 /* G */, 0x00000080 /* H */, 0x00000100 /* I */,
        0x00000200 /* J */, 0x00000400 /* K */, 0x00000800 /* L */,
        0x00001000 /* M */, 0x00002000 /* N */, 0x00004000 /* O */,
        0x00008000 /* P */, 0x00010000 /* Q */, 0x00020000 /* R */,
        0x00040000 /* S */, 0x00080000 /* T */, 0x00100000 /* U */,
        0x00200000 /* V */, 0x00400000 /* W */, 0x00800000 /* X */,
        0x01000000 /* Y */, 0x02000000 /* Z */

};
                                /* -------------------- */
                                /* char encodings       */
                                /* IUPAC                */
                                /* -------------------- */

                                /* IUPAC Proteins       */
static uint32_t sProtCode[]  =  {

        0x00000001 /* A */, 0x00002008 /* B */, 0x00000004 /* C */,
        0x00000008 /* D */, 0x00000010 /* E */, 0x00000020 /* F */,
        0x00000040 /* G */, 0x00000080 /* H */, 0x00000100 /* I */,
        0x00000000 /* J */, 0x00000400 /* K */, 0x00000800 /* L */,
        0x00001000 /* M */, 0x00002000 /* N */, 0x00000000 /* O */,
        0x00008000 /* P */, 0x00010000 /* Q */, 0x00020000 /* R */,
        0x00040000 /* S */, 0x00080000 /* T */, 0x00000000 /* U */,
        0x00200000 /* V */, 0x00400000 /* W */, 0x037fffff /* X */,
        0x01000000 /* Y */, 0x00010010 /* Z */

};
                                /* IUPAC Dna/Rna        */
static uint32_t sDnaCode[]  =  {

        0x00000001 /* A */, 0x00080044 /* B */, 0x00000004 /* C */, 
        0x00080041 /* D */, 0x00000000 /* E */, 0x00000000 /* F */, 
        0x00000040 /* G */, 0x00080005 /* H */, 0x00000000 /* I */, 
        0x00000000 /* J */, 0x00080040 /* K */, 0x00000000 /* L */, 
        0x00000005 /* M */, 0x00080045 /* N */, 0x00000000 /* O */, 
        0x00000000 /* P */, 0x00000000 /* Q */, 0x00000041 /* R */, 
        0x00000044 /* S */, 0x00080000 /* T */, 0x00080000 /* U */, 
        0x00000045 /* V */, 0x00080001 /* W */, 0x00080045 /* X */, 
        0x00080004 /* Y */, 0x00000000 /* Z */

};


/* -------------------------------------------- */
/* internal replacement of gets                 */
/* -------------------------------------------- */
static char *sGets(char *buffer, int size) {
        
        char *ebuf;

        if (! fgets(buffer, size-1, stdin))
           return NULL;

        /* remove trailing line feed */

        ebuf = buffer + strlen(buffer); 

        while (--ebuf >= buffer) {
           if ((*ebuf == '\n') || (*ebuf == '\r'))
                *ebuf = '\000';
           else
                break;
        }

        return buffer;
}

/* -------------------------------------------- */
/* returns actual code associated to type       */
/* -------------------------------------------- */

uint32_t *GetCode(CodType ctype)
{
        uint32_t *code = sDftCode;

        switch (ctype) {
           case dna     : code = sDnaCode  ; break;
           case protein : code = sProtCode ; break;
           default      : code = sDftCode  ; break;
        }

        return code;
}

/* -------------------------------------------- */

#define BAD_IF(tst)   if (tst)  return 0

int CheckPattern(Pattern *ppat)
{
        int lev;
        char *pat;

        pat = ppat->cpat;
        
        BAD_IF (*pat == '#');

        for (lev = 0; *pat ; pat++)

            switch (*pat) {

                case '[' :
                   BAD_IF (lev);
                   BAD_IF (*(pat+1) == ']');
                   lev++;
                   break;

                case ']' :
                   lev--;
                   BAD_IF (lev);
                   break;

                case '!' :
                   BAD_IF (lev);
                   BAD_IF (! *(pat+1));
                   BAD_IF (*(pat+1) == ']');
                   break;

                case '#' :
                   BAD_IF (lev);
                   BAD_IF (*(pat-1) == '[');
                   break;

                default :
                   if (! isupper(*pat))
                        return 0;
                   break;
            }

        return (lev ? 0 : 1);
}
 
#undef BAD_IF
           

/* -------------------------------------------- */
static const char *skipOblig(const char *pat)
{
        return (*(pat+1) == '#' ? pat+1 : pat);
}

/* -------------------------------------------- */
static const char *splitPattern(const char *pat)
{
        switch (*pat) {
                   
                case '[' :
                   for (; *pat; pat++)
                        if (*pat == ']')
                          return skipOblig(pat);
                   return NULL;
                   break;

                case '!' :
                   return splitPattern(pat+1);
                   break;

        }
        
        return skipOblig(pat);                  
}

/* -------------------------------------------- */
static uint32_t valPattern(char *pat, uint32_t *code)
{
        uint32_t val;
        
        switch (*pat) {
                   
                case '[' :
                   return valPattern(pat+1, code);
                   break;

                case '!' :
                   val = valPattern(pat+1, code);
                   return (~val & PATMASK);
                   break;

                default :
                   val = 0x0;
                   while (isupper(*pat)) {
                       val |= code[*pat - 'A'];
                       pat++;
                   }
                   return val;
        }

        return 0x0;                     
}

/* -------------------------------------------- */
static uint32_t obliBitPattern(char *pat)
{
        return (*(pat + strlen(pat) - 1) == '#' ? OBLIBIT : 0x0);
}       
                

/* -------------------------------------------- */
int lenPattern(const char *pat)
{
        int  lpat;

        lpat = 0;
        
        while (*pat) {
        
            if (! (pat = splitPattern(pat)))
                return 0;

            pat++;

            lpat++;
        }

        return lpat;
}

/* -------------------------------------------- */
/* Interface                                    */
/* -------------------------------------------- */

/* -------------------------------------------- */
/* encode un pattern                            */
/* -------------------------------------------- */
int EncodePattern(Pattern *ppat, CodType ctype)
{
        int   pos, lpat;
        uint32_t *code;
        char  *pp, *pa, c;

        ppat->ok = false;

        code = GetCode(ctype);

        ppat->patlen = lpat = lenPattern(ppat->cpat);
        
        if (lpat <= 0)
            return 0;
        
        // if (! (ppat->patcode = NEWN(uint32_t, lpat)))
        //    return 0;

        pa = pp = ppat->cpat;

        pos = 0;
        
        while (*pa) {
        
            pp = (char*)splitPattern(pa);

            c = *++pp;
            
            *pp = '\000';
                    
            ppat->patcode[pos++] = valPattern(pa, code) | obliBitPattern(pa);
            
            *pp = c;
            
            pa = pp;
        }

        ppat->ok = true;

        return lpat;
}

/* -------------------------------------------- */
/* remove blanks                                */
/* -------------------------------------------- */
static char *RemBlanks(char *s)
{
        char *sb, *sc;

        for (sb = sc = s ; *sb ; sb++)
           if (! isspace(*sb))
                *sc++ = *sb;

        return s;
}

/* -------------------------------------------- */
/* count non blanks                             */
/* -------------------------------------------- */
static uint32_t CountAlpha(char *s)
{
        uint32_t n;

        for (n = 0 ; *s ; s++)
           if (! isspace(*s))
                n++;

        return n;
}
           
        
/* -------------------------------------------- */
/* lit un pattern                               */
/* <pattern> #mis                               */
/* ligne starting with '/' are comments         */
/* -------------------------------------------- */
int ReadPattern(Pattern *ppat)
{
        int  val;
        char *spac;
        char buffer[BUFSIZ];

        ppat->ok = true;

        if (! sGets(buffer, sizeof(buffer)))
            return 0;

        if (*buffer == '/')
            return ReadPattern(ppat);

        if (! CountAlpha(buffer))
            return ReadPattern(ppat);

        for (spac = buffer ; *spac ; spac++)
            if ((*spac == ' ') || (*spac == '\t'))
                break;

        ppat->ok = false;

        if (! *spac)
            return 0;
        
        if (sscanf(spac, "%d", &val) != 1)
            return 0;

        ppat->hasIndel = (val < 0);
        
        ppat->maxerr = ((val >= 0) ? val : -val);

        *spac = '\000';

        (void) RemBlanks(buffer);

        if ((ppat->cpat = NEWN(char, strlen(buffer)+1)))
            strcpy(ppat->cpat, buffer);
        
        ppat->ok = (ppat->cpat != NULL);

        return (ppat->ok ? 1 : 0);
}

/* -------------------------------------------- */
/* ecrit un pattern - Debug -                   */
/* -------------------------------------------- */
void PrintDebugPattern(Pattern *ppat)
{
        int i;

        printf("Pattern  : %s (length : %d)\n", ppat->cpat, ppat->patlen);
        printf("Encoding : \n\t");

        for (i = 0 ; i < ppat->patlen ; i++) {
            printf("0x%8.8x ", ppat->patcode[i]);
            if (i%4 == 3)
                printf("\n\t");
        }
        printf("\n");
}

