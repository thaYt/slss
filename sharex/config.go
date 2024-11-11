package sharex

import "strings"

const template string = `{
  "Version": "14.0.0",
  "DestinationType": "ImageUploader, TextUploader, FileUploader",
  "RequestMethod": "POST",
  "RequestURL": "https://%SITE%/upload",
  "Headers": {
	"Authorization": "%AUTH%"
  },
  "Body": "MultipartFormData",
  "FileFormName": "file",
  "URL": "{json:url}",
  "DeletionURL": "{json:del_url}",
  "ErrorMessage": "{json:error}"
}`

func GenConfig(site, auth string) string {
	p1 := strings.ReplaceAll(template, "%SITE%", site)
	p2 := strings.ReplaceAll(p1, "%AUTH%", auth)
	return p2
}
