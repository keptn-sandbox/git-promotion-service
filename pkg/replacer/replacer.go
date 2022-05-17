package replacer

import (
	logger "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

const prefix = `{"keptn.git-promotion.replacewith":"`
const suffix = `"}`

// Replace value marked by yaml comment e.g.
// tag: 2.5.5 # {"keptn.git-promotion.replacewith":"data.image.tag"}
func Replace(fileData string, tags map[string]string) (result string) {
	replaced := fileData
	//quick check for faster processing
	for k, v := range tags {
		if strings.Contains(replaced, prefix+k+suffix) {
			replaced = replaceValue(replaced, k, v)
		}
	}
	logger.WithField("func", "Replace").Infof("tags: %v, original: %s, replaced: %s", tags, fileData, replaced)
	return replaced
}

func replaceValue(file, key, value string) string {
	splitted := strings.Split(file, "\n")
	annotation := prefix + key + suffix
	re := regexp.MustCompile(`(^.+: ).*( # ` + annotation + `$)`)
	for i, s := range splitted {
		if strings.Contains(s, annotation) {
			splitted[i] = re.ReplaceAllString(s, "${1}"+value+"${2}")
		}
	}
	return strings.Join(splitted, "\n")
}
