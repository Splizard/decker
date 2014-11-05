package plugins

import "fmt"
import "net/http"
import "net/url"
import "errors"
import "os"
import "regexp"
import "io"
import "io/ioutil"

const Yugioh = "yugioh"

func init() {
	var client http.Client
	
	//Regexes needed for parsing results from http://yugioh.wikia.com
	var yugiohregex *regexp.Regexp = regexp.MustCompile(`<a href="([A-Za-z =0-9:/.,_-]*)" class="result-link"`)
	var yugiohimageregex *regexp.Regexp = regexp.MustCompile(`"cardtable-cardimage" [A-Za-z ="0-9>]*<a href="([^"]*)"`)
	
	RegisterHeaders(Yugioh, []string{"Yu-Gi-Oh", "Yugioh", "yu-gi-oh"})
	
	RegisterPlugin(Yugioh, func(name, info string, detecting bool) string {

		if _, err := os.Stat( DeckerCachePath + "/cards/yugioh/" + name + ".jpg"); info == "" && !os.IsNotExist(err) {
			return Yugioh
		}

		//Format url, pkmncards.com does not like an empty text:"" field.
		var search string

		search = "http://yugioh.wikia.com/wiki/Special:Search?search=" + url.QueryEscape(name)

		//This returns the search results for the card.
		response, err := client.Get(search)
		Handle(err)

		if response.StatusCode == 404 {
			//No results, complain, doubt users spelling ability.
			if !detecting {
				Handle(errors.New("card name '" + name + "' invalid! (Check spelling?)"))
			} else {
				return None
			}
		} else if response.StatusCode != 200 {
			//Not sure what happens here.
			fmt.Println("possible error check card! " + name + ", status " + response.Status)
		}

		//We need find the first result.
		body, err := ioutil.ReadAll(response.Body)
		Handle(err)

		/*
		
		*/
		

		//Regex!
		submatches := yugiohimageregex.FindStringSubmatch(string(body))
		if len(submatches) < 2 {
			
			//Handle(errors.New("No image found for card " + name + ", this could be a bug !"))
			
			//Magical regex to our rescue.
			var card string
			matches := yugiohregex.FindStringSubmatch(string(body))
			if len(matches) > 1 {
				card = matches[1]
			}
			if card == "" {
				//regex failed?
				if !detecting {
					Handle(errors.New("card name '" + name + "' not found!\nCheck " + search))
				} else {
					return None
				}
			}

			//Now we need to find the link to the actual image.
			response, err = client.Get(card)
			Handle(err)
			body, err = ioutil.ReadAll(response.Body)
			Handle(err)
			
			submatches = yugiohimageregex.FindStringSubmatch(string(body))
			
			if len(submatches) < 1 {
				//Indeed.. a bug on wikia :3
				Handle(errors.New("No image found for card " + name + ", this could be a bug !"))
			}

		}
		image := string(submatches[1])
		
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
				fmt.Println("possible error check file! " + DeckerCachePath + "/cards/yugioh/" + name + ".jpg, status " + response.Status)
			}
			//Download and Save image.
			var imageOut *os.File
			imageOut, err = os.Create(DeckerCachePath + "/cards/yugioh/" + name + ".jpg")
			Handle(err)
			io.Copy(imageOut, response.Body)
			imageOut.Close()
			return Yugioh
		}
		return None
	})
}
