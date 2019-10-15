package utils

import (
	"encoding/json"
)

type NvdCVEs struct {
	Entries []NvdCVEEntry `json:"CVE_Items"`
}

type NvdCVEEntry struct {
	CVE               NvdCVE        `json:"cve"`
	Impact            nvdImpact     `json:"impact"`
	PublishedDateTime string        `json:"publishedDate"`
	Configurations    configuration `json:"configurations"`
}

type NvdCVE struct {
	Metadata nvdCVEMetadata `json:"CVE_data_meta"`
}

type configuration struct {
	Operator string `json:"operator"`
	Nodes    []node `json:"nodes"`
}

type node struct {
	CPEMatch []cpeMatch `json:"cpe_match"`
}

type cpeMatch struct {
	Vulnerable            bool   `json:"vulnerable"`
	CPE23Uri              string `json:"cpe23Uri"`
	VersionStartIncluding string `json:"versionStartIncluding"`
	VersionEndIncluding   string `json:"versionEndIncluding"`
	VersionEndExcluding   string `json:"versionEndExcluding"`
}

type nvdCVEMetadata struct {
	CVEID string `json:"ID"`
}

type nvdImpact struct {
	BaseMetricV2 nvdBaseMetricV2 `json:"baseMetricV2"`
	BaseMetricV3 nvdBaseMetricV3 `json:"baseMetricV3"`
}

type nvdBaseMetricV2 struct {
	CVSSv2              nvdCVSSv2 `json:"cvssV2"`
	Severity            string    `json:"severity"`
	ExploitabilityScore float64   `json:"exploitabilityScore"`
	ImpactScore         float64   `json:"impactScore"`
}

type nvdCVSSv2 struct {
	Score            float64 `json:"baseScore"`
	AccessVector     string  `json:"accessVector"`
	AccessComplexity string  `json:"accessComplexity"`
	Authentication   string  `json:"authentication"`
	ConfImpact       string  `json:"confidentialityImpact"`
	IntegImpact      string  `json:"integrityImpact"`
	AvailImpact      string  `json:"availabilityImpact"`
}

type nvdBaseMetricV3 struct {
	CVSSv3              nvdCVSSv3 `json:"cvssV3"`
	ExploitabilityScore float64   `json:"exploitabilityScore"`
	ImpactScore         float64   `json:"impactScore"`
}

type nvdCVSSv3 struct {
	Score              float64 `json:"baseScore"`
	AttackVector       string  `json:"attackVector"`
	AttackComplexity   string  `json:"attackComplexity"`
	PrivilegesRequired string  `json:"privilegesRequired"`
	UserInteraction    string  `json:"userInteraction"`
	Scope              string  `json:"scope"`
	ConfImpact         string  `json:"confidentialityImpact"`
	IntegImpact        string  `json:"integrityImpact"`
	AvailImpact        string  `json:"availabilityImpact"`
}

func GetCVEs(raw []byte) (*NvdCVEs, error) {
	var cves NvdCVEs
	if err := json.Unmarshal(raw, &cves); err != nil {
		return nil, err
	}
	return &cves, nil
}
