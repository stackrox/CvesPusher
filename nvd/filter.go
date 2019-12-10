package nvd

import "github.com/stackrox/nvdtools/cvefeed/nvd/schema"

func CVEAppliesToProject(project Project, cve *schema.NVDCVEFeedJSON10DefCVEItem) bool {
	matcher := NewCPEMatcher(project.Vendor() + ":" + project.String())
	return matcher.Matches(cve)
}
