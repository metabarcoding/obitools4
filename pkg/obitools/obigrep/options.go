package obigrep

import (
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/DavidGamba/go-getoptions"
)

var _BelongTaxa = make([]string, 0)
var _NotBelongTaxa = make([]string, 0)
var _RequiredRanks = make([]string, 0)
var _ValidateTaxonomy = false

var _MinimumLength = 1
var _MaximumLength = int(2e9)

var _MinimumCount = 1
var _MaximumCount = int(2e9)

var _SequencePatterns = make([]string, 0)
var _DefinitionPatterns = make([]string, 0)
var _IdPatterns = make([]string, 0)

var _Predicats = make([]string, 0)

var _IdList = ""

var _Taxdump = ""

var _RequiredAttributes = make([]string, 0)
var _AttributePatterns = make(map[string]string, 0)

var _InvertMatch = false
var _SaveRejected = ""

var _PairedMode = "forward"

var _approx_pattern = make([]string, 0)
var _pattern_error = 0
var _pattern_indel = false
var _pattern_only_forward = false

func TaxonomySelectionOptionSet(options *getoptions.GetOpt) {

	options.StringSliceVar(&_BelongTaxa, "restrict-to-taxon", 1, 1,
		options.Alias("r"),
		options.ArgName("TAXID"),
		options.Description("Require that the actual taxon of the sequence belongs the provided taxid."))

	options.StringSliceVar(&_NotBelongTaxa, "ignore-taxon", 1, 1,
		options.Alias("i"),
		options.ArgName("TAXID"),
		options.Description("Require that the actual taxon of the sequence doesn't belong the provided taxid."))

	options.StringSliceVar(&_RequiredRanks, "require-rank", 1, 1,
		options.ArgName("RANK_NAME"),
		options.Description("Select sequences belonging a taxon with a rank <RANK_NAME>"))

	options.BoolVar(&_ValidateTaxonomy, "valid-taxid", _ValidateTaxonomy,
		options.Description("Validate the taxonomic classification of the sequences."))

}

