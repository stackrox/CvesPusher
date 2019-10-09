package utils

import (
	"encoding/json"
	"fmt"
)

type NvdCVEs struct {
	Entries []NvdCVEEntry `json:"CVE_Items"`
}

type NvdCVEEntry struct {
	CVE               NvdCVE    `json:"cve"`
	Impact            nvdImpact `json:"impact"`
	PublishedDateTime string    `json:"publishedDate"`
}

type NvdCVE struct {
	Metadata nvdCVEMetadata `json:"CVE_data_meta"`
	Affects  affects        `json:"affects"`
}

type affects struct {
	Vendor vendor `json:"vendor"`
}

type vendor struct {
	VendorData []vendorData `json:"vendor_data"`
}

type vendorData struct {
	VendorName string  `json:"vendor_name"`
	Product    product `json:"product"`
}

type product struct {
	ProductData []productData `json:"product_data"`
}

type productData struct {
	ProductName    string         `json:"product_name"`
	ProductVersion productVersion `json:"version"`
}

type productVersion struct {
	ProductVersionData []productVersionData `json:"version_data"`
}

type productVersionData struct {
	ProductVersionDataValue    string `json:"version_value"`
	ProductVersionDataAffected string `json:"version_affected"`
}

type nvdCVEMetadata struct {
	CVEID string `json:"ID"`
}

type nvdImpact struct {
	BaseMetricV2 nvdBaseMetricV2 `json:"baseMetricV2"`
	BaseMetricV3 nvdBaseMetricV3 `json:"baseMetricV3"`
}

type nvdBaseMetricV2 struct {
	CVSSv2 nvdCVSSv2 `json:"cvssV2"`
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
	err := json.Unmarshal(raw, &cves)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &cves, nil
}
