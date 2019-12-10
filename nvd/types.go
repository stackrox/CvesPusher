package nvd

import (
	"bytes"
	"encoding/json"
	"github.com/stackrox/nvdtools/cvefeed/nvd/schema"
	"io"
)

func Load(data []byte) (*schema.NVDCVEFeedJSON10, error) {
	return LoadReader(bytes.NewReader(data))
}

func LoadReader(stream io.Reader) (*schema.NVDCVEFeedJSON10, error) {
	//var cves CVEs
	var cveFeed schema.NVDCVEFeedJSON10
	if err := json.NewDecoder(stream).Decode(&cveFeed); err != nil {
		return nil, err
	}
	return &cveFeed, nil
}
