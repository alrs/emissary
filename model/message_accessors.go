package model

import (
	"github.com/benpate/rosetta/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageSchema returns a JSON Schema that describes this object
func MessageSchema() schema.Element {
	return schema.Object{
		Properties: schema.ElementMap{
			"messageId":    schema.String{Format: "objectId"},
			"userId":       schema.String{Format: "objectId"},
			"folderId":     schema.String{Format: "objectId"},
			"socialRole":   schema.String{MaxLength: 64},
			"origin":       OriginLinkSchema(),
			"references":   schema.Array{Items: OriginLinkSchema()},
			"url":          schema.String{Format: "uri"},
			"label":        schema.String{MaxLength: 128},
			"summary":      schema.String{MaxLength: 1024},
			"imageUrl":     schema.String{Format: "uri"},
			"attributedTo": PersonLinkSchema(),
			"inReplyTo":    schema.String{Format: "uri"},
			"contentHtml":  schema.String{Format: "html"},
			"contentJson":  schema.String{Format: "json"},
			"responses":    ResponseSummarySchema(),
			"myResponse":   schema.String{Enum: []string{ResponseTypeLike, ResponseTypeDislike, ResponseTypeRepost}},
			"publishDate":  schema.Integer{BitSize: 64},
			"readDate":     schema.Integer{BitSize: 64},
			"rank":         schema.Integer{BitSize: 64},
		},
	}
}

/******************************************
 * Getter/Setter Methods
 ******************************************/

func (message *Message) GetPointer(name string) (any, bool) {
	switch name {

	case "socialRole":
		return &message.SocialRole, true

	case "origin":
		return &message.Origin, true

	case "references":
		return &message.References, true

	case "url":
		return &message.URL, true

	case "label":
		return &message.Label, true

	case "summary":
		return &message.Summary, true

	case "imageUrl":
		return &message.ImageURL, true

	case "attributedTo":
		return &message.AttributedTo, true

	case "inReplyTo":
		return &message.InReplyTo, true

	case "contentHtml":
		return &message.ContentHTML, true

	case "contentJson":
		return &message.ContentJSON, true

	case "responses":
		return &message.Responses, true

	case "myResponse":
		return &message.MyResponse, true

	case "publishDate":
		return &message.PublishDate, true

	case "readDate":
		return &message.ReadDate, true

	case "rank":
		return &message.Rank, true

	default:
		return nil, false
	}
}

func (message *Message) GetStringOK(name string) (string, bool) {

	switch name {

	case "messageId":
		return message.MessageID.Hex(), true

	case "userId":
		return message.UserID.Hex(), true

	case "folderId":
		return message.FolderID.Hex(), true

	}

	return "", false
}

/******************************************
 * Setter Interfaces
 ******************************************/

func (message *Message) SetString(name string, value string) bool {

	switch name {

	case "messageId":
		if objectID, err := primitive.ObjectIDFromHex(value); err == nil {
			message.MessageID = objectID
			return true
		}

	case "userId":
		if objectID, err := primitive.ObjectIDFromHex(value); err == nil {
			message.UserID = objectID
			return true
		}

	case "folderId":
		if objectID, err := primitive.ObjectIDFromHex(value); err == nil {
			message.FolderID = objectID
			return true
		}
	}

	return false
}
