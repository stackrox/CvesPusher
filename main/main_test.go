package main

import (
	"io/ioutil"
	"testing"

	"github.com/facebookincubator/nvdtools/cvefeed/nvd/schema"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stackrox/k8s-istio-cve-pusher/nvd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	malformedCVEFile = "testdata/malformedCVEs.json"
	correctCVEFile   = "testdata/correctCVEs.json"
)

func TestNVDParseRightFile(t *testing.T) {
	dat, err := ioutil.ReadFile(correctCVEFile)
	require.Nil(t, err)
	CVEs, err := nvd.Load(dat)
	assert.Nil(t, err)
	assert.Len(t, CVEs.CVEItems, 2)
}

func TestNVDParseErrorFile(t *testing.T) {
	dat, err := ioutil.ReadFile(malformedCVEFile)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = nvd.Load(dat)
	assert.NotNil(t, err)
}

func TestNVDFilter(t *testing.T) {
	data, err := ioutil.ReadFile(correctCVEFile)
	require.Nil(t, err)
	cveFeed, err := nvd.Load(data)
	assert.Nil(t, err)

	type test struct {
		title      string
		project    nvd.Project
		expectedID string
	}

	tests := []test{
		{
			"kubernetes",
			nvd.Kubernetes,
			"CVE-2019-9946",
		},
		{
			"openshift",
			nvd.Openshift,
			"CVE-2019-3884",
		},
	}

	for _, test := range tests {
		for _, cve := range cveFeed.CVEItems {
			if nvd.CVEAppliesToProject(test.project, cve) {
				assert.Equal(t, test.expectedID, cve.CVE.CVEDataMeta.ID)
			}
		}
	}
}

func TestChecksum(t *testing.T) {
	data, err := ioutil.ReadFile(correctCVEFile)
	require.Nil(t, err)
	cves, err := nvd.Load(data)
	require.Nil(t, err)

	data, checksum, err := marshalCVEs(cves.CVEItems)
	require.Nil(t, err)
	assert.True(t, len(data) > 0)
	assert.Equal(t, "8a8c5bce18d9640a5bac6e0ec553c5d67ae25645a9f580fea2f6362ef19ff0e7", checksum)
}

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
					Vulnerable: true,
					Cpe23Uri:   "cpe:2.3:a:kubernetes:kubernetes:-:*:*:*:*:*:*:*",
				},
			},
		},
		{
			Operator: "OR",
			CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
				{
					Cpe23Uri:            "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
					VersionEndExcluding: "1.2.7",
				},
				{
					Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
					VersionStartIncluding: "1.3.0",
					VersionEndExcluding:   "1.3.9",
				},
				{
					Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
					VersionStartIncluding: "1.4.0",
					VersionEndExcluding:   "1.4.3",
				},
			},
		},
	}

	updateFixedByVersionIfMissing(nvd.Kubernetes, cve)

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
						},
						{
							Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
							VersionStartIncluding: "1.3.0",
							VersionEndExcluding:   "1.3.9",
						},
						{
							Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
							VersionStartIncluding: "1.4.0",
							VersionEndExcluding:   "1.4.3",
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
				},
				{
					Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
					VersionStartIncluding: "1.3.0",
					VersionEndExcluding:   "1.3.9",
				},
				{
					Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
					VersionStartIncluding: "1.4.0",
					VersionEndExcluding:   "1.4.3",
				},
			},
		},
	}

	updateFixedByVersionIfMissing(nvd.Kubernetes, cve)

	assert.EqualValues(t, expectedCPENodes, cve.Configurations.Nodes)
}
