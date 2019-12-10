package main

import (
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

	tests := []struct {
		title   string
		project nvd.Project
	}{
		{
			"kubernetes",
			nvd.Kubernetes,
		},
		{
			"openshift",
			nvd.Openshift,
		},
	}

	for index, test := range tests {
		actual := nvd.CVEAppliesToProject(test.project, cveFeed.CVEItems[index])
		assert.True(t, actual)
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
