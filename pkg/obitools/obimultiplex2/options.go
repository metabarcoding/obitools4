package obimultiplex2

import (
	"fmt"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obingslibrary"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"

	log "github.com/sirupsen/logrus"
)

var _NGSFilterFile = ""
var _askTemplate = false
var _UnidentifiedFile = ""
var _AllowedMismatch = -1
var _AllowsIndel = false
var _ConservedError = false

// PCROptionSet defines every options related to a simulated PCR.
//
// The function adds to a CLI every options proposed to the user
// to tune the parametters of the PCR simulation algorithm.
//
// # Parameters
//
// - option : is a pointer to a getoptions.GetOpt instance normaly
// produced by the
func MultiplexOptionSet(options *getoptions.GetOpt) {
	options.StringVar(&_NGSFilterFile, "tag-list", _NGSFilterFile,
		options.Alias("t"),
		options.Description("File name of the NGSFilter file describing PCRs."))

	options.BoolVar(&_ConservedError, "keep-errors", _ConservedError,
		options.Description("Prints symbol counts."))

	options.BoolVar(&_AllowsIndel, "with-indels", _AllowsIndel,
		options.Description("Allows for indels during the primers matching."))

	options.StringVar(&_UnidentifiedFile, "unidentified", _UnidentifiedFile,
		options.Alias("u"),
		options.Description("Filename used to store the sequences unassigned to any sample."))

	options.IntVar(&_AllowedMismatch, "allowed-mismatches", _AllowedMismatch,
		options.Alias("e"),
		options.Description("Used to specify the number of errors allowed for matching primers."))

	options.BoolVar(&_askTemplate, "template", _askTemplate,
		options.Description("Print on the standard output an example of CSV configuration file."),
	)

}

func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	MultiplexOptionSet(options)
}

func CLIAllowedMismatch() int {
	return _AllowedMismatch
}

func CLIAllowsIndel() bool {
	return _AllowsIndel
}
func CLIUnidentifiedFileName() string {
	return _UnidentifiedFile
}

func CLIConservedErrors() bool {
	return _UnidentifiedFile != "" || _ConservedError
}

func CLIHasNGSFilterFile() bool {
	return _NGSFilterFile != ""
}

func CLINGSFIlter() (*obingslibrary.NGSLibrary, error) {
	file, err := os.Open(_NGSFilterFile)

	if err != nil {
		return nil, fmt.Errorf("open file error: %v", err)
	}

	log.Infof("Reading NGSFilter file: %s", _NGSFilterFile)
	ngsfiler, err := obiformats.ReadNGSFilter(file)

	if err != nil {
		return nil, fmt.Errorf("NGSfilter reading file error: %v", err)
	}

	return ngsfiler, nil
}

func CLIAskConfigTemplate() bool {
	return _askTemplate
}

func CLIConfigTemplate() string {
	return `###
### Example of NGSFilter CSV configuration file
###
#
# The CSV file can contain comments starting with the # character
# and empty lines.
# At the top of the file a set of lines of three or four columns and having
# the first column containing @param can be used to define parameters
# for the obimultiplex tool. The structure of these lines is :
#
#       @param,parameter_name,parameter_value
#       @param,parameter_name,parameter_value1,parameter_value2
#
# The following lines describes the PCR multiplexed in the sequencing library.
# The first line describes the columns of the CSV file and the following lines
# describe the PCR multiplexed.
#
# Five columns are expected :
#
# - experiment: the experiment name
# - sample: the sample (pcr) name
# - sample_tag: the tag identifying the sample
# - forward_primer: the forward primer sequence
# - reverse_primer: the reverse primer sequence
#
# Supplementary columns are allowed. Their names and content will be used to
# annotate the sequence corresponding to the sample, as the key=value; located
# after the @ sign did in the original ngsfilter file format.
#
###
###  Description of the parameters
###
#
# The forward_spacer and the reverse_spacer allow to specify the number of
# nucleotide separating the 5' end of the forward or reverse primer respectively
# to the 3' end of the tag. The default value is 0.
#
# The param spacer allows for specify this value for both forward and reverse 
# simultaneously. The spacer parameter can also, when used wirh two arguments,
# allow to specify the # the spacer value for a specific primer:
#
#       @param,spacer,CAGCTGCTATGTCGATGCTGACT,2
#
@param,forward_spacer,0
@param,reverse_spacer,0
#
# A new method for designing indel proof tag is to not use one of the four 
# nucleotides in their sequence and to flank the tag with this fourth nucleotide.
# That nucleotide is the tag delimiter. Similarly, to the spacer value, 
# three ways to specify the tag delimiter exist:
#   - the forward_tag_delimiter and reverse_tag_delimiter
#   - the tag_delimiter in its two forms with one and two arguments
#
@param,forward_tag_delimiter,0
@param,reverse_tag_delimiter,0
#
# Three algorithms are available to math a pair of tags with a sample.
# It is specified using the @matching parameter. The three possible
# values are strict, hamming, and indel. The default value is strict.
# As for previous parameters, forward_matching and reverse_matching can
# be used to specify the matching value for each primer. And spacer
# can be used with two arguments to specify the matching value for 
# a specific primer.
#
@param,matching,strict
#
# The primer_mismatches parameter allows to specify the number of errors allowed
# when matching the primer. The default value is 2. The same declination of
# the parameters forward_primer_mismatches and reverse_primer_mismatches exist.
#
@param,primer_mismatches,2
#
# The @indel parameter allows to specify if indel are allowed during the matching
# of the primers to the sequence. The default value is false. forward_indel and
# reverse_indel can be used to specify the value for each primer.
#
@param,indels,false
#
###
###  Description of the PCR multiplexed
###
#
# Below is an example for the minimal description of the PCRs multiplexed in the
# sequencing library.
#
# The first line is the column names and must exist.
# Five columns are expected :
# - experiment: the experiment name, that allows for grouping samples
# - sample: the sample (pcr) name
# - sample_tag: the tag identifying the sample
#   The sample tag must be unique in the library for a given pair of primers
#   + They can be a simple DNA word as here. This means that the same tag is used
#     for both primers.
#   + It can be two DNA words separated by a colon. For example, aagtag:gaagtag.
#     This means that the first tag is used for the forward primer and the second for the
#     reverse primers. "aagtag" is the same as "aagtag:aagtag".
#   + In the two word syntax, if a primer forward or reverse is not tagged, its tag
#     is replaced by a hyphen '-', for example 'aagtag:-' or '-:aagtag'.
#   For a given primer all the tags must have the same length.
# - forward_primer: the forward primer sequence
# - reverse_primer: the reverse primer sequence
# 
experiment,sample,sample_tag,forward_primer,reverse_primer
wolf_diet,13a_F730603,aattaac,TTAGATACCCCACTATGC,TAGAACAGGCTCCTCTAG
wolf_diet,15a_F730814,gaagtag,TTAGATACCCCACTATGC,TAGAACAGGCTCCTCTAG
wolf_diet,26a_F040644,gaatatc,TTAGATACCCCACTATGC,TAGAACAGGCTCCTCTAG
wolf_diet,29a_F260619,gcctcct,TTAGATACCCCACTATGC,TAGAACAGGCTCCTCTAG
`
}
