package handler

import (
	"github.com/graphql-go/graphql"
)

type faction struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var rebels = &faction{
	"RmFjdGlvbjox",
	"Alliance to Restore the Republic",
}

var factionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Faction",
	Description: "A faction in the Star Wars saga",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.String,
			Description: "Id of the faction",
		},
		"name": &graphql.Field{
			Type:        graphql.String,
			Description: "The name of the faction.",
		},
	},
})

var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"rebels": &graphql.Field{
			Type: factionType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return rebels, nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query: queryType,
})
