package nvd

import "github.com/facebookincubator/nvdtools/cvefeed/nvd/schema"

var (
	// cvesNotInDataFeed represents NVD CVEs that may not be available in NVD data feeds
	cvesNotInDataFeed = map[Project]map[string]*schema.NVDCVEFeedJSON10DefCVEItem{
		Kubernetes: {
			"CVE-2020-8551": {
				PublishedDate:    "2020-03-27T15:15Z",
				LastModifiedDate: "2020-03-27T16:03Z",
				CVE: &schema.CVEJSON40{
					CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
						ID: "CVE-2020-8551",
					},
					Description: &schema.CVEJSON40Description{
						DescriptionData: []*schema.CVEJSON40LangString{
							{
								Value: "The Kubelet component in versions 1.15.0-1.15.9, 1.16.0-1.16.6, and 1.17.0-1.17.2 has been found to be vulnerable to a denial of service attack via the kubelet API, including the unauthenticated HTTP read-only API typically served on port 10255, and the authenticated HTTPS API typically served on port 10250.",
							},
						},
					},
				},
				Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
					Nodes: []*schema.NVDCVEFeedJSON10DefNode{
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
				},
			},
			"CVE-2020-8552": {
				PublishedDate:    "2020-03-27T15:15Z",
				LastModifiedDate: "2020-03-27T16:03Z",
				CVE: &schema.CVEJSON40{
					CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
						ID: "CVE-2020-8552",
					},
					Description: &schema.CVEJSON40Description{
						DescriptionData: []*schema.CVEJSON40LangString{
							{
								Value: "The Kubernetes API server component in versions prior to 1.15.9, 1.16.0-1.16.6, and 1.17.0-1.17.2 has been found to be vulnerable to a denial of service attack via successful API requests.",
							},
						},
					},
				},
				Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
					Nodes: []*schema.NVDCVEFeedJSON10DefNode{
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
				},
			},
		},
	}
)
