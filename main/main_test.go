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
