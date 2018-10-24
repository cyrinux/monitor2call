package translate

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"github.com/cyrinux/monitor2call/models" //structs
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

// DoTTS convert message to mp3 audio
func DoTTS(tts *models.TTS) (filename string, err error) {
	filename = tts.Filename + "-" + tts.LanguageCode + ".mp3"
	cacheDir := os.Getenv("CACHE_DIR")
	if cacheDir == "" {
		cacheDir = "./cache"
	}
	filepath := cacheDir + "/" + filename
	log.Printf("filepath is %s", filepath)

	// if file allready exists we return the filename
	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		log.Printf("Audio content file allready exists: %v\n", filename)
		return filename, err
	}

	// Instantiates a client.
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Print(err)
	}

	// Perform the text-to-speech request on the text input with the selected
	// voice parameters and audio file type.
	req := texttospeechpb.SynthesizeSpeechRequest{
		// Set the text input to be synthesized.
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: tts.Text},
		},
		// Build the voice request, select the language code ("en-US") and the SSML
		// voice gender ("neutral").
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: tts.LanguageCode,
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		log.Print(err)
	}

	// The resp's AudioContent is binary.
	err = ioutil.WriteFile(filepath, resp.AudioContent, 0666)
	if err != nil {
		log.Print(err)
	}
	log.Printf("Audio content written to file: %v\n", filepath)

	return filename, err
}
