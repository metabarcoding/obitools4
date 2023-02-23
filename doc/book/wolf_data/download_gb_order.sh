#!/bin/bash

#
# Used resources URLs
#

NCBIURL="https://ftp.ncbi.nlm.nih.gov/"            # NCBI Web site URL
GBURL="${NCBIURL}genbank/"                         # Directory of Genbank flat files 
TAXOURL="${NCBIURL}pub/taxonomy/taxdump.tar.gz"    # NCBI Taxdump

LOGFILE="download.log"

#
# List of downloaded Genbank divisions
#

DIV="bct|inv|mam|phg|pln|pri|rod|vrl|vrt"

############################
#
#  Functions
#
############################

pattern_at_rank() {
   local taxo="$1"
   local rank="$2"

   echo "^($(awk -F "|" -v rank="$rank" 'BEGIN {
                          ORS="|";
                          rank="\t" rank "\t"
                        } 
            ($3 ~ rank) {sub(/^[ \t]+/,"",$1); 
                         sub(/[ \t]+$/,"",$1); 
                         print $1}
           ' "${taxo}/nodes.dmp" \
       | sed 's/|$//'))$"
}


#
# Extrate from the web site the current Genbank release number
# end create the corresponding directory
#

echo "Looking at current Genbank release number"
GB_Release_Number=$(curl "${GBURL}GB_Release_Number")
echo "identified release number is : ${GB_Release_Number}"

mkdir -p "Release-${GB_Release_Number}"
cd "Release-${GB_Release_Number}" || exit

#
# Download the current NCBI taxonomy
#
mkdir -p "ncbitaxo"

if [[ ! -f ncbitaxo/nodes.dmp ]] || [[ ! -f ncbitaxo/names.dmp ]] ; then
   curl "${TAXOURL}" \
      | tar -C "ncbitaxo" -zxf -
fi

curl $GBURL > index.html

for f in $(grep -E "gb(${DIV})[0-9]+\.seq\.gz" index.html \
            | sed -E 's@^.*<a href="([^"]+)">.*</a>.*$@\1@' ) ; do
   fasta=${f/seq.gz/fasta}
   stamp=${f/seq.gz/stamp}

   echo "File : $f saved into $fasta"

   rm -f ${fasta}.downloading

   while [[ ! -f "stamp/${stamp}" ]] ; do
      status=""
      wget "${GBURL}${f}" && \
      status=$( ( (gzip -dc "$f" 2>> "$LOGFILE"  || echo "Unzipping error" 1>&2) \
                  | (obiannotate --genbank -t ncbitaxo \
                                 --with-taxon-at-rank kingdom \
                                 --with-taxon-at-rank superkingdom \
                                 --with-taxon-at-rank phylum\
                                 --with-taxon-at-rank order  \
                                 --with-taxon-at-rank family  \
                                 --with-taxon-at-rank genus  \
                                 -S division='"misc-@-0"' \
                                 -S section='"misc-@-0"' \
                                 2>> "$LOGFILE" || echo "Fasta conversion error" 1>&2) \
                  | (obigrep -A genus_taxid -A family_taxid 2>> "$LOGFILE" \
                     | obigrep -p 'annotations.genus_taxid > 0 && annotations.family_taxid > 0' \
                               -p 'annotations.phylum_taxid > 0 || annotations.order_taxid > 0' \
                                 2>> "$LOGFILE" || echo "Fasta filtering error" 1>&2) \
                  | (obiannotate -p 'annotations.superkingdom_taxid > 0' \
                                 -S division='printf("%s-S-%d",subspc(annotations.superkingdom_name),annotations.superkingdom_taxid)' \
                                 2>> "$LOGFILE" || echo "Fasta annotation error" 1>&2) \
                  | (obiannotate -p 'annotations.kingdom_taxid > 0' \
                                 -S division='printf("%s-K-%d",subspc(annotations.kingdom_name),annotations.kingdom_taxid)' \
                                 2>> "$LOGFILE" || echo "Fasta annotation error" 1>&2) \
                  | (obiannotate -p 'annotations.phylum_taxid > 0' \
                                 -S section='printf("%s-P-%d",subspc(annotations.phylum_name),annotations.phylum_taxid)' \
                                 2>> "$LOGFILE" || echo "Fasta annotation error" 1>&2) \
                  | (obiannotate -p 'annotations.order_taxid > 0' \
                                 -S section='printf("%s-O-%d",subspc(annotations.order_name),annotations.order_taxid)' \
                                 2>> "$LOGFILE" || echo "Fasta annotation error" 1>&2) > "${fasta}.downloading") 2>&1 )
      echo
      rm -f "${f}"

      if [[ -z "$status" ]] ; then
         echo "Downloading of $f succeded ($(obicount -v "${fasta}.downloading" 2>/dev/null) sequences)"
         mv "${fasta}.downloading" "${fasta}"
         mkdir -p stamp
         touch "stamp/${stamp}"
      else
         echo "Downloading of $f failed"
         echo "$status"
         rm -f "${fasta}"
         rm -f "${fasta}.downloading"
      fi
   done

   if [[ -f "$fasta" ]] ; then 
      obidistribute -Z -A -p "%s.fasta" -c section -d division "$fasta"
      rm -f "$fasta"
   fi

done

