package ship_mapped

import (
	"testing"

	"github.com/darklab8/fl-configs/configs/configs_mapped/parserutils/filefind/file"
	"github.com/darklab8/fl-configs/configs/configs_mapped/parserutils/iniload"
	"github.com/darklab8/go-utils/goutils/utils"
	"github.com/darklab8/go-utils/goutils/utils/utils_filepath"
	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	test_directory := utils.GetCurrrentTestFolder()
	fileref1 := file.NewFile(utils_filepath.Join(test_directory, "shiparch.ini"))
	fileref2 := file.NewFile(utils_filepath.Join(test_directory, "rtc_shiparch.ini"))

	config := Read([]*iniload.IniLoader{iniload.NewLoader(fileref1).Scan(), iniload.NewLoader(fileref2).Scan()})

	assert.Greater(t, len(config.Ships), 0)

	for _, commodity := range config.Ships {
		commodity.IdsName.Get()
	}
}