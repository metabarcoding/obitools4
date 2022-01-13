

#ifndef __APAT_MEM_H__
#define __APAT_MEM_H__

/* ----------------------------------------------- */
/* macros                                          */
/* ----------------------------------------------- */

#define NEW(typ)                (typ*)malloc(sizeof(typ)) 
#define NEWN(typ, dim)          (typ*)malloc((uint64_t)(dim) * sizeof(typ))
#define REALLOC(typ, ptr, dim)  (typ*)realloc((void *) (ptr), (uint64_t)(dim) * sizeof(typ))
#define FREE(ptr)               free((void *) ptr)

#endif /* __APAT_MEM_H__ */