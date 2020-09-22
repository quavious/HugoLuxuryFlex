package translate

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/translate"
	"github.com/quavious/GoSummary/credential"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

//Convert returns translated text.
func Convert(targetLanguage, text string) (string, error) {
	ctx := context.Background()
	jsonPath := credential.GoogleJSONPath
	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", fmt.Errorf("language.Parse: %v", err)
	}

	client, err := translate.NewClient(ctx, option.WithCredentialsFile(jsonPath))
	time.Sleep(1)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer client.Close()

	resp, err := client.Translate(ctx, []string{text}, lang, nil)

	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("Translate: %v", err)
	}
	if len(resp) == 0 {
		return "", fmt.Errorf("Translate returned empty response to text: %s", text)
	}
	return resp[0].Text, nil
}
