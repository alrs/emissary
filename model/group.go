package model

import (
	"github.com/benpate/data/journal"
	"github.com/benpate/rosetta/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group struct {
	GroupID primitive.ObjectID `path:"groupId" json:"groupId" bson:"_id"`
	Label   string             `path:"label"   json:"label"   bson:"label"`

	journal.Journal `json:"journal" bson:"journal"`
}

func NewGroup() Group {
	return Group{
		GroupID: primitive.NewObjectID(),
	}
}

func GroupSchema() schema.Element {
	return schema.Object{
		Properties: map[string]schema.Element{
			"groupId": schema.String{Format: "objectId"},
			"label":   schema.String{MaxLength: 50},
		},
	}
}

/*******************************************
 * data.Object Interface
 *******************************************/

func (group *Group) ID() string {
	return group.GroupID.Hex()
}
