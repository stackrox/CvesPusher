package main

import (
	"fmt"
	"io/ioutil"
	"testing"

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
	assert.Len(t, CVEs.Entries, 2)
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
	cves, err := nvd.Load(data)
	assert.Nil(t, err)

	tests := []struct {
		title       string
		substring   string
		expectedIDs []string
	}{
		{
			"match none",
			"initech",
			nil,
		},
		{
			"match first",
			"juniper:junos:17.1",
			[]string{"CVE-2019-0001"},
		},
		{
			"match second",
			"juniper:junos:15.1x53",
			[]string{"CVE-2019-0002"},
		},
		{
			"match both",
			"juniper:junos",
			[]string{"CVE-2019-0001", "CVE-2019-0002"},
		},
	}

	for index, test := range tests {
		name := fmt.Sprintf("%d-%s", index, test.title)
		t.Run(name, func(t *testing.T) {
			entries := cves.Filter(nvd.NewCPEMatcher(test.substring))

			assert.Len(t, entries, len(test.expectedIDs))
			for i, entry := range entries {
				assert.Equal(t, test.expectedIDs[i], entry.CVE.Metadata.CVEID)
			}
		})
	}

}

func TestChecksum(t *testing.T) {
	data, err := ioutil.ReadFile(correctCVEFile)
	require.Nil(t, err)
	cves, err := nvd.Load(data)
	assert.Nil(t, err)

	data, checksum, err := marshalCVEs(cves.Entries)
	assert.True(t, len(data) > 0)
	assert.Equal(t, "956be5c313cbeae497de4cab6d46a8ad2fc2a86fda6836d7abac114db9113854", checksum)
}
