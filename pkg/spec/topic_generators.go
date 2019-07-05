package spec

import (
	"bytes"
	"text/template"
)

// TopicFromTemplate crafts an MQTT topic from a given template string and context related information.
// The provided thing and item will be accessible via .Thing and .Item as well as the aliases .thing and .item
// Any additional maps will be merged into the template context
//
// Example:
//
//		myThing := spec.Thing{ID: "myThing"}
//		myItem := spec.Item{ID: "myItem"}
//		topic, err := TopicFromTemplate("{{.thing.ID}}/set/{{.item.ID}}", &myThing, &myItem)
//		// topic == "myThing/set/myItem"
//
func TopicFromTemplate(tmpl string, thing *Thing, property *Property, extra ...map[string]interface{}) (string, error) {
	t, err := template.New(tmpl).Parse(tmpl)
	if err != nil {
		return "", err
	}

	ctx := make(map[string]interface{})

	if thing != nil {
		ctx["Thing"] = thing
		ctx["thing"] = thing
	}

	if property != nil {
		ctx["Item"] = property
		ctx["item"] = property
	}

	for _, m := range extra {
		if m == nil {
			continue
		}

		for k, v := range m {
			ctx[k] = v
		}
	}

	var res bytes.Buffer
	if err := t.Execute(&res, ctx); err != nil {
		return "", err
	}

	return res.String(), nil
}
