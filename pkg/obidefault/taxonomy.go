package obidefault

var __taxonomy__ = ""
var __alternative_name__ = false

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
