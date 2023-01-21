package model

import (
	"math"
	"time"

	"github.com/EmissarySocial/emissary/tools/id"
	"github.com/benpate/data/journal"
	"github.com/benpate/rosetta/mapof"
	"github.com/benpate/rosetta/maps"
	"github.com/benpate/rosetta/sliceof"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Stream corresponds to a top-level path on any Domain.
type Stream struct {
	StreamID        primitive.ObjectID           `json:"streamId"            bson:"_id"`                 // Unique identifier of this Stream.  (NOT USED PUBLICLY)
	ParentID        primitive.ObjectID           `json:"parentId"            bson:"parentId"`            // Unique identifier of the "parent" stream. (NOT USED PUBLICLY)
	Token           string                       `json:"token"               bson:"token"`               // Unique value that identifies this element in the URL
	TopLevelID      string                       `json:"topLevelId"          bson:"topLevelId"`          // Unique identifier of the "top-level" stream. (NOT USED PUBLICLY)
	TemplateID      string                       `json:"templateId"          bson:"templateId"`          // Unique identifier (name) of the Template to use when rendering this Stream in HTML.
	StateID         string                       `json:"stateId"             bson:"stateId"`             // Unique identifier of the State this Stream is in.  This is used to populate the State information from the Template service at load time.
	Permissions     mapof.Object[sliceof.String] `json:"permissions"         bson:"permissions"`         // Permissions for which users can access this stream.
	DefaultAllow    id.Slice                     `json:"defaultAllow"        bson:"defaultAllow"`        // List of Groups that are allowed to perform the 'default' (view) action.  This is used to query general access to the Stream from the database, before performing server-based authentication.
	Document        DocumentLink                 `json:"document"            bson:"document"`            // Summary content of this document
	InReplyTo       DocumentLink                 `json:"inReplyTo,omitempty" bson:"inReplyTo,omitempty"` // If this stream is a reply to another stream or web page, then this links to the original document.
	Origin          OriginLink                   `json:"origin,omitempty"    bson:"origin,omitempty"`    // If this stream is imported from an external service, this is a link to the original document
	Content         Content                      `json:"content"             bson:"content,omitempty"`   // Content objects for this stream.
	Data            maps.Map                     `json:"data"                bson:"data,omitempty"`      // Set of data to populate into the Template.  This is validated by the JSON-Schema of the Template.
	Rank            int                          `json:"rank"                bson:"rank"`                // If Template uses a custom sort order, then this is the value used to determine the position of this Stream.
	AsFeature       bool                         `json:"asFeature"           bson:"asFeature"`           // If TRUE, then this stream is a "feature" that is meant to be embedded into other stream views.
	PublishDate     int64                        `json:"publishDate"         bson:"publishDate"`         // Unix timestamp of the date/time when this document is/was/will be first available on the domain.
	UnPublishDate   int64                        `json:"unpublishDate"       bson:"unpublishDate"`       // Unix timestemp of the date/time when this document will no longer be available on the domain.
	journal.Journal `json:"journal" bson:"journal"`
}

// NewStream returns a fully initialized Stream object.
func NewStream() Stream {

	streamID := primitive.NewObjectID()

	return Stream{
		StreamID:      streamID,
		Token:         streamID.Hex(),
		ParentID:      primitive.NilObjectID,
		StateID:       "new",
		Permissions:   NewStreamPermissions(),
		Data:          make(maps.Map),
		PublishDate:   math.MaxInt64,
		UnPublishDate: math.MaxInt64,
	}
}

// NewStreamPermissions returns a fully initialized Permissions object
func NewStreamPermissions() mapof.Object[sliceof.String] {
	return make(mapof.Object[sliceof.String])
}

/******************************************
 * data.Object Interface
 ******************************************/

// ID returns the primary key of this object
func (stream *Stream) ID() string {
	return stream.StreamID.Hex()
}

/******************************************
 * Other Data Accessors
 ******************************************/

// GetSort returns the sortable value for this stream, based onthe provided fieldName
func (stream *Stream) GetSort(fieldName string) any {
	switch fieldName {
	case "publishDate":
		return stream.PublishDate
	case "document.label":
		return stream.Document.Label
	case "rank":
		return stream.Rank
	default:
		return 0
	}
}

// Links returns all resources linked to this Stream.  Some links may be empty.
func (stream *Stream) Links() []Link {

	result := make([]Link, 0, 2)

	if !stream.Document.Author.IsEmpty() {
		result = append(result, stream.Document.AuthorLink())
	}

	if !stream.InReplyTo.IsEmpty() {
		result = append(result, stream.InReplyTo.Link(LinkRelationInReplyTo))
	}

	return result
}

// SetAuthor populates the `Author` link of this `Stream`.
func (stream *Stream) SetAuthor(user *User) {
	stream.Document.Author = user.PersonLink()
}

// OutboxItem generates a new Stream that will sit in the author's Outbox
func (stream *Stream) OutboxItem() Stream {
	result := NewStream()
	result.Document = stream.Document
	result.TopLevelID = "outbox"
	result.TemplateID = "outbox-item"
	result.ParentID = stream.Document.Author.InternalID

	return result
}

/******************************************
 * RoleStateEnumerator Methods
 ******************************************/

// State returns the current state of this Stream.  It is
// part of the implementation of the RoleStateEmulator interface
func (stream *Stream) State() string {
	return stream.StateID
}

// Roles returns a list of all roles that match the provided authorization
func (stream *Stream) Roles(authorization *Authorization) []string {

	// Everyone has "anonymous" access
	result := []string{MagicRoleAnonymous}

	if authorization == nil {
		return result
	}

	// Owners are hard-coded to do everything, so no other roles need to be returned.
	if authorization.DomainOwner {
		return []string{MagicRoleOwner}
	}

	if authorization.IsAuthenticated() {
		result = append(result, MagicRoleAuthenticated)
	}

	// Authors sometimes have special permissions, too.
	if !stream.Document.Author.InternalID.IsZero() {
		if authorization.UserID == stream.Document.Author.InternalID {
			result = append(result, MagicRoleAuthor)
		}
	}

	// Otherwise, append all roles matched from the permissions
	result = append(result, stream.PermissionRoles(authorization.AllGroupIDs()...)...)

	return result
}

// DefaultAllowAnonymous returns TRUE if a Stream's default action (VIEW)
// is visible to anonymous visitors
func (stream *Stream) DefaultAllowAnonymous() bool {
	for index := range stream.DefaultAllow {
		if stream.DefaultAllow[index] == MagicGroupIDAnonymous {
			return true
		}
	}
	return false
}

/******************************************
 * Permission Methods
 ******************************************/

// AssignPermissions assigns a role to a group
func (stream *Stream) AssignPermission(role string, groupID primitive.ObjectID) {
	groupIDHex := groupID.Hex()

	if _, ok := stream.Permissions[groupIDHex]; !ok {
		stream.Permissions[groupIDHex] = []string{role}
		return
	}

	stream.Permissions[groupIDHex] = append(stream.Permissions[groupIDHex], role)
}

// PermissionGroups returns all groups that match the provided roles
func (stream *Stream) PermissionGroups(roles ...string) []primitive.ObjectID {

	result := make([]primitive.ObjectID, 0)

	for _, role := range roles {
		switch role {

		case "anonymous":
			result = append(result, MagicGroupIDAnonymous)

		case "authenticated":
			result = append(result, MagicGroupIDAuthenticated)

		}
	}

	for groupID, groupRoles := range stream.Permissions {
		if matchAny(roles, groupRoles) {
			if groupID, err := primitive.ObjectIDFromHex(groupID); err == nil {
				result = append(result, groupID)
			}
		}
	}

	return result
}

// PermissionRoles returns a unique list of all roles that the provided groups can access.
func (stream *Stream) PermissionRoles(groupIDs ...primitive.ObjectID) []string {

	result := []string{}

	// Copy values from group roles
	for _, groupID := range groupIDs {
		if roles, ok := stream.Permissions[groupID.Hex()]; ok {
			result = append(result, roles...)
		}
	}

	return result
}

// SimplePermissionModel returns a model object for displaying Simple Sharing.
func (stream *Stream) SimplePermissionModel() maps.Map {

	// Special case if this is for EVERYBODY
	if _, ok := stream.Permissions[MagicGroupIDAnonymous.Hex()]; ok {
		return maps.Map{
			"rule":     "anonymous",
			"groupIds": []string{},
		}
	}

	// Special case if this is for AUTHENTICATED
	if _, ok := stream.Permissions[MagicGroupIDAuthenticated.Hex()]; ok {
		return maps.Map{
			"rule":     "authenticated",
			"groupIds": []string{},
		}
	}

	// Fall through means that additional groups are selected.
	// First, get all keys to the Groups map
	groupIDs := make([]string, len(stream.Permissions))
	index := 0

	for groupID := range stream.Permissions {
		groupIDs[index] = groupID
		index++
	}

	return maps.Map{
		"rule":     "private",
		"groupIds": groupIDs,
	}
}

/******************************************
 * Other Methods
 ******************************************/

// HasParent returns TRUE if this Stream has a valid parentID
func (stream *Stream) HasParent() bool {
	return !stream.ParentID.IsZero()
}

// NewAttachment creates a new file Attachment linked to this Stream.
func (stream *Stream) NewAttachment(filename string) Attachment {
	result := NewAttachment(AttachmentTypeStream, stream.StreamID)
	result.Original = filename

	return result
}

func (stream *Stream) IsPublished() bool {

	now := time.Now().Unix()
	return (stream.PublishDate <= now) && (stream.UnPublishDate > now)
}
