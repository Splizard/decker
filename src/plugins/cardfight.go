package plugins

import "fmt"
import "net/http"
import "net/url"
import "errors"
import "os"
import "regexp"
import "io"
import "io/ioutil"

const Cardfight = "cardfight"

func init() {
	var client http.Client
	
	//Regexes needed for parsing results from http://cardfight.wikia.com
	var cardfightregex *regexp.Regexp = regexp.MustCompile(`<a href="(.[0-9a-zA-z _\/\.\-(),:]*)" class="result-link"`)
	var cardfightimageregex *regexp.Regexp = regexp.MustCompile(`<a href="([^"]*)" 	class="image image-thumbnail"`)
	
	RegisterHeaders(Cardfight, []string{"Cardfight!! Vanguard", "Cardfight", "Cardfight! Vanguard"})
	RegisterBack(Cardfight, "https://vignette.wikia.nocookie.net/cardfight/images/3/37/Cfv_back.jpg/revision/latest?cb=20140801101556")
	
	RegisterPlugin(Cardfight, func(name, info string, detecting bool) string {

		if _, err := os.Stat( DeckerCachePath + "/cards/cardfight/" + name + ".jpg"); info == "" && !os.IsNotExist(err) {
			return Yugioh
		}

		var search string

		search = "http://cardfight.wikia.com/wiki/Special:Search?search=" + url.QueryEscape(name)

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
		submatches := cardfightimageregex.FindStringSubmatch(string(body))
		if len(submatches) < 2 {
			
			//Handle(errors.New("No image found for card " + name + ", this could be a bug !"))
			
			//Magical regex to our rescue.
			var card string
			matches := cardfightregex.FindStringSubmatch(string(body))
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
			
			submatches = cardfightimageregex.FindStringSubmatch(string(body))
			
			if len(submatches) < 1 {
				//Indeed.. a bug on wikia :3
				if !detecting {
					Handle(errors.New("No image found for card " + name + ", this could be a bug !"))
				} else {
					return None
				}
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
				fmt.Println("possible error check file! " + DeckerCachePath + "/cards/cardfight/" + name + ".jpg, status " + response.Status)
			}
			//Download and Save image.
			var imageOut *os.File
			imageOut, err = os.Create(DeckerCachePath + "/cards/cardfight/" + name + ".jpg")
			Handle(err)
			io.Copy(imageOut, response.Body)
			imageOut.Close()
			
			return Cardfight
		}
		return None
	})
}
