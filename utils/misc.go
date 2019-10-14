package utils

import "fmt"

func GetDataFeedURL(feedName string) string {
	return fmt.Sprintf("https://nvd.nist.gov/feeds/json/cve/1.0/nvdcve-1.0-%s.json.gz", feedName)
}

func GetDataFeedMetaURL(feedName string) string {
	return fmt.Sprintf("https://nvd.nist.gov/feeds/json/cve/1.0/nvdcve-1.0-%s.meta", feedName)
}
