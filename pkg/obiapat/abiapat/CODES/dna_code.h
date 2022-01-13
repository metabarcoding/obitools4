/* ----------------------------------------------- */
/* dna_code.h                                      */
/* alphabet encoding for dna/rna                   */
/* -----------------------------------------       */
/* IUPAC encoding                                  */
/* -----------------------------------------       */
/* G/A/T/C                                         */
/* U=T                                             */
/* R=AG                                            */
/* Y=CT                                            */
/* M=AC                                            */
/* K=GT                                            */
/* S=CG                                            */
/* W=AT                                            */
/* H=ACT                                           */
/* B=CGT                                           */
/* V=ACG                                           */
/* D=AGT                                           */
/* N=ACGT                                          */
/* X=ACGT                                          */
/* EFIJLOPQZ  not recognized                       */
/* -----------------------------------------       */
/* dual encoding                                   */
/* -----------------------------------------       */
/* A=ADHMNRVW                                      */
/* B=BCDGHKMNRSTUVWY                               */
/* C=BCHMNSVY                                      */
/* D=ABDGHKMNRSTUVWY                               */
/* G=BDGKNRSV                                      */
/* H=ABCDHKMNRSTUVWY                               */
/* K=BDGHKNRSTUVWY                                 */
/* M=ABCDHMNRSVWY                                  */
/* N=ABCDGHKMNRSTUVWY                              */
/* R=ABDGHKMNRSVW                                  */
/* S=BCDGHKMNRSVY                                  */
/* T=BDHKNTUWY                                     */
/* U=BDHKNTUWY                                     */
/* V=ABCDGHKMNRSVWY                                */
/* W=ABDHKMNRTUVWY                                 */
/* X=ABCDGHKMNRSTUVWY                              */
/* Y=BCDHKMNSTUVWY                                 */
/* EFIJLOPQZ  not recognized                       */
/* ----------------------------------------------- */

#ifndef USE_DUAL

        /* IUPAC */

        0x00000001 /* A */, 0x00080044 /* B */, 0x00000004 /* C */, 
        0x00080041 /* D */, 0x00000000 /* E */, 0x00000000 /* F */, 
        0x00000040 /* G */, 0x00080005 /* H */, 0x00000000 /* I */, 
        0x00000000 /* J */, 0x00080040 /* K */, 0x00000000 /* L */, 
        0x00000005 /* M */, 0x00080045 /* N */, 0x00000000 /* O */, 
        0x00000000 /* P */, 0x00000000 /* Q */, 0x00000041 /* R */, 
        0x00000044 /* S */, 0x00080000 /* T */, 0x00080000 /* U */, 
        0x00000045 /* V */, 0x00080001 /* W */, 0x00080045 /* X */, 
        0x00080004 /* Y */, 0x00000000 /* Z */

#else
        /* DUAL  */

        0x00623089 /* A */, 0x017e34ce /* B */, 0x01243086 /* C */, 
        0x017e34cb /* D */, 0x00000000 /* E */, 0x00000000 /* F */, 
        0x0026244a /* G */, 0x017e348f /* H */, 0x00000000 /* I */, 
        0x00000000 /* J */, 0x017e24ca /* K */, 0x00000000 /* L */, 
        0x0166308f /* M */, 0x017e34cf /* N */, 0x00000000 /* O */, 
        0x00000000 /* P */, 0x00000000 /* Q */, 0x006634cb /* R */, 
        0x012634ce /* S */, 0x0158248a /* T */, 0x0158248a /* U */, 
        0x016634cf /* V */, 0x017a348b /* W */, 0x017e34cf /* X */, 
        0x017c348e /* Y */, 0x00000000 /* Z */
#endif