func SequenceSelectionOptionSet(options *getoptions.GetOpt) {
	TaxonomySelectionOptionSet(options)

	options.StringVar(&_IdList, "id-list", _IdList,
		options.ArgName("FILENAME"),
		options.Description("<FILENAME> points to a text file containing the list of sequence record identifiers to be selected."+
			" The file format consists in a single identifier per line."))

	options.BoolVar(&_InvertMatch, "inverse-match", _InvertMatch,
		options.Alias("v"),
		options.Description("Inverts the sequence record selection."))

	options.IntVar(&_MinimumLength, "min-length", _MinimumLength,
		options.ArgName("LENGTH"),
		options.Alias("l"),
		options.Description("Selects sequence records whose sequence length is equal or longer than lmin."))

	options.IntVar(&_MaximumLength, "max-length", _MaximumLength,
		options.ArgName("LENGTH"),
		options.Alias("L"),
		options.Description("Selects sequence records whose sequence length is equal or shorter than lmax."))

	options.IntVar(&_MinimumCount, "min-count", _MinimumCount,
		options.ArgName("COUNT"),
		options.Alias("c"),
		options.Description("Selects sequence records occuring at least min count time in the data set."))

	options.IntVar(&_MaximumCount, "max-count", _MaximumCount,
		options.ArgName("COUNT"),
		options.Alias("C"),
		options.Description("Selects sequence records occuring no more than max count time in the data set."))

	options.StringSliceVar(&_Predicats, "predicate", 1, 1,
		options.Alias("p"),
		options.ArgName("EXPRESSION"),
		options.Description("boolean expression to be evaluated for each sequence record. "+
			"The attribute keys defined for each sequence record can be used in the "+
			"expression as variable names. An extra variable named ‘sequence’ refers "+
			"to the sequence record itself. Several -p options can be used on the same "+
			"command line and in this last case, the selected sequence records will match "+
			"all constraints."))

	options.StringSliceVar(&_SequencePatterns, "sequence", 1, 1,
		options.Alias("s"),
		options.ArgName("PATTERN"),
		options.Description("Regular expression pattern to be tested against the sequence itself. The pattern is case insensitive."))

	options.StringSliceVar(&_DefinitionPatterns, "definition", 1, 1,
		options.Alias("D"),
		options.ArgName("PATTERN"),
		options.Description("Regular expression pattern to be tested against the sequence definition. The pattern is case insensitive."))

	options.StringSliceVar(&_IdPatterns, "identifier", 1, 1,
		options.Alias("I"),
		options.ArgName("PATTERN"),
		options.Description("Regular expression pattern to be tested against the sequence identifier. The pattern is case insensitive."))

	options.StringSliceVar(&_RequiredAttributes, "has-attribute", 1, 1,
		options.Alias("A"),
		options.ArgName("KEY"),
		options.Description("Selects sequence records having an attribute whose key = <KEY>."))

	options.StringMapVar(&_AttributePatterns, "attribute", 1, 1,
		options.Alias("a"),
		options.ArgName("KEY=VALUE"),
		options.Description("Regular expression pattern matched against the attributes of the sequence record. "+
			"the value of this attribute is of the form : key:regular_pattern. The pattern is case sensitive. "+
			"Several -a options can be used on the same command line and in this last case, the selected "+
			"sequence records will match all constraints."))

	options.StringVar(&_PairedMode, "paired-mode", _PairedMode,
		options.ArgName("forward|reverse|and|or|andnot|xor"),
		options.Description("If paired reads are passed to obibrep, that option determines how the conditions "+
			"are applied to both reads."),
	)

	options.StringSliceVar(&_approx_pattern, "approx-pattern", 1, 1,
		options.ArgName("PATTERN"),
		options.Description("Regular expression pattern to be tested against the sequence itself. The pattern is case insensitive."))

	options.IntVar(&_pattern_error, "pattern-error", _pattern_error,
		options.Description("Maximum number of allowed error during pattern matching"),
	)

	options.BoolVar(&_pattern_indel, "allows-indels", _pattern_indel,
		options.Description("Allows for indel during pattern matching"),
	)

	options.BoolVar(&_pattern_only_forward, "only-forward", _pattern_only_forward,
		options.Description("Look for pattern only on forward strand"),
	)

}

// OptionSet adds to the basic option set every options declared for
// the obipcr command
func OptionSet(options *getoptions.GetOpt) {
	obiconvert.OptionSet(options)
	SequenceSelectionOptionSet(options)

	options.StringVar(&_SaveRejected, "save-discarded", _SaveRejected,
		options.ArgName("FILENAME"),
		options.Description("Indicates in which file not selected sequences must be saved."))
}

func CLIMinSequenceLength() int {
	return _MinimumLength
}

func CLIMaxSequenceLength() int {
	return _MaximumLength
}

func CLISequenceSizePredicate() obiseq.SequencePredicate {
	if _MinimumLength > 1 {
		p := obiseq.IsLongerOrEqualTo(_MinimumLength)

		if _MaximumLength != int(2e9) {
			p = p.And(obiseq.IsShorterOrEqualTo(_MaximumLength))
		}

		return p
	}

	if _MaximumLength != int(2e9) {
		return obiseq.IsShorterOrEqualTo(_MaximumLength)
	}

	return nil
}

func CLIMinSequenceCount() int {
	return _MinimumCount
}

func CLIMaxSequenceCount() int {
	return _MaximumCount
}

func CLIRequiredRanks() []string {
	return _RequiredRanks
}

func CLISequenceCountPredicate() obiseq.SequencePredicate {
	if _MinimumCount > 1 {
		p := obiseq.IsMoreAbundantOrEqualTo(_MinimumCount)

		if _MaximumCount != int(2e9) {
			p = p.And(obiseq.IsLessAbundantOrEqualTo(_MaximumCount))
		}

		return p
	}

	if _MaximumLength != int(2e9) {
		return obiseq.IsLessAbundantOrEqualTo(_MaximumCount)
	}

	return nil
}

func CLISelectedNCBITaxDump() string {
	return _Taxdump
}

