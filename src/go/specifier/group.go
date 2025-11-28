package specifier

import "log/slog"

type SpecifierGroup struct {
	Specifiers []Specifier
}

func (sg *SpecifierGroup) AsMap() map[string]string {
	specifierMap := make(map[string]string)
	if sg.Specifiers == nil {
		return specifierMap
	}

	for _, specifier := range sg.Specifiers {
		if existingValue, ok := specifierMap[specifier.Key]; ok {
			slog.Warn("Specifier Key being overwritten",
				"identifier", specifier.Key,
				"existingValue", existingValue,
				"newValue", specifier.Value,
			)
		}
		specifierMap[specifier.Key] = specifier.Value
	}
	return specifierMap
}
