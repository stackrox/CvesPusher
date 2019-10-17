package nvd

import "strings"

type Matcher interface {
	Matches(CVEEntry) bool
}

type cpeMatcher struct {
	substring string
}

func NewCPEMatcher(substring string) Matcher {
	return cpeMatcher{
		substring: substring,
	}
}

func (m cpeMatcher) Matches(entry CVEEntry) bool {
	for _, node := range entry.Configurations.Nodes {
		for _, cpeMatch := range node.CPEMatch {
			if strings.Contains(cpeMatch.CPE23Uri, m.substring) {
				return true
			}
		}
	}
	return false
}
