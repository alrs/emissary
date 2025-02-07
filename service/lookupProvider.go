package service

import (
	"github.com/EmissarySocial/emissary/model"
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

	case "block-types":
		return form.NewReadOnlyLookupGroup(
			form.LookupCode{Label: "Block a Domain", Value: model.BlockTypeDomain},
			form.LookupCode{Label: "Block a Person", Value: model.BlockTypeActor},
			form.LookupCode{Label: "Block Tags & Keywords", Value: model.BlockTypeContent},
		)

	case "folders":
		return NewFolderLookupProvider(service.folderService, service.userID)

	case "folder-icons":
		return form.NewReadOnlyLookupGroup(dataset.Icons()...)

	case "groups":
		return NewGroupLookupProvider(service.groupService)

	case "reaction-icons":
		return form.NewReadOnlyLookupGroup(
			form.LookupCode{Label: "Love", Group: "Like", Value: "❤️"},
			form.LookupCode{Label: "Like", Group: "Like", Value: "👍"},
			form.LookupCode{Label: "Dislike", Group: "Dislike", Value: "👎"},
			form.LookupCode{Label: "Smile", Group: "Like", Value: "😀"},
			form.LookupCode{Label: "Laugh", Group: "Like", Value: "🤣"},
			form.LookupCode{Label: "Frown", Group: "Dislike", Value: "🙁"},
			form.LookupCode{Label: "Emphasize", Group: "Like", Value: "‼️", Icon: ""},
			form.LookupCode{Label: "Celebrate", Group: "Like", Value: "🎉"},
			form.LookupCode{Label: "Question", Group: "Like", Value: "❓"},
			form.LookupCode{Label: "Crown", Group: "Like", Value: "👑"},
			form.LookupCode{Label: "Fire", Group: "Like", Value: "🔥"},
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
