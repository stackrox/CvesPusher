package nvd

import "encoding/json"

type CVEs struct {
	Entries []CVEEntry `json:"CVE_Items"`
}

type CVEEntry struct {
	CVE                  CVE           `json:"cve"`
	Impact               Impact        `json:"impact"`
	PublishedDateTime    string        `json:"publishedDate"`
	LastModifiedDateTime string        `json:"publishedDate"`
	Configurations       Configuration `json:"configurations"`
}

type CVE struct {
	Metadata     CVEMetadata    `json:"CVE_data_meta"`
	References   ReferenceData  `json:"references"`
	Descriptions CVEDescription `json:"description"`
}

type CVEDescription struct {
	Description []DescriptionData `json:"description_data"`
}

type DescriptionData struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type Configuration struct {
	Operator string `json:"operator"`
	Nodes    []Node `json:"nodes"`
}

type Node struct {
	CPEMatch []CPEMatch `json:"cpe_match"`
}

type CPEMatch struct {
	Vulnerable            bool   `json:"vulnerable"`
	CPE23Uri              string `json:"cpe23Uri"`
	VersionStartIncluding string `json:"versionStartIncluding"`
	VersionEndIncluding   string `json:"versionEndIncluding"`
	VersionEndExcluding   string `json:"versionEndExcluding"`
}

type CVEMetadata struct {
	CVEID string `json:"ID"`
}

type ReferenceData struct {
	Reference []CVEReference `json:"reference_data"`
}

type CVEReference struct {
	URL string `json:"url"`
}

type Impact struct {
	BaseMetricV2 BaseMetricV2 `json:"baseMetricV2"`
	BaseMetricV3 BaseMetricV3 `json:"baseMetricV3"`
}

type BaseMetricV2 struct {
	CVSSv2              CVSSv2  `json:"cvssV2"`
	Severity            string  `json:"severity"`
	ExploitabilityScore float64 `json:"exploitabilityScore"`
	ImpactScore         float64 `json:"impactScore"`
}

type CVSSv2 struct {
	Score            float64 `json:"baseScore"`
	AccessVector     string  `json:"accessVector"`
	AccessComplexity string  `json:"accessComplexity"`
	Authentication   string  `json:"authentication"`
	ConfImpact       string  `json:"confidentialityImpact"`
	IntegImpact      string  `json:"integrityImpact"`
	AvailImpact      string  `json:"availabilityImpact"`
}

type BaseMetricV3 struct {
	CVSSv3              CVSSv3  `json:"cvssV3"`
	ExploitabilityScore float64 `json:"exploitabilityScore"`
	ImpactScore         float64 `json:"impactScore"`
}

type CVSSv3 struct {
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

func Load(data []byte) (*CVEs, error) {
	var cves CVEs
	if err := json.Unmarshal(data, &cves); err != nil {
		return nil, err
	}
	return &cves, nil
}
