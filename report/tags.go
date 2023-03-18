package report

import (
	"github.com/avahidi/molly/types"
)

// ExtractTags gathers tags from a file match
func ExtractTags(fr *types.FileData) []string {
	var ret []string
	seen := make(map[string]bool)

	for _, match := range fr.Matches {
		tags := ExtractTagsFromRule(match.Rule)
		for _, tag := range tags {
			if _, old := seen[tag]; !old {
				seen[tag] = true
				ret = append(ret, tag)
			}
		}
	}
	return ret
}

// ExtractReverseTagHierarchy creates file hierarchy for file -> tags
func ExtractReverseTagHierarchy(mr *types.Report) map[string][]string {
	ret := make(map[string][]string)
	for _, fr := range mr.Files {
		tags := ExtractTags(fr)
		if len(tags) > 0 {
			ret[fr.Filename] = tags
		}
	}
	return ret
}

// ExtractTagHierarchy creates file hierarchy for tag -> files
func ExtractTagHierarchy(mr *types.Report) map[string][]string {
	ret := make(map[string][]string)
	for _, fr := range mr.Files {
		tags := ExtractTags(fr)
		for _, tag := range tags {
			ret[tag] = append(ret[tag], fr.Filename)
		}
	}
	return ret
}
