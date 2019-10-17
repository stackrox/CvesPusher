package nvd

import "strings"

type Matcher interface {
	Matches(CVEEntry) bool
}

var _ Matcher = (*cpeMatcher)(nil)

type cpeMatcher struct {
	substring string
}

func NewCPEMatcher(substring string) cpeMatcher {
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
