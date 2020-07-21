package nvd

import (
	"testing"

	"github.com/facebookincubator/nvdtools/cvefeed/nvd/schema"
	"github.com/stretchr/testify/assert"
)

func TestUpdateFixedByWithNoDataFromNVD(t *testing.T) {
	cve := &schema.NVDCVEFeedJSON10DefCVEItem{
		CVE: &schema.CVEJSON40{
			CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
				ID: "CVE-2016-7075",
			},
		},
		Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
			Nodes: []*schema.NVDCVEFeedJSON10DefNode{
				{
					Operator: "OR",
					CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
						{
							Vulnerable: true,
							Cpe23Uri:   "cpe:2.3:a:kubernetes:kubernetes:-:*:*:*:*:*:*:*",
						},
					},
				},
			},
		},
	}
	expectedCPENodes := []*schema.NVDCVEFeedJSON10DefNode{
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
	}

	tryUpdateCVEWithAffectedVersions([]*schema.NVDCVEFeedJSON10DefCVEItem{cve})
	assert.EqualValues(t, expectedCPENodes, cve.Configurations.Nodes)

}

func TestUpdateFixedByWithDataFromNVD(t *testing.T) {
	cve := &schema.NVDCVEFeedJSON10DefCVEItem{
		CVE: &schema.CVEJSON40{
			CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
				ID: "CVE-2016-7075",
			},
		},
		Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
			Nodes: []*schema.NVDCVEFeedJSON10DefNode{
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
		},
	}
	expectedCPENodes := []*schema.NVDCVEFeedJSON10DefNode{
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
	}

	tryUpdateCVEWithAffectedVersions([]*schema.NVDCVEFeedJSON10DefCVEItem{cve})
	assert.EqualValues(t, expectedCPENodes, cve.Configurations.Nodes)
}

func TestAddMissing(t *testing.T) {
	incomingCVEs := []*schema.NVDCVEFeedJSON10DefCVEItem{
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "Blah",
				},
			},
		},
	}

	actual := addMissingCVEs(Kubernetes, incomingCVEs)
	assert.EqualValues(t, cvesNotInDataFeed[Kubernetes]["CVE-2020-8551"].Configurations.Nodes, actual[1].Configurations.Nodes)
	assert.EqualValues(t, cvesNotInDataFeed[Kubernetes]["CVE-2020-8552"].Configurations.Nodes, actual[2].Configurations.Nodes)
}

func TestAddMissingWithUpdate(t *testing.T) {
	incomingCVEs := []*schema.NVDCVEFeedJSON10DefCVEItem{
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "CVE-2020-8551",
				},
			},
			Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
				Nodes: []*schema.NVDCVEFeedJSON10DefNode{},
			},
		},
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "CVE-2020-8552",
				},
			},
			Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
				Nodes: []*schema.NVDCVEFeedJSON10DefNode{},
			},
		},
	}

	expectedCVEs := []*schema.NVDCVEFeedJSON10DefCVEItem{
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "CVE-2020-8551",
				},
			},
			Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
				Nodes: []*schema.NVDCVEFeedJSON10DefNode{},
			},
		},
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "CVE-2020-8552",
				},
			},
			Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
				Nodes: []*schema.NVDCVEFeedJSON10DefNode{},
			},
		},
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "CVE-2020-10749",
				},
				Description: &schema.CVEJSON40Description{
					DescriptionData: []*schema.CVEJSON40LangString{
						{
							Value: "The Kubelet component in versions prior to v1.16.11, v1.17.0-v1.17.6, and v1.18.0-v1.18.3 have an affected kubernetes-cni package that has been found vulnerable to man-in-the-middle attacks.",
						},
					},
				},
			},
			Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
				Nodes: []*schema.NVDCVEFeedJSON10DefNode{},
			},
		},
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "CVE-2020-8557",
				},
				Description: &schema.CVEJSON40Description{
					DescriptionData: []*schema.CVEJSON40LangString{
						{
							Value: "The /etc/hosts file mounted in a pod by kubelet is not included by the kubelet eviction manager when calculating ephemeral storage usage by a pod. If a pod writes a large amount of data to the /etc/hosts file, it could fill the storage space of the node and cause the node to fail. This affects kublet v1.18.0-1.18.5, kubelet v1.17.0-1.17.8, and kubelet < v1.16.13.",
						},
					},
				},
			},
			Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
				Nodes: []*schema.NVDCVEFeedJSON10DefNode{},
			},
		},
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "CVE-2020-8559",
				},
				Description: &schema.CVEJSON40Description{
					DescriptionData: []*schema.CVEJSON40LangString{
						{
							Value: "If an attacker is able to intercept certain requests to the Kubelet, they can send a redirect response that may be followed by a client using the credentials from the original request. This can lead to compromise of other nodes. If multiple clusters share the same certificate authority trusted by the client, and the same authentication credentials, this vulnerability may allow an attacker to redirect the client to another cluster. In this configuration, this vulnerability should be considered High severity. This affects kube-apiserver v1.18.0-1.18.5, kube-apiserver v1.17.0-1.17.8, and all kube-apiserver versions prior to v1.16.0.",
						},
					},
				},
			},
			Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
				Nodes: []*schema.NVDCVEFeedJSON10DefNode{},
			},
		},
	}

	actual := addMissingCVEs(Kubernetes, incomingCVEs)
	assert.EqualValues(t, expectedCVEs, actual)

	tryUpdateCVEWithAffectedVersions(actual)
	assert.EqualValues(t, cvesNotInDataFeed[Kubernetes]["CVE-2020-8551"].Configurations.Nodes, actual[0].Configurations.Nodes)
	assert.EqualValues(t, cvesNotInDataFeed[Kubernetes]["CVE-2020-8552"].Configurations.Nodes, actual[1].Configurations.Nodes)
}
