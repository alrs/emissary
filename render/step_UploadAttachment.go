package render

import (
	"encoding/json"
	"io"

	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/derp"
	"github.com/benpate/rosetta/mapof"
)

// StepUploadAttachment represents an action that can upload attachments.  It can only be used on a StreamRenderer
type StepUploadAttachment struct {
	Maximum int
}

func (step StepUploadAttachment) Get(renderer Renderer, _ io.Writer) PipelineBehavior {
	return nil
}

func (step StepUploadAttachment) Post(renderer Renderer, buffer io.Writer) PipelineBehavior {

	// Read the multipart form from the request
	form, err := multipartForm(renderer.request())

	if err != nil {
		return Halt().WithError(derp.Wrap(err, "handler.StepUploadAttachment.Post", "Error reading multipart form."))
	}

	files := form.File["file"]
	isEditorJS := false

	// Auto-detect EditorJS
	if len(files) == 0 {
		files = form.File["image"]
		isEditorJS = true
		step.Maximum = 1
	}

	factory := renderer.factory()
	attachmentService := factory.Attachment()

	objectID := renderer.objectID()
	objectType := renderer.service().ObjectType()

	// Special case:  If we're uploading a draft, then we need to attach the document to the parent stream.
	if objectType == "StreamDraft" {
		objectType = "Stream"
	}

	for index, fileHeader := range files {

		attachment := model.NewAttachment(objectType, objectID)
		attachment.Original = fileHeader.Filename

		// Open the source (from the POST request)
		source, err := fileHeader.Open()

		if err != nil {
			return Halt().WithError(derp.Wrap(err, "handler.StepUploadAttachment.Post", "Error reading file from multi-part header", fileHeader))
		}

		defer source.Close()

		// Add the image into the media server
		width, height, err := factory.MediaServer().Put(attachment.AttachmentID.Hex(), source)

		if err != nil {
			return Halt().WithError(derp.Wrap(err, "handler.StepUploadAttachment.Post", "Error saving attachment to mediaserver", attachment))
		}

		// Update original dimensions
		attachment.Width = width
		attachment.Height = height

		// Try to save
		if err := attachmentService.Save(&attachment, "Uploaded file: "+fileHeader.Filename); err != nil {
			return Halt().WithError(derp.Wrap(err, "handler.StepUploadAttachment.Post", "Error saving attachment", attachment))
		}

		// EditorJS can only upload a single file at a time.
		if isEditorJS {
			response := mapof.Any{
				"success": 1,
				"file": mapof.Any{
					"url":    renderer.Host() + attachment.URL(),
					"height": attachment.Height,
					"width":  attachment.Width,
				},
				"data": mapof.Any{
					"filePath": renderer.Host() + attachment.URL(),
				},
			}

			// Marshal the response into JSON
			bytes, err := json.Marshal(response)

			if err != nil {
				return Halt().WithError(derp.Wrap(err, "handler.StepUploadAttachment.Post", "Error marshalling response", response))
			}

			// Write the response to the buffer
			if _, err := buffer.Write(bytes); err != nil {
				return Halt().WithError(derp.Wrap(err, "handler.StepUploadAttachment.Post", "Error writing response to buffer", response))
			}

			// Tell the client that we're done.
			return Halt().WithContentType("application/json")
		}

		// If we have reached the maximum configured number of attachments,
		// then stop processing here.
		if (step.Maximum > 0) && index >= step.Maximum {
			break
		}
	}

	// After all files are uploaded, tell the client that we're done.
	return Continue().WithEvent("attachments-updated", "true")
}
