package replacer

import (
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"log"
	"reflect"
)

func ConvertToMap(event cloudevents.Event) (res map[string]string) {
	temp := make(map[string]interface{})
	res = make(map[string]string)
	if err := event.DataAs(&temp); err != nil {
		log.Fatalf("marshall err %s", err)
	}
	addKeysToMap("data", &res, temp)
	addKeysToMap("", &res, event.Extensions())
	res["id"] = event.ID()
	res["source"] = event.Source()
	res["specversion"] = event.SpecVersion()
	return res
}

func addKeysToMap(root string, m *map[string]string, temp map[string]interface{}) {
	for k, v := range temp {
		key := k
		if root != "" {
			key = root + "." + k
		}
		if v != nil {
			if reflect.TypeOf(v).Kind() != reflect.Map {
				(*m)[key] = fmt.Sprintf("%v", v)
			} else {
				addKeysToMap(key, m, v.(map[string]interface{}))
			}
		}
	}
}
