package routes

import (
	"context"

	"github.com/go-macaron/binding"
	"github.com/own-home/central/pkg/errors"
	"github.com/own-home/central/pkg/spec"
	"gopkg.in/macaron.v1"
)

type ThingID string
type PropertyID string

// Install will install all routes exposed via the REST API
func Install(m *macaron.Macaron) error {
	// directly expose the context via macarons dependecy injection
	m.Use(func(ctx *macaron.Context) {
		ctx.MapTo(ctx.Req.Context(), (*context.Context)(nil))
	})

	m.Get("/", func(ctx *macaron.Context) {
		ctx.JSON(200, map[string]interface{}{
			"version": "v1",
		})
	})

	thingRequestBinding := binding.Bind(spec.Thing{})

	thingID := func(ctx *macaron.Context) {
		thingID := ctx.Params("thingID")
		if thingID == "" {
			ctx.JSON(400, errors.NewWithStatus(400, "invalid thing ID"))
			return
		}

		ctx.Map(ThingID(thingID))
	}
	propID := func(ctx *macaron.Context) {
		propID := ctx.Params("propID")
		if propID == "" {
			ctx.JSON(400, errors.NewWithStatus(400, "invalid property ID"))
			return
		}

		ctx.Map(PropertyID(propID))
	}

	// /api/v1
	m.Group("/api/v1", func() {

		// /api/v1/things
		m.Group("/things", func() {

			m.Get("", getAllThings)
			m.Post("", thingRequestBinding, createThing)

			// /api/v1/things/{thingID}
			m.Group("/:thingID", func() {
				m.Get("", getThing)
				m.Put("", thingRequestBinding, updateThing)
				m.Delete("", deleteThing)

				// /api/v1/things/{thingID}/properties
				m.Group("/properties", func() {
					m.Get("", getProperties)

					// /api/v1/things/{thingID}/properties/{propID}
					m.Group("/:propID", func() {
						m.Get("", getProperty)
						m.Get("/history", getValues)
						m.Post("", setProperty)
					}, propID)
				})
			}, thingID)
		})
	})

	return nil
}
