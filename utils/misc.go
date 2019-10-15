package utils

import "fmt"

func GetDataFeedURL(feedName string) string {
	return fmt.Sprintf("https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-%s.json.gz", feedName)
}

func GetDataFeedMetaURL(feedName string) string {
	return fmt.Sprintf("https://nvd.nist.gov/feeds/json/cve/1.1/nvdcve-1.1-%s.meta", feedName)
}
