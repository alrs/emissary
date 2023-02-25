package service

import (
	"github.com/EmissarySocial/emissary/tools/dataset"
	"github.com/benpate/form"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LookupProvider struct {
	themeService  *Theme
	groupService  *Group
	folderService *Folder
	userID        primitive.ObjectID
}

func NewLookupProvider(themeService *Theme, groupService *Group, folderService *Folder, userID primitive.ObjectID) LookupProvider {
	return LookupProvider{
		themeService:  themeService,
		groupService:  groupService,
		folderService: folderService,
		userID:        userID,
	}
}

func (service LookupProvider) Group(path string) form.LookupGroup {

	switch path {

	case "folders":
		return NewFolderLookupProvider(service.folderService, service.userID)

	case "folder-icons":
		return form.NewReadOnlyLookupGroup(dataset.Icons()...)

	case "groups":
		return NewGroupLookupProvider(service.groupService)

	case "purgeDurations":
		return form.NewReadOnlyLookupGroup(
			form.LookupCode{Label: "1 Day", Value: "1"},
			form.LookupCode{Label: "1 Week", Value: "7"},
			form.LookupCode{Label: "1 Month", Value: "31"},
			form.LookupCode{Label: "1 Year", Value: "365"},
			form.LookupCode{Label: "Forever", Value: "0"},
		)

	case "sharing":
		return form.NewReadOnlyLookupGroup(
			form.LookupCode{Value: "anonymous", Label: "Everyone (including anonymous visitors)"},
			form.LookupCode{Value: "authenticated", Label: "Authenticated People Only"},
			form.LookupCode{Value: "private", Label: "Only Selected Groups"},
		)

	case "themes":
		return NewThemeLookupProvider(service.themeService)

	default:
		return form.NewReadOnlyLookupGroup()
	}
}
