package convert

import (
	"time"

	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/data"
	"github.com/benpate/rosetta/first"
	"github.com/benpate/rosetta/html"
	"github.com/benpate/rosetta/iterator"
	"github.com/benpate/rosetta/slice"
	"github.com/kr/jsonfeed"
)

func IteratorToJSonFeed(url string, title string, description string, it data.Iterator) jsonfeed.Feed {

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
		Items: slice.Map(iterator.Slice(it, model.NewStream), StreamToJsonFeed),
	}
}

func StreamToJsonFeed(stream model.Stream) jsonfeed.Item {

	result := jsonfeed.Item{
		ID:            stream.Token,
		URL:           stream.URL,
		Title:         stream.Label,
		ContentHTML:   first.String(stream.Content.HTML, " "),
		Summary:       stream.Summary,
		Image:         stream.ImageURL,
		DatePublished: time.Unix(stream.PublishDate, 0),
		DateModified:  time.Unix(stream.UpdateDate, 0),
	}

	// Attach author if available
	if stream.AttributedTo.NotEmpty() {
		result.Author = &jsonfeed.Author{
			Name:   stream.AttributedTo.Name,
			URL:    stream.AttributedTo.ProfileURL,
			Avatar: stream.AttributedTo.ImageURL,
		}
	}

	// TODO: LOW: Attachments for podcasts, etc.

	return result
}

func JsonFeedToActivity(feed jsonfeed.Feed, item jsonfeed.Item) model.Message {

	message := model.NewMessage()

	message.Origin = model.OriginLink{
		Label:    feed.Title,
		URL:      feed.HomePageURL,
		ImageURL: feed.Icon,
	}

	message.URL = item.URL
	message.Label = item.Title
	message.Summary = item.Summary
	message.ImageURL = item.Image
	message.ContentHTML = JsonFeedToContentHTML(item)
	message.PublishDate = item.DatePublished.Unix()
	message.SetAttributedTo(JsonFeedToAuthor(feed, item))

	return message
}

func JsonFeedToAuthor(feed jsonfeed.Feed, item jsonfeed.Item) model.PersonLink {

	result := model.NewPersonLink()

	if feed.Author != nil {
		result.Name = feed.Author.Name
		result.ProfileURL = feed.Author.URL
		result.ImageURL = feed.Author.Avatar
	}

	if item.Author != nil {
		result.Name = first.String(item.Author.Name, result.Name)
		result.ProfileURL = first.String(item.Author.URL, result.ProfileURL)
		result.ImageURL = first.String(item.Author.Avatar, result.ImageURL)
	}

	return result
}

func JsonFeedToContentHTML(item jsonfeed.Item) string {

	var result string

	if item.ContentHTML != "" {
		result = item.ContentHTML
	} else if item.ContentText != "" {
		result = html.FromText(item.ContentText)
	}

	return SanitizeHTML(result)
}
