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

	// Any represent CPE Any string punctuation
	Any = "*"
	// NotApplicable represents N/A CPE string punctuation
	NotApplicable = "-"
)

// FillMissingData fills NVD CVEs with missing data, e.g. affected product versions
func FillMissingData(project Project, cves []*schema.NVDCVEFeedJSON10DefCVEItem) []*schema.NVDCVEFeedJSON10DefCVEItem {
	cves = filterCVEs(project, addMissingCVEs(project, cves))
	tryUpdateCVEWithAffectedVersions(cves)
	if project == Istio {
		// Only do this for Istio to ensure other vulns are not affected at this time.
		populateCVSSv2IfMissing(cves)
	}
	return cves
}

func filterCVEs(project Project, cves []*schema.NVDCVEFeedJSON10DefCVEItem) []*schema.NVDCVEFeedJSON10DefCVEItem {
	filtered := cves[:0]
	for _, cve := range cves {
		if cve == nil {
			log.Warnf("Filtered out nil %s CVE", project)
			continue
		}
		// Note: it's ok if PublishedDate and LastModifiedDate are nil.
		if cve.CVE == nil || cve.Impact == nil || cve.Configurations == nil {
			log.Warnf("Filtered out %s CVE due to missing field: %v", project, *cve)
			continue
		}
		if cve.CVE.CVEDataMeta == nil || cve.CVE.CVEDataMeta.ID == "" {
			log.Warnf("Filtered out %s CVE due to missing or invalid CVEDataMeta: %v", project, *cve)
			continue
		}
		// Note: We already assumed there are both CVSSv2 and CVSSv3 scores, so let's just enforce there is CVSSv3.
		if cve.Impact.BaseMetricV3 == nil {
			log.Warnf("Filtered out %s CVE %s due to missing CVSSv3 score", project, cve.CVE.CVEDataMeta.ID)
			continue
		}
		filtered = append(filtered, cve)
	}
	return filtered
}

func addMissingCVEs(project Project, cves []*schema.NVDCVEFeedJSON10DefCVEItem) []*schema.NVDCVEFeedJSON10DefCVEItem {
	cvesFromNVD := make(map[string]struct{})
	for _, cve := range cves {
		cvesFromNVD[cve.CVE.CVEDataMeta.ID] = struct{}{}
	}

	for _, cve := range cvesNotInDataFeed[project] {
		if _, ok := cvesFromNVD[cve.CVE.CVEDataMeta.ID]; !ok {
			cves = append(cves, cve)
		}
	}
	return cves
}

func tryUpdateCVEWithAffectedVersions(cves []*schema.NVDCVEFeedJSON10DefCVEItem) {
	for _, cve := range cves {
		if _, ok := cvesWithoutFixedBy[cve.CVE.CVEDataMeta.ID]; !ok {
			continue
		}

		updateRequired := true
		filteredNodes := make([]*schema.NVDCVEFeedJSON10DefNode, 0, len(cve.Configurations.Nodes))
		for _, node := range cve.Configurations.Nodes {
			filteredCPE := make([]*schema.NVDCVEFeedJSON10DefCPEMatch, 0, len(node.CPEMatch))
			for _, cpeMatch := range node.CPEMatch {
				hasVersionInfo, err := cpeHasVersionInfo(cpeMatch)
				if err != nil {
					log.Error(err)
					continue
				}

				if !hasVersionInfo {
					continue
				}

				updateRequired = false
				// We filter because we want to get rid of any CPE not providing applicability statement.
				filteredCPE = append(filteredCPE, cpeMatch)
			}

			if len(filteredCPE) > 0 {
				filteredNodes = append(filteredNodes, &schema.NVDCVEFeedJSON10DefNode{
					CPEMatch: filteredCPE,
					Operator: node.Operator,
					Children: node.Children,
					Negate:   node.Negate,
				})
			}
		}

		// For the concerned CVEs, if NVD is providing data, use it, else update CVE.
		cve.Configurations.Nodes = filteredNodes
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
	if version == NotApplicable {
		return false, nil
	}

	// If CPE string matches any version check if ranges are provided. Note: CPE23URI provides affected versions and not ranges.
	if version == Any && update == Any {
		return cpeMatch.VersionStartIncluding != "" || cpeMatch.VersionEndIncluding != "" || cpeMatch.VersionEndExcluding != "", nil
	}

	return true, nil
}

// populateCVSSv2IfMissing populates missing CVSSv2 field, if necessary.
// This is required because it was previously assumed BOTH CVSSv2 and CVSSv3 are populated;
// however, this is no longer true. See https://nvd.nist.gov/general/news/retire-cvss-v2.
func populateCVSSv2IfMissing(cves []*schema.NVDCVEFeedJSON10DefCVEItem) {
	for _, cve := range cves {
		// Note: We know Impact is not nil, as this CVE would have been filtered out earlier.
		if cve.Impact.BaseMetricV2 == nil {
			// Populate with a dummy CVSSv2 score.
			cve.Impact.BaseMetricV2 = &schema.NVDCVEFeedJSON10DefImpactBaseMetricV2{
				CVSSV2: &schema.CVSSV20{
					AccessComplexity:      "LOW",
					AccessVector:          "LOCAL",
					Authentication:        "NONE",
					AvailabilityImpact:    "NONE",
					BaseScore:             0.0,
					ConfidentialityImpact: "NONE",
					IntegrityImpact:       "NONE",
					VectorString:          "AV:L/AC:L/Au:N/C:N/I:N/A:N",
					Version:               "2.0",
				},
				ExploitabilityScore: 3.9,
				ImpactScore:         0.0,
				Severity:            "LOW",
			}
		}
	}
}
