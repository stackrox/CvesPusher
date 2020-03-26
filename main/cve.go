package main

import (
	"strings"

	"github.com/facebookincubator/nvdtools/cvefeed/nvd/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stackrox/k8s-istio-cve-pusher/nvd"
)

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
						Cpe23Uri: "cpe:2.3:a:redhat:openshift_container_platform:3.0.0:*:*:*:*:*:*:*",
					},
					{
						Cpe23Uri: "cpe:2.3:a:redhat:openshift_container_platform:3.1.0:*:*:*:*:*:*:*",
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
	}
)

func updateFixedByVersionIfMissing(project nvd.Project, cves ...*schema.NVDCVEFeedJSON10DefCVEItem) {
	for _, cve := range cves {
		if _, ok := cvesWithoutFixedBy[cve.CVE.CVEDataMeta.ID]; !ok {
			continue
		}

		updateRequired := true
		for _, node := range cve.Configurations.Nodes {
			for _, cpeMatch := range node.CPEMatch {
				cpeVersionAndUpdate, err := getVersionAndUpdateFromCPE(cpeMatch.Cpe23Uri, project)
				if err != nil {
					log.Error(errors.Wrapf(err, "could not get version and update from cpe: %q", cpeMatch.Cpe23Uri))
					continue
				}

				// Possibly invalid CPE or non-kube and non-istio CPE.
				if cpeVersionAndUpdate == "" {
					continue
				}

				// We need to update CVE only if none of the CPE provide any information about affected version
				if cpeVersionAndUpdate != "-:*" {
					// If affected version is specified move to next CPE
					if cpeVersionAndUpdate != "*:*" {
						continue
					}
					// If CPE string does not specify affected version, look for version range
					if cpeMatch.VersionStartIncluding != "" || cpeMatch.VersionEndIncluding != "" || cpeMatch.VersionEndExcluding != "" {
						updateRequired = false
					}
				}
			}
		}

		// For the concerned CVEs, if NVD is providing data, use it, else update CVE.
		if updateRequired {
			cve.Configurations.Nodes = append(cve.Configurations.Nodes, cvesWithoutFixedBy[cve.CVE.CVEDataMeta.ID]...)
		}
	}
}

func getVersionAndUpdateFromCPE(cpe string, project nvd.Project) (string, error) {
	if ok := strings.HasPrefix(cpe, "cpe:2.3:a:"); !ok {
		return "", errors.Errorf("cpe: %q not a valid cpe23Uri format", cpe)
	}

	ss := strings.Split(cpe, ":")
	if len(ss) != cpePartsLength {
		return "", errors.Errorf("cpe: %q not a valid cpe23Uri format", cpe)
	}
	if _, ok := specs[project]; !ok {
		return "", errors.Errorf("unknown CVE type: %s", project.String())
	}
	if project == nvd.Kubernetes && (ss[cpeVendorIndex] != "kubernetes" || ss[cpeProductIndex] != "kubernetes") {
		return "", nil
	}
	if project == nvd.Istio && (ss[cpeVendorIndex] != "istio" || ss[cpeProductIndex] != "istio") {
		return "", nil
	}
	if project == nvd.Openshift && (ss[cpeVendorIndex] != "redhat" || ss[cpeProductIndex] != "openshift_container_platform") {
		return "", nil
	}
	return strings.Join(ss[cpeVersionIndex:cpeUpdateIndex+1], ":"), nil
}
