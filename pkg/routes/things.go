package routes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ppacher/mqtt-home/controller/pkg/middleware/render"
	"github.com/ppacher/mqtt-home/controller/pkg/registry"
	"github.com/ppacher/mqtt-home/controller/pkg/spec"
	"gopkg.in/macaron.v1"
)

// getAllThings handles a `GET /api/v1/things` request and returns all
// things registered
func getAllThings(ctx context.Context, m *macaron.Context, store registry.Registry) interface{} {
	things, err := store.All(ctx)
	if err != nil {
		return err
	}

	var models []*thingModel

	proto := "http://"
	if m.Req.TLS != nil {
		proto = "https://"
	}

	baseURL := fmt.Sprintf("%s%s/api/v1/things", proto, m.Req.Host)
	for _, t := range things {
		model, err := getThingModel(baseURL, t)
		if err != nil {
			return err
		}

		models = append(models, model)
	}

	return models
}

// createThing handles a `POST /api/v1/things` request and creates a new thing
func createThing(ctx context.Context, m *macaron.Context, thing spec.Thing, store registry.Registry) (int, interface{}) {
	// Make sure we have default values for all important fields
	thing.ApplyDefaults()

	if err := spec.ValidateThing(&thing); err != nil {
		return render.Unspecified, err
	}

	if err := store.Create(ctx, &thing); err != nil {
		return render.Unspecified, err
	}

	return http.StatusCreated, thing
}
