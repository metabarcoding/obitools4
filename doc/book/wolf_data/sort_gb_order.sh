#!/bin/bash
#!/bin/bash
#OAR -n gbsort
##OAR --array 50
##OAR --array-param-file 50_first.tsv
#OAR --project phyloalps
#OAR -l nodes=1/core=10,walltime=24:00:00
#OAR -O gbsort.%jobid%.log
#OAR -E gbsort.%jobid%.log


# /silenus/PROJECTS/pr-phyloalps/coissac
# /bettik/LECA/ENVIRONMENT/data/biodatabase/genbank

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


GBDIR=$1

#
# Extrate from the web site the current Genbank release number
# end create the corresponding directory
#

echo "Looking at current Genbank release number"
GB_Release_Number=$(for r in $(ls -d "${GBDIR}/Release-"* ); do 
                        basename $r; 
                    done \
                    | sort -r \
                    | head -1 \
                    | sed 's/^Release-//')

GB_Release_Number=251

echo "identified latest release number is : ${GB_Release_Number}"

GBSOURCE="${GBDIR}/Release-${GB_Release_Number}"

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


for f in $(ls -1 "${GBSOURCE}/"*.seq.gz ) ; do

   echo "PROCESSING : $f saved into $fasta" $(pwd)

      obiannotate --genbank -t ncbitaxo \
                                 --with-taxon-at-rank kingdom \
                                 --with-taxon-at-rank superkingdom \
                                 --with-taxon-at-rank phylum\
                                 --with-taxon-at-rank order  \
                                 --with-taxon-at-rank family  \
                                 --with-taxon-at-rank genus  \
                                 -S division='"misc-@-0"' \
                                 -S section='"misc-@-0"' \
                                 "$f" \
      | obigrep -A genus_taxid -A family_taxid \
      | obigrep -p 'annotations.genus_taxid > 0 && annotations.family_taxid > 0' \
                -p 'annotations.phylum_taxid > 0 || annotations.order_taxid > 0' \
      | obiannotate -p 'annotations.superkingdom_taxid > 0' \
                -S division='printf("%s-S-%d",subspc(annotations.superkingdom_name),annotations.superkingdom_taxid)' \
      | obiannotate -p 'annotations.kingdom_taxid > 0' \
                -S division='printf("%s-K-%d",subspc(annotations.kingdom_name),annotations.kingdom_taxid)' \
      | obiannotate -p 'annotations.phylum_taxid > 0' \
                -S section='printf("%s-P-%d",subspc(annotations.phylum_name),annotations.phylum_taxid)' \
      | obiannotate -p 'annotations.order_taxid > 0' \
                -S section='printf("%s-O-%d",subspc(annotations.order_name),annotations.order_taxid)' \
      | obidistribute -Z -A -p "%s.fasta" -c section -d division 



done

