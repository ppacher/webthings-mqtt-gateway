package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ppacher/mqtt-home/controller/pkg/control"
	"github.com/ppacher/mqtt-home/controller/pkg/errors"
	"github.com/ppacher/mqtt-home/controller/pkg/registry"
	"github.com/ppacher/mqtt-home/controller/pkg/spec"
	"github.com/sirupsen/logrus"
	"gopkg.in/macaron.v1"
)

// GetItems handles `GET /api/v1/things/:thingID/items` and returns the thing
func getProperties(ctx context.Context, thingID ThingID, store registry.Registry) interface{} {
	thing, err := store.Get(ctx, string(thingID))
	if err != nil {
		return err
	}

	values := make(map[string]interface{})

	for key := range thing.Properties {
		val, err := store.GetItemValue(ctx, string(thingID), key)
		if err != nil {
			logrus.Errorf("failed to get property value: %s", err.Error())
			continue
		}

		values[key] = val
	}

	return values
}

func getProperty(ctx context.Context, m *macaron.Context, thingID ThingID, propID PropertyID, store registry.Registry) interface{} {
	thing, err := store.Get(ctx, string(thingID))
	if err != nil {
		return err
	}

	prop, ok := thing.Properties[string(propID)]
	if !ok {
		return errors.NewWithStatus(404, "unknown property")
	}

	value, err := store.GetItemValue(ctx, string(thingID), string(propID))
	if err != nil {
		return err
	}

	if !spec.IsJSONEncodableValue(prop.Type) {
		switch v := value.(type) {
		case string:
			m.RawData(200, []byte(v))
		case []byte:
			m.RawData(200, v)
		default:
			return v
		}
		return nil
	}

	return map[string]interface{}{
		string(propID): value,
	}
}

func getValues(ctx context.Context, thingID ThingID, propID PropertyID, store registry.Registry) interface{} {
	value, err := store.GetItemValue(ctx, string(thingID), string(propID))
	if err != nil {
		return err
	}

	return value
}

func setProperty(ctx context.Context, m *macaron.Context, thingID ThingID, propID PropertyID, control *control.MissionControl) interface{} {
	var x interface{}
	defer m.Req.Request.Body.Close()
	err := json.NewDecoder(m.Req.Request.Body).Decode(&x)
	if err != nil {
		return errors.WrapWithStatus(400, err)
	}

	if err := control.SetItem(ctx, string(thingID), string(propID), x); err != nil {
		return err
	}

	return http.StatusAccepted
}
