package nvd

func (cves CVEs) FilterProject(project string) []CVEEntry {
	matcher := NewCPEMatcher(project + ":" + project)
	return cves.Filter(matcher)
}

func (cves CVEs) Filter(matcher Matcher) []CVEEntry {
	var filtered []CVEEntry
	for _, cve := range cves.Entries {
		if matcher.Matches(cve) {
			filtered = append(filtered, cve)
		}
	}
	return filtered
}
