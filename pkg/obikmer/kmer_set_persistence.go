package obikmer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// MetadataFormat représente le format de sérialisation des métadonnées
type MetadataFormat int

const (
	FormatTOML MetadataFormat = iota
	FormatYAML
	FormatJSON
)

// String retourne l'extension de fichier pour le format
func (f MetadataFormat) String() string {
	switch f {
	case FormatTOML:
		return "toml"
	case FormatYAML:
		return "yaml"
	case FormatJSON:
		return "json"
	default:
		return "toml"
	}
}

// KmerSetMetadata contient les métadonnées d'un KmerSet ou KmerSetGroup
type KmerSetMetadata struct {
	K              int                      `toml:"k" yaml:"k" json:"k"`                                     // Taille des k-mers
	Type           string                   `toml:"type" yaml:"type" json:"type"`                            // "KmerSet" ou "KmerSetGroup"
	Size           int                      `toml:"size" yaml:"size" json:"size"`                            // 1 pour KmerSet, n pour KmerSetGroup
	Files          []string                 `toml:"files" yaml:"files" json:"files"`                         // Liste des fichiers .roaring
	UserMetadata   map[string]interface{}   `toml:"user_metadata,omitempty" yaml:"user_metadata,omitempty" json:"user_metadata,omitempty"`       // Métadonnées KmerSet unique
	SetsMetadata   []map[string]interface{} `toml:"sets_metadata,omitempty" yaml:"sets_metadata,omitempty" json:"sets_metadata,omitempty"`       // Métadonnées par set (KmerSetGroup)
}

// SaveKmerSet sauvegarde un KmerSet dans un répertoire
// Format: directory/metadata.{toml,yaml,json} + directory/set_0.roaring
func (ks *KmerSet) Save(directory string, format MetadataFormat) error {
	// Créer le répertoire si nécessaire
	if err := os.MkdirAll(directory, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", directory, err)
	}

	// Métadonnées
	metadata := KmerSetMetadata{
		K:            ks.K,
		Type:         "KmerSet",
		Size:         1,
		Files:        []string{"set_0.roaring"},
		UserMetadata: ks.Metadata, // Sauvegarder les métadonnées utilisateur
	}

	// Sauvegarder les métadonnées
	if err := saveMetadata(filepath.Join(directory, "metadata."+format.String()), metadata, format); err != nil {
		return err
	}

	// Sauvegarder le bitmap
	bitmapPath := filepath.Join(directory, "set_0.roaring")
	file, err := os.Create(bitmapPath)
	if err != nil {
		return fmt.Errorf("failed to create bitmap file %s: %w", bitmapPath, err)
	}
	defer file.Close()

	if _, err := ks.bitmap.WriteTo(file); err != nil {
		return fmt.Errorf("failed to write bitmap: %w", err)
	}

	return nil
}

// LoadKmerSet charge un KmerSet depuis un répertoire
func LoadKmerSet(directory string) (*KmerSet, error) {
	// Lire les métadonnées (essayer tous les formats)
	metadata, err := loadMetadata(directory)
	if err != nil {
		return nil, err
	}

	// Vérifier le type
	if metadata.Type != "KmerSet" {
		return nil, fmt.Errorf("invalid type: expected KmerSet, got %s", metadata.Type)
	}

	// Vérifier qu'il n'y a qu'un seul fichier
	if metadata.Size != 1 || len(metadata.Files) != 1 {
		return nil, fmt.Errorf("KmerSet must have exactly 1 bitmap file, got %d", len(metadata.Files))
	}

	// Charger le bitmap
	bitmapPath := filepath.Join(directory, metadata.Files[0])
	file, err := os.Open(bitmapPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bitmap file %s: %w", bitmapPath, err)
	}
	defer file.Close()

	ks := NewKmerSet(metadata.K)

	// Charger les métadonnées utilisateur
	if metadata.UserMetadata != nil {
		ks.Metadata = metadata.UserMetadata
	}

	if _, err := ks.bitmap.ReadFrom(file); err != nil {
		return nil, fmt.Errorf("failed to read bitmap: %w", err)
	}

	return ks, nil
}

// SaveKmerSetGroup sauvegarde un KmerSetGroup dans un répertoire
// Format: directory/metadata.{toml,yaml,json} + directory/set_0.roaring, set_1.roaring, ...
func (ksg *KmerSetGroup) Save(directory string, format MetadataFormat) error {
	// Créer le répertoire si nécessaire
	if err := os.MkdirAll(directory, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", directory, err)
	}

	// Métadonnées
	files := make([]string, len(ksg.sets))
	for i := range ksg.sets {
		files[i] = fmt.Sprintf("set_%d.roaring", i)
	}

	metadata := KmerSetMetadata{
		K:            ksg.K,
		Type:         "KmerSetGroup",
		Size:         len(ksg.sets),
		Files:        files,
		SetsMetadata: ksg.Metadata, // Sauvegarder les métadonnées de chaque set
	}

	// Sauvegarder les métadonnées
	if err := saveMetadata(filepath.Join(directory, "metadata."+format.String()), metadata, format); err != nil {
		return err
	}

	// Sauvegarder chaque bitmap
	for i, ks := range ksg.sets {
		bitmapPath := filepath.Join(directory, files[i])
		file, err := os.Create(bitmapPath)
		if err != nil {
			return fmt.Errorf("failed to create bitmap file %s: %w", bitmapPath, err)
		}

		if _, err := ks.bitmap.WriteTo(file); err != nil {
			file.Close()
			return fmt.Errorf("failed to write bitmap %d: %w", i, err)
		}
		file.Close()
	}

	return nil
}

