package nvd

import (
	"strings"

	"github.com/facebookincubator/nvdtools/cvefeed/nvd/schema"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	cpeVersionIndex = 5
	cpeUpdateIndex  = 6
	cpePartsLength  = 13

	// ANY represent CPE ANY string punctuation
	ANY = "*"
	// NOT_APPLICABLE represents N/A CPE string punctuation
	NOT_APPLICABLE = "-"
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

func FillMissingData(cves ...*schema.NVDCVEFeedJSON10DefCVEItem) {
	tryUpdateCVEWithAffectedVersions(cves...)
}

func tryUpdateCVEWithAffectedVersions(cves ...*schema.NVDCVEFeedJSON10DefCVEItem) {
	for _, cve := range cves {
		if _, ok := cvesWithoutFixedBy[cve.CVE.CVEDataMeta.ID]; !ok {
			continue
		}

		updateRequired := true
		for _, node := range cve.Configurations.Nodes {
			for _, cpeMatch := range node.CPEMatch {
				hasVersionInfo, err := cpeHasVersionInfo(cpeMatch)
				if err != nil {
					log.Error(err)
					continue
				}

				if hasVersionInfo {
					updateRequired = false
				}
			}
		}

		// For the concerned CVEs, if NVD is providing data, use it, else update CVE.
		if updateRequired {
			cve.Configurations.Nodes = append(cve.Configurations.Nodes, cvesWithoutFixedBy[cve.CVE.CVEDataMeta.ID]...)
		}
	}
}

func getVersionAndUpdate(cpe string) (string, string, error) {
	if ok := strings.HasPrefix(cpe, "cpe:2.3:a:"); !ok {
		return "", "", errors.Errorf("cpe: %q not a valid cpe23Uri format", cpe)
	}

	ss := strings.Split(cpe, ":")
	if len(ss) != cpePartsLength {
		return "", "", errors.Errorf("cpe: %q not a valid cpe23Uri format", cpe)
	}
	return ss[cpeVersionIndex], ss[cpeUpdateIndex], nil
}

func cpeHasVersionInfo(cpeMatch *schema.NVDCVEFeedJSON10DefCPEMatch) (bool, error) {
	version, update, err := getVersionAndUpdate(cpeMatch.Cpe23Uri)
	if err != nil {
		return false, errors.Wrapf(err, "could not get version and update from cpe: %q", cpeMatch.Cpe23Uri)
	}

	// Possibly invalid CPE
	if version == "" && update == "" {
		return false, nil
	}

	// CPE string does not provide affected versions information
	if version == NOT_APPLICABLE {
		return false, nil
	}

	// If CPE string matches any version check if ranges are provided. Note: CPE23URI provides affected versions and not ranges.
	if version == ANY && update == ANY {
		return cpeMatch.VersionStartIncluding != "" || cpeMatch.VersionEndIncluding != "" || cpeMatch.VersionEndExcluding != "", nil
	}

	return true, nil
}