func CLIPatternError() int {
	return _pattern_error
}

func CLIPatternInDels() bool {
	return _pattern_indel
}

func CLIPatternBothStrand() bool {
	return !_pattern_only_forward
}

func CLIRestrictTaxonomyPredicate() obiseq.SequencePredicate {
	var p obiseq.SequencePredicate
	var p2 obiseq.SequencePredicate

	if len(_BelongTaxa) > 0 {
		taxonomy := obitax.DefaultTaxonomy()

		taxon, _, err := taxonomy.Taxon(_BelongTaxa[0])
		if err != nil {
			p = obiseq.IsSubCladeOfSlot(taxonomy, _BelongTaxa[0])
		} else {
			p = obiseq.IsSubCladeOf(taxonomy, taxon)
		}
		for _, staxid := range _BelongTaxa[1:] {
			taxon, _, err := taxonomy.Taxon(staxid)
			if err != nil {
				p2 = obiseq.IsSubCladeOfSlot(taxonomy, staxid)
			} else {
				p2 = obiseq.IsSubCladeOf(taxonomy, taxon)
			}

			p = p.Or(p2)
		}

		return p
	}

	return nil
}

func CLIIsValidTaxonomyPredicate() obiseq.SequencePredicate {
	if _ValidateTaxonomy {
		if !obidefault.HasSelectedTaxonomy() {
			log.Fatal("Taxonomy not found")
		}
		taxonomy := obitax.DefaultTaxonomy()
		if taxonomy == nil {
			log.Fatal("Taxonomy not found")
		}

		predicat := func(sequences *obiseq.BioSequence) bool {
			taxon := sequences.Taxon(taxonomy)
			return taxon != nil
		}

		return predicat
	}

	return nil
}

func CLIAvoidTaxonomyPredicate() obiseq.SequencePredicate {
	var p obiseq.SequencePredicate
	var p2 obiseq.SequencePredicate

	if len(_NotBelongTaxa) > 0 {
		taxonomy := obitax.DefaultTaxonomy()

		taxon, _, err := taxonomy.Taxon(_NotBelongTaxa[0])
		if err != nil {
			p = obiseq.IsSubCladeOfSlot(taxonomy, _NotBelongTaxa[0])
		} else {
			p = obiseq.IsSubCladeOf(taxonomy, taxon)
		}

		for _, taxid := range _NotBelongTaxa[1:] {
			taxon, _, err := taxonomy.Taxon(taxid)
			if err != nil {
				p2 = obiseq.IsSubCladeOfSlot(taxonomy, taxid)
			} else {
				p2 = obiseq.IsSubCladeOf(taxonomy, taxon)
			}

			p = p.Or(p2)
		}

		return p.Not()
	}

	return nil
}

func CLIHasRankDefinedPredicate() obiseq.SequencePredicate {

	if len(_RequiredRanks) > 0 {
		taxonomy := obitax.DefaultTaxonomy()
		p := obiseq.HasRequiredRank(taxonomy, _RequiredRanks[0])

		for _, rank := range _RequiredRanks[1:] {
			p = p.And(obiseq.HasRequiredRank(taxonomy, rank))
		}

		return p
	}

	return nil
}

func CLITaxonomyFilterPredicate() obiseq.SequencePredicate {
	return CLIIsValidTaxonomyPredicate().And(CLIAvoidTaxonomyPredicate()).And(CLIHasRankDefinedPredicate()).And(CLIRestrictTaxonomyPredicate())
}

func CLIPredicatesPredicate() obiseq.SequencePredicate {

	if len(_Predicats) > 0 {
		p := obiseq.ExpressionPredicat(_Predicats[0])

		for _, expression := range _Predicats[1:] {
			p = p.And(obiseq.ExpressionPredicat(expression))
		}

		return p
	}

	return nil
}

func CLISequencePatternPredicate() obiseq.SequencePredicate {

	if len(_SequencePatterns) > 0 {
		p := obiseq.IsSequenceMatch(_SequencePatterns[0])

		for _, pattern := range _SequencePatterns[1:] {
			p = p.And(obiseq.IsSequenceMatch(pattern))
		}

		return p
	}

	return nil
}