// LoadKmerSetGroup charge un KmerSetGroup depuis un répertoire
func LoadKmerSetGroup(directory string) (*KmerSetGroup, error) {
	// Lire les métadonnées (essayer tous les formats)
	metadata, err := loadMetadata(directory)
	if err != nil {
		return nil, err
	}

	// Vérifier le type
	if metadata.Type != "KmerSetGroup" {
		return nil, fmt.Errorf("invalid type: expected KmerSetGroup, got %s", metadata.Type)
	}

	// Vérifier la cohérence
	if metadata.Size != len(metadata.Files) {
		return nil, fmt.Errorf("size mismatch: size=%d but %d files listed", metadata.Size, len(metadata.Files))
	}

	// Créer le groupe
	ksg := NewKmerSetGroup(metadata.K, metadata.Size)

	// Charger les métadonnées de chaque set
	if metadata.SetsMetadata != nil {
		if len(metadata.SetsMetadata) != metadata.Size {
			return nil, fmt.Errorf("metadata size mismatch: expected %d, got %d", metadata.Size, len(metadata.SetsMetadata))
		}
		ksg.Metadata = metadata.SetsMetadata
	}

	// Charger chaque bitmap
	for i, filename := range metadata.Files {
		bitmapPath := filepath.Join(directory, filename)
		file, err := os.Open(bitmapPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open bitmap file %s: %w", bitmapPath, err)
		}

		if _, err := ksg.sets[i].bitmap.ReadFrom(file); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to read bitmap %d: %w", i, err)
		}
		file.Close()
	}

	return ksg, nil
}

// saveMetadata sauvegarde les métadonnées dans le format spécifié
func saveMetadata(path string, metadata KmerSetMetadata, format MetadataFormat) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create metadata file %s: %w", path, err)
	}
	defer file.Close()

	var encoder interface{ Encode(interface{}) error }

	switch format {
	case FormatTOML:
		encoder = toml.NewEncoder(file)
	case FormatYAML:
		encoder = yaml.NewEncoder(file)
	case FormatJSON:
		jsonEncoder := json.NewEncoder(file)
		jsonEncoder.SetIndent("", "  ")
		encoder = jsonEncoder
	default:
		return fmt.Errorf("unsupported format: %v", format)
	}

	if err := encoder.Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	return nil
}

// loadMetadata charge les métadonnées depuis un répertoire
// Essaie tous les formats (TOML, YAML, JSON) dans l'ordre
func loadMetadata(directory string) (*KmerSetMetadata, error) {
	formats := []MetadataFormat{FormatTOML, FormatYAML, FormatJSON}

	var lastErr error
	for _, format := range formats {
		path := filepath.Join(directory, "metadata."+format.String())

		// Vérifier si le fichier existe
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		metadata, err := loadMetadataFromFile(path, format)
		if err != nil {
			lastErr = err
			continue
		}
		return metadata, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to load metadata: %w", lastErr)
	}
	return nil, fmt.Errorf("no metadata file found in %s (tried .toml, .yaml, .json)", directory)
}

// loadMetadataFromFile charge les métadonnées depuis un fichier spécifique
func loadMetadataFromFile(path string, format MetadataFormat) (*KmerSetMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata file %s: %w", path, err)
	}
	defer file.Close()

	var metadata KmerSetMetadata
	var decoder interface{ Decode(interface{}) error }

	switch format {
	case FormatTOML:
		decoder = toml.NewDecoder(file)
	case FormatYAML:
		decoder = yaml.NewDecoder(file)
	case FormatJSON:
		decoder = json.NewDecoder(file)
	default:
		return nil, fmt.Errorf("unsupported format: %v", format)
	}

	if err := decoder.Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &metadata, nil
}

// DetectFormat détecte le format des métadonnées dans un répertoire
func DetectFormat(directory string) (MetadataFormat, error) {
	formats := []MetadataFormat{FormatTOML, FormatYAML, FormatJSON}

	for _, format := range formats {
		path := filepath.Join(directory, "metadata."+format.String())
		if _, err := os.Stat(path); err == nil {
			return format, nil
		}
	}

	return FormatTOML, fmt.Errorf("no metadata file found in %s", directory)
}

// IsKmerSetDirectory vérifie si un répertoire contient un KmerSet ou KmerSetGroup
func IsKmerSetDirectory(directory string) (bool, string, error) {
	metadata, err := loadMetadata(directory)
	if err != nil {
		return false, "", err
	}

	return true, metadata.Type, nil
}

// ListBitmapFiles liste tous les fichiers .roaring dans un répertoire
func ListBitmapFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", directory, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".roaring") {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}
