package obitax

import (
	"errors"
	"strings"
)

// ParseTaxonString parses a string in the format "code:taxid [scientific name]@rank"
// and returns the individual components. It handles extra whitespace around components.
//
// Parameters:
//   - taxonStr: The string to parse in the format "code:taxid [scientific name]@rank"
//
// Returns:
//   - code: The taxonomy code
//   - taxid: The taxon identifier
//   - scientificName: The scientific name (without brackets)
//   - rank: The rank
//   - error: An error if the string format is invalid
func ParseTaxonString(taxonStr string) (code, taxid, scientificName, rank string, err error) {
	// Trim any leading/trailing whitespace from the entire string
	taxonStr = strings.TrimSpace(taxonStr)

	// Split by '@' to separate rank
	parts := strings.Split(taxonStr, "@")
	if len(parts) > 2 {
		return "", "", "", "", errors.New("invalid format: multiple '@' characters found")
	}

	mainPart := strings.TrimSpace(parts[0])
	if len(parts) == 2 {
		rank = strings.TrimSpace(parts[1])
	} else {
		rank = "no rank"
	}

	// Find scientific name part (enclosed in square brackets)
	startBracket := strings.Index(mainPart, "[")
	endBracket := strings.LastIndex(mainPart, "]")

	if startBracket == -1 || endBracket == -1 || startBracket > endBracket {
		return "", "", "", "", errors.New("invalid format: scientific name must be enclosed in square brackets")
	}

	// Extract and clean scientific name
	scientificName = strings.TrimSpace(mainPart[startBracket+1 : endBracket])

	// Process code:taxid part
	idPart := strings.TrimSpace(mainPart[:startBracket])
	idComponents := strings.Split(idPart, ":")

	if len(idComponents) != 2 {
		return "", "", "", "", errors.New("invalid format: missing taxonomy code separator ':'")
	}

	code = strings.TrimSpace(idComponents[0])
	taxid = strings.TrimSpace(idComponents[1])

	if code == "" || taxid == "" || scientificName == "" {
		return "", "", "", "", errors.New("invalid format: code, taxid and scientific name cannot be empty")
	}

	return code, taxid, scientificName, rank, nil
}