func CLISequenceAgrep() obiseq.SequencePredicate {
	if len(_approx_pattern) > 0 {
		p := obiapat.IsPatternMatchSequence(
			_approx_pattern[0],
			CLIPatternError(),
			CLIPatternBothStrand(),
			CLIPatternInDels(),
		)

		for _, pattern := range _approx_pattern[1:] {
			p = p.And(obiapat.IsPatternMatchSequence(
				pattern,
				CLIPatternError(),
				CLIPatternBothStrand(),
				CLIPatternInDels(),
			))
		}

		return p
	}

	return nil
}

func CLIDefinitionPatternPredicate() obiseq.SequencePredicate {

	if len(_DefinitionPatterns) > 0 {
		p := obiseq.IsDefinitionMatch(_DefinitionPatterns[0])

		for _, pattern := range _DefinitionPatterns[1:] {
			p = p.And(obiseq.IsDefinitionMatch(pattern))
		}

		return p
	}

	return nil
}

func CLIIdPatternPredicate() obiseq.SequencePredicate {

	if len(_IdPatterns) > 0 {
		p := obiseq.IsIdMatch(_IdPatterns[0])

		for _, pattern := range _IdPatterns[1:] {
			p = p.And(obiseq.IsIdMatch(pattern))
		}

		return p
	}

	return nil
}

func CLIIdListPredicate() obiseq.SequencePredicate {

	if _IdList != "" {
		ids, err := obiutils.ReadLines(_IdList)

		if err != nil {
			log.Fatalf("cannot read the id file %s : %v", _IdList, err)
		}

		for i, v := range ids {
			ids[i] = strings.TrimSpace(v)
		}

		p := obiseq.IsIdIn(ids...)

		return p
	}
	return nil
}

func CLIHasAttibutePredicate() obiseq.SequencePredicate {

	if len(_RequiredAttributes) > 0 {
		p := obiseq.HasAttribute(_RequiredAttributes[0])

		for _, rank := range _RequiredAttributes[1:] {
			p = p.And(obiseq.HasAttribute(rank))
		}

		return p
	}

	return nil
}

func CLIIsAttibuteMatchPredicate() obiseq.SequencePredicate {

	if len(_AttributePatterns) > 0 {
		p := obiseq.SequencePredicate(nil)

		for key, pattern := range _AttributePatterns {
			log.Println(key, pattern)
			p = p.And(obiseq.IsAttributeMatch(key, pattern))
		}

		return p
	}

	return nil
}

func CLISequenceSelectionPredicate() obiseq.SequencePredicate {
	p := CLISequenceSizePredicate()
	p = p.And(CLISequenceCountPredicate())
	p = p.And(CLITaxonomyFilterPredicate())
	p = p.And(CLIPredicatesPredicate())
	p = p.And(CLISequencePatternPredicate())
	p = p.And(CLIDefinitionPatternPredicate())
	p = p.And(CLIIdPatternPredicate())
	p = p.And(CLIIdListPredicate())
	p = p.And(CLIHasAttibutePredicate())
	p = p.And(CLIIsAttibuteMatchPredicate())
	p = p.And(CLISequenceAgrep())

	if _InvertMatch {
		p = p.Not()
	}

	return p
}

func CLISaveDiscardedSequences() bool {
	return _SaveRejected != ""
}

func CLIDiscardedFileName() string {
	return _SaveRejected
}

func CLIPairedReadMode() obiseq.SeqPredicateMode {
	switch _PairedMode {
	case "forward":
		return obiseq.ForwardOnly
	case "reverse":
		return obiseq.ReverseOnly
	case "and":
		return obiseq.And
	case "or":
		return obiseq.Or
	case "andnot":
		return obiseq.AndNot
	case "xor":
		return obiseq.Xor
	default:
		log.Fatalf("Paired reads mode must be forward, reverse, and, or, andnot, or xor (%s)", _PairedMode)
	}

	return obiseq.ForwardOnly
}
