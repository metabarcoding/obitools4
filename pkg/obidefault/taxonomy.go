package obidefault

var __taxonomy__ = ""
var __alternative_name__ = false
var __fail_on_taxonomy__ = false
var __update_taxid__ = false
var __raw_taxid__ = false

func UseRawTaxids() bool {
	return __raw_taxid__
}

func UseRawTaxidsPtr() *bool {
	return &__raw_taxid__
}

func SelectedTaxonomy() string {
	return __taxonomy__
}

func HasSelectedTaxonomy() bool {
	return __taxonomy__ != ""
}

func AreAlternativeNamesSelected() bool {
	return __alternative_name__
}

func SelectedTaxonomyPtr() *string {
	return &__taxonomy__
}

func AlternativeNamesSelectedPtr() *bool {
	return &__alternative_name__
}

func SetSelectedTaxonomy(taxonomy string) {
	__taxonomy__ = taxonomy
}

func SetAlternativeNamesSelected(alt bool) {
	__alternative_name__ = alt
}

func SetFailOnTaxonomy(fail bool) {
	__fail_on_taxonomy__ = fail
}

func SetUpdateTaxid(update bool) {
	__update_taxid__ = update
}

func FailOnTaxonomyPtr() *bool {
	return &__fail_on_taxonomy__
}

func UpdateTaxidPtr() *bool {
	return &__update_taxid__
}

func FailOnTaxonomy() bool {
	return __fail_on_taxonomy__
}

func UpdateTaxid() bool {
	return __update_taxid__
}
