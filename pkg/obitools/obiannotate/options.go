package obiannotate

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obigrep"
	"github.com/DavidGamba/go-getoptions"
)

var _addRank = false
var _toBeRenamed = make(map[string]string, 0)
var _toBeDeleted = make([]string, 0)
var _keepOnly = make([]string, 0)
var _taxonAtRank = make([]string, 0)
var _evalAttribute = make(map[string]string, 0)
var _tagList = ""
var _clearAll = false
var _setSeqLength = false
var _uniqueID = false
var _ahoCorazick = ""
var _lcaSlot = ""
var _lcaError = 0.0

func SequenceAnnotationOptionSet(options *getoptions.GetOpt) {
	// options.BoolVar(&_addRank, "seq-rank", _addRank,
	// 	options.Description("Adds a new attribute named seq_rank to the sequence record indicating its entry number in the sequence file."),
	// )

	options.BoolVar(&_clearAll, "clear", _clearAll,
		options.Description("Clears all attributes associated to the sequence records."),
	)

	options.BoolVar(&_setSeqLength, "length", _setSeqLength,
		options.Description("Adds attribute with seq_length as a key and sequence length as a value."),
	)

	options.StringVar(&_ahoCorazick, "aho-corasick", _ahoCorazick,
		options.Description("Adds an aho-corasick attribut with the count of matches of the provided patterns."))

	options.StringVar(&_lcaSlot, "add-lca", _lcaSlot,
		options.ArgName("SLOT_NAME"),
		options.Description("From the taxonomic annotation of the sequence (taxid slot or merged_taxid slot), "+
			"a new slot named <SLOT_NAME> is added with the taxid of the lowest common ancester corresponding "+
			"to the current annotation."))

	options.Float64Var(&_lcaError, "lca-error", _lcaError,
		options.ArgName("#.###"),
		options.Description("Error rate tolerated on the taxonomical discription during the lowest common "+
			"ancestor. At most a fraction of lca-error of the taxonomic information can disagree with the "+
			"estimated LCA."),
	)

	// options.BoolVar(&_uniqueID, "uniq-id", _uniqueID,
	// 	options.Description("Forces sequence record ids to be unique."),
	// )

	options.StringMapVar(&_evalAttribute, "set-tag", 1, 1,
		options.Alias("S"),
		options.ArgName("KEY=EXPRESSION"),
		options.Description("Creates a new attribute named with a key <KEY> "+
			"sets with a value computed from <EXPRESSION>."))

	options.StringMapVar(&_toBeRenamed, "rename-tag", 1, 1,
		options.Alias("R"),
		options.ArgName("NEW_NAME=OLD_NAME"),
		options.Description("Changes attribute name <OLD_NAME> to <NEW_NAME>. When attribute named <OLD_NAME>"+
			" is missing, the sequence record is skipped and the next one is examined."))

	options.StringSliceVar(&_toBeDeleted, "delete-tag", 1, 1,
		options.ArgName("KEY"),
		options.Description(" Deletes attribute named <KEY>.When this attribute is missing,"+
			" the sequence record is skipped and the next one is examined."))

	options.StringSliceVar(&_taxonAtRank, "with-taxon-at-rank", 1, 1,
		options.ArgName("RANK_NAME"),
		options.Description("Adds taxonomic annotation at taxonomic rank <RANK_NAME>."))

	// options.StringVar(&_tagList, "tag-list", _tagList,
	// 	options.ArgName("FILENAME"),
	// 	options.Description("<FILENAME> points to a file containing attribute names"+
	// 		" and values to modify for specified sequence records."))

	options.StringSliceVar(&_keepOnly, "keep", 1, 1,
		options.Alias("k"),
		options.ArgName("KEY"),
		options.Description("Keeps only attribute with key <KEY>. Several -k options can be combined."))

}

// OptionSet adds to the basic option set every options declared for
// the obipcr command
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	obigrep.SequenceSelectionOptionSet(options)
	SequenceAnnotationOptionSet(options)
}

// -S <KEY>:<PYTHON_EXPRESSION>, --set-tag=<KEY>:<PYTHON_EXPRESSION>
// Creates a new attribute named with a key <KEY> and a value computed from <PYTHON_EXPRESSION>.

// --set-identifier=<PYTHON_EXPRESSION>
// Sets sequence record identifier with a value computed from <PYTHON_EXPRESSION>.

// --run=<PYTHON_EXPRESSION>
// Runs a python expression on each selected sequence.

// --set-sequence=<PYTHON_EXPRESSION>
// Changes the sequence itself with a value computed from <PYTHON_EXPRESSION>.

// -T, --set-definition=<PYTHON_EXPRESSION>
// Sets sequence definition with a value computed from <PYTHON_EXPRESSION>.

// -O, --only-valid-python
// Allows only valid python expressions.

// -m <MCLFILE>, --mcl=<MCLFILE>
// Creates a new attribute containing the number of the cluster the sequence record was assigned to, as indicated in file <MCLFILE>.

// --uniq-id
// Forces sequence record ids to be unique.

func CLIHasAttributeToBeRenamed() bool {
	return len(_toBeRenamed) > 0
}

func CLIAttributeToBeRenamed() map[string]string {
	return _toBeRenamed
}

func CLIHasAttibuteToDelete() bool {
	return len(_toBeDeleted) > 0
}

func CLIAttibuteToDelete() []string {
	return _toBeDeleted
}

func CLIHasToBeKeptAttributes() bool {
	return len(_keepOnly) > 0
}

func CLIToBeKeptAttributes() []string {
	return _keepOnly
}

func CLIHasTaxonAtRank() bool {
	return len(_taxonAtRank) > 0
}

func CLITaxonAtRank() []string {
	return _taxonAtRank
}

func CLIHasSetLengthFlag() bool {
	return _setSeqLength
}

func CLIHasClearAllFlag() bool {
	return _clearAll
}

func CLIHasSetAttributeExpression() bool {
	return len(_evalAttribute) > 0
}

func CLISetAttributeExpression() map[string]string {
	return _evalAttribute
}

func CLIHasAhoCorasick() bool {
	_, err := os.Stat(_ahoCorazick)
	return err == nil
}

func CLIAhoCorazick() []string {
	content, err := ioutil.ReadFile(_ahoCorazick)
	if err != nil {
		log.Fatalln("Cannot open file ", _ahoCorazick)
	}
	lines := strings.Split(string(content), "\n")

	j := 0
	for _, s := range lines {
		if len(s) > 0 {
			lines[j] = strings.ToLower(s)
			j++
		}
	}

	lines = lines[0:j]

	return lines
}

func CLILCASlotName() string {
	return _lcaSlot
}

func CLIHasAddLCA() bool {
	return _lcaSlot != ""
}

func CLILCAThreshold() float64 {
	return 1 - _lcaError
}
