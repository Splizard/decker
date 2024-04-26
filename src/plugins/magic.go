// Magic.
// This plugin will download magic cards from scryfall.
// As a fallback it will use the gatherer website! (http://gatherer.wizards.com/)
package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const Magic = "magic"

type MagicCard struct {
	Images struct {
		Large string `json:"large"`
	} `json:"image_uris"`
	CardFaces []struct {
		ImageUris struct {
			Large string `json:"large"`
		} `json:"image_uris"`
	} `json:"card_faces"`
}

func init() {

	var client http.Client

	RegisterHeaders(Magic, []string{"Magic: The Gathering", "Magic", "MTG"})

	RegisterBack(Magic, "https://upload.wikimedia.org/wikipedia/en/a/aa/Magic_the_gathering-card_back.jpg")

	RegisterPlugin(Magic, func(name, info string, detecting bool) string {

		if _, err := os.Stat(DeckerCachePath + "/cards/magic/" + name + ".jpg"); !os.IsNotExist(err) {
			return Magic
		}

		response, err := client.Get("https://api.scryfall.com/cards/named?fuzzy=" + url.QueryEscape(name))
		Handle(err)

		if response.StatusCode != 200 {
			//Not sure what happens here.
			fmt.Println("possible error check card! " + name + ", status " + response.Status)
		}

		//Now we should parse the Json data of the card.
		var decoder = json.NewDecoder(response.Body)
		var card MagicCard

		err = decoder.Decode(&card)
		Handle(err)

		var image = card.Images.Large
		var imagename string

		if len(card.CardFaces) > 0 {
			image = card.CardFaces[0].ImageUris.Large
		}
		if image == "" {
			Handle(fmt.Errorf("no image found for card %s", name))
			return Magic
		}

		//Now we can check if we already have the image cached, otherwise download it.
		if _, err := os.Stat(DeckerCachePath + "/cards/magic/" + name + ".jpg"); !os.IsNotExist(err) {
			return Magic
		} else {
			if !detecting {
				fmt.Println("getting", image)
			}
			response, err = client.Get(image)
			Handle(err)
			if response.StatusCode == 404 {
				//Broken link?
				Handle(errors.New("broken link? " + image))
			} else {
				if response.StatusCode != 200 {
					fmt.Println("possible error check file! " + DeckerCachePath + "/cards/magic/" + imagename + ".jpg, status " + response.Status)
				}
				//Download and Save image.
				var imageOut *os.File
				if info != "" {
					imageOut, err = os.Create(DeckerCachePath + "/cards/magic/" + imagename + ".jpg")
				} else {
					imageOut, err = os.Create(DeckerCachePath + "/cards/magic/" + name + ".jpg")
				}
				Handle(err)
				io.Copy(imageOut, response.Body)
				imageOut.Close()
				return Magic
			}
		}
		return None
	})
}
