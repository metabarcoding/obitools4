#!/bin/bash

URL=https://ftp.ncbi.nlm.nih.gov/genbank/
DIV="bct|inv|mam|phg|pln|pri|rod|vrl|vrt"

GB_Release_Number=$(curl "${URL}GB_Release_Number")

mkdir -p "Release-${GB_Release_Number}"
cd "Release-${GB_Release_Number}"

curl $URL > index.html

for f in $(egrep "gb(${DIV})[0-9]+\.seq\.gz" index.html \
            | sed -E 's@^.*<a href="([^"]+)">.*</a>.*$@\1@' ) ; do
   echo -n "File : $f"

   if [[ -f $f ]] ; then
      gzip -t $f && echo " ok" || rm -f $f
   fi

   while [[ ! -f $f ]] ; do
        echo downloading
        wget2 --progress bar -v -o - $URL$f 
           if [[ -f $f ]] ; then
              gzip -t $f && echo " ok" || rm -f $f
           fi
   done
done

