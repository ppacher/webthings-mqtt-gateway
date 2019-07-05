package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ppacher/mqtt-home/controller/pkg/errors"
	"github.com/ppacher/mqtt-home/controller/pkg/registry"
	"github.com/ppacher/mqtt-home/controller/pkg/spec"
	"gopkg.in/macaron.v1"
)

// getThing handles `GET /api/v1/things/:thingID` and returns the thing
func getThing(ctx context.Context, m *macaron.Context, thingID ThingID, store registry.Registry) interface{} {
	thing, err := store.Get(ctx, string(thingID))
	if err != nil {
		return err
	}

	proto := "http://"
	if m.Req.TLS != nil {
		proto = "https://"
	}

	baseURL := fmt.Sprintf("%s%s/api/v1/things", proto, m.Req.Host)
	model, err := getThingModel(baseURL, thing)
	if err != nil {
		return err
	}

	return model
}

// updateThing handles `PUT /api/v1/things/:thingID` and updates the thing
func updateThing(ctx context.Context, thingID ThingID, updated spec.Thing, store registry.Registry) interface{} {
	if err := spec.ValidateThing(&updated); err != nil {
		return err
	}

	// Seems like the user want's to change the thingID, check that we don't collide
	// with an existing one
	if string(thingID) != updated.ID {
		// TODO(ppacher): support it
		return errors.NewWithStatus(http.StatusForbidden, "thing IDs cannot be changed")
	}

	_, err := store.Get(ctx, string(thingID))
	if err != nil {
		return err
	}

	if err := store.Update(ctx, &updated); err != nil {
		return err
	}

	return http.StatusNoContent
}

// deleteThing handles `DELETE /api/v1/things/:thingID` and returns the thing
func deleteThing(ctx context.Context, thingID ThingID, store registry.Registry) interface{} {
	if err := store.Delete(ctx, string(thingID)); err != nil {
		return err
	}

	return http.StatusAccepted
}
