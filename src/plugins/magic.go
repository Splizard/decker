//Magic.
//This plugin will download magic cards from scryfall.
//As a fallback it will use the gatherer website! (http://gatherer.wizards.com/)
package plugins

import "fmt"
import "net/http"
import "net/url"
import "errors"
import "os"
import "io"
import "encoding/json"

const Magic = "magic"

type MagicCard struct {
	Images struct{
		
		Large string `json:"large"`
		
	} `json:"image_uris"`
}

func init() {

	var client http.Client

	RegisterHeaders(Magic, []string{"Magic: The Gathering", "Magic", "MTG"})
	
	RegisterBack(Magic, "https://upload.wikimedia.org/wikipedia/en/a/aa/Magic_the_gathering-card_back.jpg")

	RegisterPlugin(Magic, func(name, info string, detecting bool) string {
		
		if _, err := os.Stat( DeckerCachePath + "/cards/magic/" + name + ".jpg"); !os.IsNotExist(err) {
			return Magic
		}
		
		response, err := client.Get("https://api.scryfall.com/cards/named?fuzzy="+url.QueryEscape(name))
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

		//Now we can check if we already have the image cached, otherwise download it.
		if _, err := os.Stat(DeckerCachePath  + "/cards/magic/" + name + ".jpg"); !os.IsNotExist(err) {
			return Magic
		} else {
			if ! detecting {
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
