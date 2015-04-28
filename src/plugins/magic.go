//Magic.
//This plugin will download magic cards from http://magiccards.info
package plugins

import "fmt"
import "net/http"
import "net/url"
import "errors"
import "os"
import "regexp"
import "io"
import "io/ioutil"

const Magic = "magic"

func init() {

	var client http.Client
	
	var magicimageregex *regexp.Regexp = regexp.MustCompile(`http://magiccards.info/scans/([0-9a-zA-z \/_\.\-,:]*)`)

	RegisterHeaders(Magic, []string{"Magic: The Gathering", "Magic", "MTG"})
	
	RegisterBack(Magic, "http://mtgimage.com/card/cardback.hq.jpg")

	RegisterPlugin(Magic, func(name, info string, detecting bool) string {

		var search string
		var imagename string = name

		if _, err := os.Stat( DeckerCachePath + "/cards/magic/" + name + ".jpg"); !os.IsNotExist(err) {
			return Magic
		}
		
		search = "http://magiccards.info/query?q="+url.QueryEscape(name)+"&v=card&s=cname"
		
		response, err := client.Get(search)
		Handle(err)
		
		if response.StatusCode != 200 {
			//Not sure what happens here.
			fmt.Println("possible error check card! " + name + ", status " + response.Status)
		}
		
		body, err := ioutil.ReadAll(response.Body)
		Handle(err)
		
		//Magic Image Regex!
		submatches := magicimageregex.FindStringSubmatch(string(body))
		if len(submatches) < 2 {
			//Indeed.. a bug on magiccards.info :3
			Handle(errors.New("No image found for card " + name + ", this could be a bug !"))
		}
		image := "http://magiccards.info/scans/"+string(submatches[1])

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
