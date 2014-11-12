package plugins

import "fmt"
import "net/http"
import "net/url"
import "errors"
import "os"
import "regexp"
import "io"
import "io/ioutil"
import "strings"
import "path/filepath"

const Pokemon = "pokemon"

func init() {
	var client http.Client
	
	//Regexes needed for parsing results from pkmncards.com
	var pokemonregex *regexp.Regexp = regexp.MustCompile(`http://pkmncards\.com/card/(.[0-9a-zA-z _\.\-,:]*)/`)
	var pokemonimageregex *regexp.Regexp = regexp.MustCompile(`"og:image"\scontent="([0-9a-zA-z \/_\.\-,:]*)`)
	
	RegisterHeaders(Pokemon, []string{"Pokémon Trading Card Game", "Pokemon Trading Card Game", "Pokemon"})
	
	RegisterPlugin(Pokemon, func(name, info string, detecting bool) string {

		if _, err := os.Stat( DeckerCachePath + "/cards/pokemon/" + name + ".jpg"); info == "" && !os.IsNotExist(err) {
			return Pokemon
		}

		//Format url, pkmncards.com does not like an empty text:"" field.
		var oldname string = name
		var search string
		var imagename string
		
		//This bit recognises extra information to be queried along with the card name.
		//This solves the problem with card games where there are many cards of the same name.
		//Looking at you Pokemon -.-
		//So people who don't know their pokemon set names can be like:
		//
		//	1x Pikachu with Thundershock
		//  1x Pikachu, Thundershock
		//  1x Pikachu that has Thundershock
		//
		//Hopefully they get the card they want or at the very least they get a Pickachu that knows Thundershock.
		info = ""
		if strings.Contains(name, ",") {
			splits := strings.Split(name, ",")
			name = splits[0]
			info = strings.TrimSpace(splits[1])
		}
		if strings.Contains(name, " with ") {
			splits := strings.Split(name, " with ")
			name = splits[0]
			info = strings.TrimSpace(splits[1])
		}
		if strings.Contains(name, " that has ") {
			splits := strings.Split(name, " that has ")
			name = splits[0]
			info = strings.TrimSpace(splits[1])
		}
		
		if info != "" {
			search = "http://pkmncards.com/?s=" + url.QueryEscape(name) + "+text%3A%22" + url.QueryEscape(info) + "%22&display=scan&sort=date"
		} else {
			search = "http://pkmncards.com/?s=" + url.QueryEscape(name) + "%22&display=scan&sort=date"
		}

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

		//Magical regex to our rescue.
		card := string(pokemonregex.Find([]byte(body)))
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
		body, err = ioutil.ReadAll(response.Body)
		Handle(err)

		//Regex!
		submatches := pokemonimageregex.FindStringSubmatch(string(body))
		if len(submatches) < 2 {
			//Indeed.. a bug on pkmncards.com :3
			Handle(errors.New("No image found for card " + name + ", this could be a bug !"))
		}
		image := string(submatches[1])

		//Extract the filename for the cache.
		path, err := url.Parse(image)
		Handle(err)
		
		imagename = strings.Replace(filepath.Base(path.Path), ".jpg", "", 1)
		if info != "" {
			SetImageName(Pokemon, oldname, imagename)
		}

		//Now we can check if we already have the image cached, otherwise download it.
		if _, err := os.Stat(DeckerCachePath  + "/cards/pokemon/" + imagename + ".jpg"); !os.IsNotExist(err) {
			return Pokemon
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
					fmt.Println("possible error check file! " + DeckerCachePath + "/cards/pokemon/" + imagename + ".jpg, status " + response.Status)
				}
				//Download and Save image.
				var imageOut *os.File
				if info != "" {
					imageOut, err = os.Create(DeckerCachePath + "/cards/pokemon/" + imagename + ".jpg")
				} else {
					imageOut, err = os.Create(DeckerCachePath + "/cards/pokemon/" + name + ".jpg")
				}
				Handle(err)
				io.Copy(imageOut, response.Body)
				imageOut.Close()
				return Pokemon
			}
		}
		return None
	})
}
