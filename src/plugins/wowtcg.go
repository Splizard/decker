//Magic.
//This plugin will download magic cards from http://wowcards.info
/*
	Card images can be found on a line like this:
	 <img id="cardimage" width="312" height="440" class="img-rounded lazy swapImage {src: '/scans/timewalkers/en/3-back.jpg'}"  style="border: 1px solid black;" alt="Barador, Wildhammer Timewalker" title="Barador, Wildhammer Timewalker" data-original="/scans/timewalkers/en/3_Barador-Wildhammer-Timewalker.jpg" src="/scans/no-scan-upperdeck.jpg" />
	 data-original is what we are after.
*/
package plugins

import "fmt"
import "net/http"
import "net/url"
import "errors"
import "os"
import "regexp"
import "io"
import "io/ioutil"

const Wow = "wow"

func init() {

	var client http.Client
	
	//data-original is what we are after.
	var wowimageregex *regexp.Regexp = regexp.MustCompile(`data-original="([0-9a-zA-z \/_\.\-,:]*)`)

	RegisterHeaders(Wow, []string{"World of Warcraft", "WoW", "WoWTCT", "World of Warcraft TCG", "World of Warcraft Trading Card Game"})
	
	RegisterBack(Wow, "http://img1.wikia.nocookie.net/__cb20061106231523/wowwiki/images/9/9e/WoWTCG-Full.jpg")

	RegisterPlugin(Wow, func(name, info string, detecting bool) (game string) {
		//Don't crash the whole program when a bad error panics a goroutine.
		//Simply report and let the others continue.
		defer func() {
			if r := recover(); r != nil {
				game = None
			}
		}()

		var imagename string = name

		if _, err := os.Stat( DeckerCachePath + "/cards/wow/" + name + ".jpg"); !os.IsNotExist(err) {
			return Wow
		}
		
		//We need to post a search form.
		response, err := http.PostForm("http://wowcards.info/search",
	url.Values{"title": {name}})

		Handle(err)
		
		if response.StatusCode != 200 {
			//Not sure what happens here.
			fmt.Println("possible error check card! " + name + ", status " + response.Status)
		}
		
		body, err := ioutil.ReadAll(response.Body)
		Handle(err)
		
		//Magic Image Regex!
		submatches := wowimageregex.FindStringSubmatch(string(body))
		if len(submatches) < 2 {
			//Indeed.. a bug on wowcards.info :3
			Handle(errors.New("No image found for card " + name + ", this could be a bug !"))
		}
		image := "http://wowcards.info"+string(submatches[1])

		//Now we can check if we already have the image cached, otherwise download it.
		if _, err := os.Stat(DeckerCachePath  + "/cards/wow/" + name + ".jpg"); !os.IsNotExist(err) {
			return Wow
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
					fmt.Println("possible error check file! " + DeckerCachePath + "/cards/wow/" + imagename + ".jpg, status " + response.Status)
				}
				//Download and Save image.
				var imageOut *os.File
				if info != "" {
					imageOut, err = os.Create(DeckerCachePath + "/cards/wow/" + imagename + ".jpg")
				} else {
					imageOut, err = os.Create(DeckerCachePath + "/cards/wow/" + name + ".jpg")
				}
				Handle(err)
				io.Copy(imageOut, response.Body)
				imageOut.Close()
				return Wow
			}
		}
		return None
	})
}
