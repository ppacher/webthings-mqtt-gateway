package routes

import (
	"fmt"
	"regexp"

	"github.com/own-home/central/pkg/spec"
)

var schemeRe = regexp.MustCompile("^[a-z]+://.*$")

type linkObject struct {
	Href      string `json:"href"`
	Rel       string `json:"rel,omitempty"`
	MediaType string `json:"mediaType,omitempty"`
}

type thingModel struct {
	*spec.Thing

	// Icon is moved into the links section
	Icon string `json:"-"`

	// Links holds additional links
	Links []linkObject `json:"links,omitempty"`

	Properties map[string]*propertyModel `json:"properties"`
}

type propertyModel struct {
	*spec.Property

	// Links holds additional links
	Links []linkObject `json:"links,omitempty"`
}

func getThingModel(baseUrl string, thing *spec.Thing) (*thingModel, error) {
	copy := *thing
	if !schemeRe.MatchString(thing.ID) {
		// we assume that this is not a valid URI so make one
		if copy.ID[0] != '/' {
			copy.ID = "/" + copy.ID
		}
		copy.ID = baseUrl + copy.ID
	} else {
		fmt.Println(thing.ID + " matches")
	}

	// TODO(ppacher): onyl add events/actions if there are some
	links := []linkObject{
		{
			Rel:  "properties",
			Href: copy.ID + "/properties",
		},
		{
			Rel:  "events",
			Href: copy.ID + "/events",
		},
		{
			Rel:  "actions",
			Href: copy.ID + "/actions",
		},
	}

	if thing.Icon != "" {
		links = append(links, linkObject{
			Rel:  "icon",
			Href: thing.Icon,
		})
	}

	properties := make(map[string]*propertyModel)
	for key, value := range copy.Properties {
		p, err := getPropertyModel(copy.ID, &copy, value)
		if err != nil {
			return nil, err
		}
		properties[key] = p
	}

	return &thingModel{
		Thing:      &copy,
		Links:      links,
		Properties: properties,
	}, nil
}

func getPropertyModel(baseURL string, thing *spec.Thing, prop *spec.Property) (*propertyModel, error) {
	if baseURL[len(baseURL)-1] != '/' {
		baseURL = baseURL + "/"
	}

	baseURL = baseURL + "properties/" + prop.ID
	var links []linkObject

	if spec.IsJSONEncodableValue(prop.Type) {
		// if it is a normal primitive value we support JSON and
		// history
		links = append(links, []linkObject{
			{
				Href:      baseURL,
				MediaType: "application/json",
			},
			{
				Href:      baseURL + "/history",
				Rel:       "history",
				MediaType: "application/json",
			},
		}...)
	} else {
		links = append(links, []linkObject{
			{
				Href:      baseURL,
				MediaType: string(prop.Type),
			},
		}...)
	}

	// TODO(ppacher): add links
	return &propertyModel{
		Property: prop,
		Links:    links,
	}, nil
}
