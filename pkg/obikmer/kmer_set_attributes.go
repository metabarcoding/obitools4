package obikmer

import (
	"fmt"
	"strconv"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

// ==================================
// KMER SET ATTRIBUTE API
// Mimic BioSequence attribute API from obiseq/attributes.go
// ==================================

// HasAttribute vérifie si une clé d'attribut existe
func (ks *KmerSet) HasAttribute(key string) bool {
	_, ok := ks.Metadata[key]
	return ok
}

// GetAttribute récupère la valeur d'un attribut
// Cas particuliers: "id" utilise Id(), "k" utilise K()
func (ks *KmerSet) GetAttribute(key string) (interface{}, bool) {
	switch key {
	case "id":
		return ks.Id(), true
	case "k":
		return ks.K(), true
	default:
		value, ok := ks.Metadata[key]
		return value, ok
	}
}

// SetAttribute sets the value of an attribute
// Cas particuliers: "id" utilise SetId(), "k" est immutable (panique)
func (ks *KmerSet) SetAttribute(key string, value interface{}) {
	switch key {
	case "id":
		if id, ok := value.(string); ok {
			ks.SetId(id)
		} else {
			panic(fmt.Sprintf("id must be a string, got %T", value))
		}
	case "k":
		panic("k is immutable and cannot be modified via SetAttribute")
	default:
		ks.Metadata[key] = value
	}
}

// DeleteAttribute supprime un attribut
func (ks *KmerSet) DeleteAttribute(key string) {
	delete(ks.Metadata, key)
}

// RemoveAttribute supprime un attribut (alias de DeleteAttribute)
func (ks *KmerSet) RemoveAttribute(key string) {
	ks.DeleteAttribute(key)
}

// RenameAttribute renomme un attribut
func (ks *KmerSet) RenameAttribute(newName, oldName string) {
	if value, ok := ks.Metadata[oldName]; ok {
		ks.Metadata[newName] = value
		delete(ks.Metadata, oldName)
	}
}

// GetIntAttribute récupère un attribut en tant qu'entier
func (ks *KmerSet) GetIntAttribute(key string) (int, bool) {
	value, ok := ks.Metadata[key]
	if !ok {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

// GetFloatAttribute récupère un attribut en tant que float64
func (ks *KmerSet) GetFloatAttribute(key string) (float64, bool) {
	value, ok := ks.Metadata[key]
	if !ok {
		return 0, false
	}

	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// GetNumericAttribute récupère un attribut numérique (alias de GetFloatAttribute)
func (ks *KmerSet) GetNumericAttribute(key string) (float64, bool) {
	return ks.GetFloatAttribute(key)
}

// GetStringAttribute récupère un attribut en tant que chaîne
func (ks *KmerSet) GetStringAttribute(key string) (string, bool) {
	value, ok := ks.Metadata[key]
	if !ok {
		return "", false
	}

	switch v := value.(type) {
	case string:
		return v, true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

// GetBoolAttribute récupère un attribut en tant que booléen
func (ks *KmerSet) GetBoolAttribute(key string) (bool, bool) {
	value, ok := ks.Metadata[key]
	if !ok {
		return false, false
	}

	switch v := value.(type) {
	case bool:
		return v, true
	case int:
		return v != 0, true
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return b, true
		}
	}
	return false, false
}

// AttributeKeys returns the set of attribute keys
func (ks *KmerSet) AttributeKeys() obiutils.Set[string] {
	keys := obiutils.MakeSet[string]()
	for key := range ks.Metadata {
		keys.Add(key)
	}
	return keys
}

// Keys returns the set of attribute keys (alias of AttributeKeys)
func (ks *KmerSet) Keys() obiutils.Set[string] {
	return ks.AttributeKeys()
}

// ==================================
// KMER SET GROUP ATTRIBUTE API
// Métadonnées du groupe + accès via Get() pour les sets individuels
// ==================================

// HasAttribute vérifie si une clé d'attribut existe pour le groupe
func (ksg *KmerSetGroup) HasAttribute(key string) bool {
	_, ok := ksg.Metadata[key]
	return ok
}

// GetAttribute récupère la valeur d'un attribut du groupe
// Cas particuliers: "id" utilise Id(), "k" utilise K()
func (ksg *KmerSetGroup) GetAttribute(key string) (interface{}, bool) {
	switch key {
	case "id":
		return ksg.Id(), true
	case "k":
		return ksg.K(), true
	default:
		value, ok := ksg.Metadata[key]
		return value, ok
	}
}

// SetAttribute sets the value of an attribute du groupe
// Cas particuliers: "id" utilise SetId(), "k" est immutable (panique)
func (ksg *KmerSetGroup) SetAttribute(key string, value interface{}) {
	switch key {
	case "id":
		if id, ok := value.(string); ok {
			ksg.SetId(id)
		} else {
			panic(fmt.Sprintf("id must be a string, got %T", value))
		}
	case "k":
		panic("k is immutable and cannot be modified via SetAttribute")
	default:
		ksg.Metadata[key] = value
	}
}

// DeleteAttribute supprime un attribut du groupe
func (ksg *KmerSetGroup) DeleteAttribute(key string) {
	delete(ksg.Metadata, key)
}

// RemoveAttribute supprime un attribut du groupe (alias)
func (ksg *KmerSetGroup) RemoveAttribute(key string) {
	ksg.DeleteAttribute(key)
}

// RenameAttribute renomme un attribut du groupe
func (ksg *KmerSetGroup) RenameAttribute(newName, oldName string) {
	if value, ok := ksg.Metadata[oldName]; ok {
		ksg.Metadata[newName] = value
		delete(ksg.Metadata, oldName)
	}
}

// GetIntAttribute récupère un attribut entier du groupe
func (ksg *KmerSetGroup) GetIntAttribute(key string) (int, bool) {
	value, ok := ksg.GetAttribute(key)
	if !ok {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}

// GetFloatAttribute récupère un attribut float64 du groupe
func (ksg *KmerSetGroup) GetFloatAttribute(key string) (float64, bool) {
	value, ok := ksg.GetAttribute(key)
	if !ok {
		return 0, false
	}

	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// GetNumericAttribute récupère un attribut numérique du groupe
func (ksg *KmerSetGroup) GetNumericAttribute(key string) (float64, bool) {
	return ksg.GetFloatAttribute(key)
}

// GetStringAttribute récupère un attribut chaîne du groupe
func (ksg *KmerSetGroup) GetStringAttribute(key string) (string, bool) {
	value, ok := ksg.GetAttribute(key)
	if !ok {
		return "", false
	}

	switch v := value.(type) {
	case string:
		return v, true
	default:
		return fmt.Sprintf("%v", v), true
	}
}

// GetBoolAttribute récupère un attribut booléen du groupe
func (ksg *KmerSetGroup) GetBoolAttribute(key string) (bool, bool) {
	value, ok := ksg.GetAttribute(key)
	if !ok {
		return false, false
	}

	switch v := value.(type) {
	case bool:
		return v, true
	case int:
		return v != 0, true
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return b, true
		}
	}
	return false, false
}

// AttributeKeys returns the set of attribute keys du groupe
func (ksg *KmerSetGroup) AttributeKeys() obiutils.Set[string] {
	keys := obiutils.MakeSet[string]()
	for key := range ksg.Metadata {
		keys.Add(key)
	}
	return keys
}

// Keys returns the set of group attribute keys (alias)
func (ksg *KmerSetGroup) Keys() obiutils.Set[string] {
	return ksg.AttributeKeys()
}

// ==================================
// MÉTHODES POUR ACCÉDER AUX ATTRIBUTS DES SETS INDIVIDUELS VIA Get()
// Architecture zero-copy: ksg.Get(i).SetAttribute(...)
// ==================================

// Exemple d'utilisation:
// Pour accéder aux métadonnées d'un KmerSet individuel dans un groupe:
//   ks := ksg.Get(0)
//   ks.SetAttribute("level", 1)
//   hasLevel := ks.HasAttribute("level")
//
// Pour les métadonnées du groupe:
//   ksg.SetAttribute("name", "FrequencyFilter")
//   name, ok := ksg.GetStringAttribute("name")

// AllAttributeKeys returns all unique attribute keys of the group AND all its sets
func (ksg *KmerSetGroup) AllAttributeKeys() obiutils.Set[string] {
	keys := obiutils.MakeSet[string]()

	// Ajouter les clés du groupe
	for key := range ksg.Metadata {
		keys.Add(key)
	}

	// Ajouter les clés de chaque set
	for _, ks := range ksg.sets {
		for key := range ks.Metadata {
			keys.Add(key)
		}
	}

	return keys
}
