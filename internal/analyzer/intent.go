package analyzer

import (
	"bufio"
	"os"
	"strings"
)

type Intent string

const (
	IntentPlanned Intent = "planned"
	IntentUsed    Intent = "used"
	IntentIgnore  Intent = "ignore"
)

// ParseIntentTags finds @intent annotations in code
// Example: // @intent:planned
func ParseIntentTags(filePath string) map[string]Intent {
	intents := make(map[string]Intent)

	f, err := os.Open(filePath)
	if err != nil {
		return intents
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var currentFunc string

	for scanner.Scan() {
		line := scanner.Text()

		// Check for intent tags
		if strings.Contains(line, "@intent:") {
			if strings.Contains(line, "@intent:planned") {
				intents[currentFunc] = IntentPlanned
			} else if strings.Contains(line, "@intent:used") {
				intents[currentFunc] = IntentUsed
			} else if strings.Contains(line, "@intent:ignore") {
				intents[currentFunc] = IntentIgnore
			}
		}

		// Track current function (simple heuristic)
		if strings.Contains(line, "func ") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "func" && i+1 < len(parts) {
					funcName := strings.Split(parts[i+1], "(")[0]
					currentFunc = funcName
					break
				}
			}
		}
	}

	return intents
}
