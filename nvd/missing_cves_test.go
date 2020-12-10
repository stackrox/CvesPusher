package nvd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStaleCVEs(t *testing.T) {
	for _, projectMap := range cvesNotInDataFeed {
		for _, cve := range projectMap {
			assert.NotNil(t, cve)
			assert.NotNil(t, cve.CVE)
			assert.NotNil(t, cve.CVE.CVEDataMeta)

			if cve.Impact != nil {
				require.NotNil(t, cve.Impact.BaseMetricV2)
				assert.NotNil(t, cve.Impact.BaseMetricV2.CVSSV2)

				require.NotNil(t, cve.Impact.BaseMetricV3)
				assert.NotNil(t, cve.Impact.BaseMetricV3.CVSSV3)
			}
		}
	}
}
