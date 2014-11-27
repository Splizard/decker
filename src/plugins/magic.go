//Magic.
//This plugin will download magic cards from mtgimage.com
package plugins

import (
	"fmt"
	"net/http"
	"errors"
	"os"
	"io"
)

const Magic = "magic"

func init() {

	var client http.Client

	RegisterHeaders(Magic, []string{"Magic: The Gathering", "Magic", "MTG"})
	
	RegisterBack(Magic, "http://mtgimage.com/card/cardback.hq.jpg")

	RegisterPlugin(Magic, func(name, info string, detecting bool) string {

		if !detecting {
			fmt.Println("getting", "http://mtgimage.com/card/"+name+".jpg")
		} else {
			if _, err := os.Stat( DeckerCachePath + "/cards/magic/" + name + ".jpg"); !os.IsNotExist(err) {
				return Magic
			}
		}

		//For magic cards it is easy we just request the name from mtgimage.com and tada! we have an image.
		response, err := client.Get("http://mtgimage.com/card/" + name + ".jpg")
		Handle(err)

		//Unless we get a 404 which means the name wasn't given correctly.
		if response.StatusCode == 404 {
			if !detecting {
				//Complain about it.
				Handle(errors.New("card name '" + name + "' seems to be invalid!\nCheck http://mtgimage.com/card/" + name + ".jpg"))
			} else {
				//Or it just means this is not a magic deck.
				return None
			}
		} else {
			if response.StatusCode != 200 {
				//Hmmm why is the status code not 200?
				println("possible error check file! " + DeckerCachePath + "/cards/magic/" + name + ".jpg, status " + response.Status)
			}
			//Download and Save image.
			imageOut, err := os.Create(DeckerCachePath + "/cards/magic/" + name + ".jpg")
			Handle(err)
			io.Copy(imageOut, response.Body)
			imageOut.Close()
			return Magic
		}
		return None
	})
}
