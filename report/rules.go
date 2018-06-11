package report

import (
	"strings"

	"bitbucket.org/vahidi/molly/types"
)

// ExtractTagsFromRule extracts tags from a rule
func ExtractTagsFromRule(rule *types.Rule) []string {
	var ret []string
	if tagmeta, valid := rule.Metadata.GetString("tag", ""); valid {
		tags := strings.Split(tagmeta, ",")
		for _, tag := range tags {
			tag = strings.Trim(tag, " \t")
			ret = append(ret, tag)
		}
	}
	return ret
}
