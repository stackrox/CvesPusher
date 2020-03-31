package nvd

import "github.com/facebookincubator/nvdtools/cvefeed/nvd/schema"

var (
	// cvesWithoutFixedBy represents NVD CVEs that may not have correct fixed by version
	cvesWithoutFixedBy = map[string][]*schema.NVDCVEFeedJSON10DefNode{
		// https://github.com/kubernetes/kubernetes/issues/19479
		"CVE-2016-1905": {
			{
				Operator: "OR",
				CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
					{
						Cpe23Uri:            "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
						VersionEndExcluding: "1.2.0",
						Vulnerable:          true,
					},
				},
			},
		},
		// https://access.redhat.com/errata/RHSA-2016:0351
		// https://access.redhat.com/errata/RHSA-2016:0070
		"CVE-2016-1906": {
			{
				Operator: "OR",
				CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
					{
						Cpe23Uri:   "cpe:2.3:a:redhat:openshift_container_platform:3.0.0:*:*:*:*:*:*:*",
						Vulnerable: true,
					},
					{
						Cpe23Uri:   "cpe:2.3:a:redhat:openshift_container_platform:3.1.0:*:*:*:*:*:*:*",
						Vulnerable: true,
					},
				},
			},
		},
		// https://github.com/kubernetes/kubernetes/issues/34517
		"CVE-2016-7075": {
			{
				Operator: "OR",
				CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
					{
						Cpe23Uri:            "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
						VersionEndExcluding: "1.2.7",
						Vulnerable:          true,
					},
					{
						Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
						VersionStartIncluding: "1.3.0",
						VersionEndExcluding:   "1.3.9",
						Vulnerable:            true,
					},
					{
						Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
						VersionStartIncluding: "1.4.0",
						VersionEndExcluding:   "1.4.3",
						Vulnerable:            true,
					},
				},
			},
		},
		"CVE-2020-8551": {
			{
				Operator: "OR",
				CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
					{
						Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
						VersionStartIncluding: "1.15.0",
						VersionEndIncluding:   "1.15.9",
						Vulnerable:            true,
					},
					{
						Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
						VersionStartIncluding: "1.16.0",
						VersionEndIncluding:   "1.16.6",
						Vulnerable:            true,
					},
					{
						Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
						VersionStartIncluding: "1.17.0",
						VersionEndIncluding:   "1.17.2",
						Vulnerable:            true,
					},
				},
			},
		},
		"CVE-2020-8552": {
			{
				Operator: "OR",
				CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
					{
						Cpe23Uri:            "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
						VersionEndExcluding: "1.15.9",
						Vulnerable:          true,
					},
					{
						Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
						VersionStartIncluding: "1.16.0",
						VersionEndIncluding:   "1.16.6",
						Vulnerable:            true,
					},
					{
						Cpe23Uri:            "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
						VersionEndExcluding: "1.17.0",
						VersionEndIncluding: "1.17.2",
						Vulnerable:          true,
					},
				},
			},
		},
	}
)
