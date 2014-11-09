package plugins

import "fmt"
import "net/http"
import "net/url"
import "errors"
import "os"
import "io"
import "path/filepath"
import "strings"

const Custom = "custom"

func init() {

	var client http.Client

	RegisterHeaders(Custom, []string{"CUSTOM", "Custom"})
		
	RegisterCacheImageName(Custom, func(name string) string {
		path, err := url.Parse(name)
		Handle(err)
		return strings.Replace(filepath.Base(path.Path), ".jpg", "", 1)
	})

	RegisterPlugin(Custom, func(name, info string, detecting bool) string {

		path, err := url.Parse(name)
		Handle(err)
		imagename := strings.Replace(filepath.Base(path.Path), ".jpg", "", 1)
		SetImageName(Custom, name, imagename)	
		
		if _, err := os.Stat( DeckerCachePath + "/cards/custom/" + imagename + ".jpg"); !os.IsNotExist(err) {
			return Custom
		}
		
		if !detecting {
			fmt.Println("getting", name)
		}
		
		//For custom cards, simply download the image with the provided link.
		response, err := client.Get(name)
		if detecting && err != nil {
			return None
		}
		Handle(err)

		//Unless we get a 404 which means the link is probably broken.
		if response.StatusCode == 404 {
			if !detecting {
				//Complain about it.
				Handle(errors.New("link '" + name + "' seems to be broken!\nCheck it!"))
			}
			if detecting {
				//Or it just means this is not a custom deck.
				return None
			}
		} else {
			if response.StatusCode != 200 {
				//Hmmm why is the status code not 200?
				println("possible error check file! " + DeckerCachePath + "/cards/magic/" + filepath.Base(path.Path)+ " status " + response.Status)
			}
			//Download and Save image.
			
			imageOut, err := os.Create(DeckerCachePath + "/cards/custom/" + imagename + ".jpg")
			Handle(err)
			io.Copy(imageOut, response.Body)
			imageOut.Close()
			return Custom
		}
		return None
	})
}
