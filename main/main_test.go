package main

import (
	"github.com/opentracing/opentracing-go/log"
	"github.com/stackrox/CvesPusher/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

const (
	malformedCVEFile = "scratch/malformedCVEs"
	correctCVEFile   = "scratch/correctCVEs"
)

func TestEnvVarErrors(t *testing.T) {
	_, err := utils.IsEnvVarNonEmpty("SOME_LOCAL_PATH")
	assert.NotNil(t, err)

	_, err = utils.IsEnvVarNonEmpty("SOME_GCLOUD_CONFIG")
	assert.NotNil(t, err)
}

func TestNVDParseRightFile(t *testing.T) {
	dat, err := ioutil.ReadFile(correctCVEFile)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = utils.GetCVEs(dat)
	assert.Nil(t, err)
}

func TestNVDParseErrorFile(t *testing.T) {
	dat, err := ioutil.ReadFile(malformedCVEFile)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = utils.GetCVEs(dat)
	assert.NotNil(t, err)
}
