package handler

import (
	"github.com/EmissarySocial/emissary/domain"
	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/hannibal/streams"
	"github.com/benpate/hannibal/vocab"
	"github.com/davecgh/go-spew/spew"
)

func init() {
	inboxRouter.Add(vocab.ActivityTypeUndo, vocab.ActivityTypeLike, func(factory *domain.Factory, user *model.User, document streams.Document) error {

		spew.Dump("UndoLike", document.Value())

		// Hooo-dat?!?!?
		return nil
	})
}
