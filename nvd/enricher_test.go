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

	addMissingCVEs(Kubernetes, &incomingCVEs)
	assert.EqualValues(t, cvesNotInDataFeed[Kubernetes]["CVE-2020-8551"].Configurations.Nodes, incomingCVEs[1].Configurations.Nodes)
	assert.EqualValues(t, cvesNotInDataFeed[Kubernetes]["CVE-2020-8552"].Configurations.Nodes, incomingCVEs[2].Configurations.Nodes)
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
	}

	addMissingCVEs(Kubernetes, &incomingCVEs)
	assert.EqualValues(t, expectedCVEs, incomingCVEs)

	tryUpdateCVEWithAffectedVersions(incomingCVEs)
	assert.EqualValues(t, cvesNotInDataFeed[Kubernetes]["CVE-2020-8551"].Configurations.Nodes, incomingCVEs[0].Configurations.Nodes)
	assert.EqualValues(t, cvesNotInDataFeed[Kubernetes]["CVE-2020-8552"].Configurations.Nodes, incomingCVEs[1].Configurations.Nodes)
}
