package convert

import (
	"time"

	"github.com/EmissarySocial/emissary/model"
	"github.com/EmissarySocial/emissary/tools/iterators"
	"github.com/benpate/data"
	"github.com/kr/jsonfeed"
)

func IteratorToJSonFeed(url string, title string, description string, iterator data.Iterator) jsonfeed.Feed {

	return jsonfeed.Feed{
		Version:     "https://jsonfeed.org/version/1.1",
		Title:       title,
		Description: description,
		HomePageURL: url,
		FeedURL:     url + "/feed?format=json",
		Hubs: []jsonfeed.Hub{{
			Type: "WebSub",
			URL:  url + "/websub",
		}},
		Items: iterators.Map(iterator, model.NewStream, StreamToJsonFeed),
	}
}

func StreamToJsonFeed(stream model.Stream) jsonfeed.Item {

	return jsonfeed.Item{
		ID:            stream.Token,
		URL:           stream.Document.URL,
		Title:         stream.Document.Label,
		ContentHTML:   stream.Content.HTML,
		Summary:       stream.Document.Summary,
		Image:         stream.Document.ImageURL,
		DatePublished: time.UnixMilli(stream.PublishDate),
		DateModified:  time.UnixMilli(stream.UpdateDate),
		Author: &jsonfeed.Author{
			Name:   stream.Document.Author.Name,
			URL:    stream.Document.Author.ProfileURL,
			Avatar: stream.Document.Author.ImageURL,
		},
		// TODO: Attachments for podcasts, etc.
	}
}

func JsonFeedToActivity(item jsonfeed.Item) model.Activity {

	activity := model.NewActivity()
	activity.Document = model.DocumentLink{
		URL:         item.URL,
		Label:       item.Title,
		Summary:     item.Summary,
		ImageURL:    item.Image,
		PublishDate: item.DatePublished.UnixMilli(),
		Author: model.PersonLink{
			Name:       item.Author.Name,
			ProfileURL: item.Author.URL,
			ImageURL:   item.Author.Avatar,
		},
	}

	if item.ContentHTML != "" {
		activity.Content = model.NewHTMLContent(item.ContentHTML)
	} else if item.ContentText != "" {
		activity.Content = model.NewTextContent(item.ContentText)
	}

	return activity
}
