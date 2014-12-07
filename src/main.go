/*
	Decker, a tool for generating card decks for Tabletop Simulator
	Copyright (C) 2014 Quentin Quaadgras

    This program is free software; you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation; version 2 of the License.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License along
    with this program; if not, write to the Free Software Foundation, Inc.,
    51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"path"
	"regexp"
	"runtime"
	"strings"
	"strconv"
	"sync"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
	"encoding/json"
	"encoding/base64"
	"errors"
	"runtime/debug"
	//"html"
)

import "./ct"
import "./plugins"

//Error handler, all bad errors will be sent here.
func handle(err error) {
	if err != nil {
		panic(err.Error())
	}
}

//A nice copy function that will handle errors.
func Copy(src, dst string) (int64, error) {
	src_file, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer src_file.Close()

	src_file_stat, err := src_file.Stat()
	if err != nil {
		return 0, err
	}

	if !src_file_stat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	dst_file, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dst_file.Close()
	return io.Copy(dst_file, src_file)
}

var deck string
var output string //The output file.

var client http.Client = http.Client{
    Timeout: time.Duration(1 * time.Second),
}

func init() {
	flag.StringVar(&output, "o", "", "output file")
}

//Define the current card games decker supports.
const (
	None    = "none"
)

type ImgurData struct {
	Error string 		`json:"error"`
	Link string 		`json:"link"`
	DeleteHash string 	`json:"deletehash"`
}

type ImgurResponse struct {
	Data ImgurData 	`json:"data"`
	Success bool	`json:"success"`
	Status int		`json:"status"`
}

type ImgurRequest struct {
	Image string	`json:"image"`
	Type string 	`json:"type"`
	Name string		`json:"name"`
}

func upload(filename string) {
	if file, err := os.Open(filename); err == nil {
			
		fmt.Println("Uploading "+filename+"...")		
		img, err := ioutil.ReadAll(file)
		handle(err)
		
		v := url.Values{}
		v.Set("image", base64.StdEncoding.EncodeToString(img))
		v.Set("type", "base64")
		handle(err)
		
		reader := strings.NewReader(v.Encode())
	
		request, err := http.NewRequest("POST", "https://api.imgur.com/3/image", reader)
		handle(err)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Authorization", "Client-ID 47122dc42ac36ad")
		
		response, err := http.DefaultClient.Do(request)
		handle(err)
		
		body, err := ioutil.ReadAll(response.Body)
		handle(err)
		
		imgur := ImgurResponse{}
		json.Unmarshal(body, &imgur)
		if imgur.Success {
			link, err := url.QueryUnescape(imgur.Data.Link)
			handle(err)
			
			name := filepath.Base(filename)[:len(filepath.Base(filename))-4]+".json"
			
			file, err := os.Open(chest+"/"+name)
			if err == nil {
		 	   	data, err := ioutil.ReadAll(file)
				if err == nil {
					data = []byte(strings.Replace(string(data), "http://"+ip_address+":20002/"+filepath.Base(filename), link, -1))
					data = []byte(strings.Replace(string(data), "http://localhost:20002/"+filepath.Base(filename), link, -1))
					ioutil.WriteFile(chest+"/"+name, data, 0644)
				}
		    } else {
		    	handle(err)
		    }
		    
		    //Yay we did it!
			ct.ChangeColor(ct.Green, true, ct.None, false)
			fmt.Print("Done ")
			ct.ResetColor()
			fmt.Println(filename + "!")
		    
		    fmt.Println(filepath.Base(filename)+" can now be found at the following location: "+link)
		    fmt.Println(filepath.Base(filename)+" can be deleted by visiting :  https://api.imgur.com/3/image/"+imgur.Data.DeleteHash)
		} else {
			handle(errors.New(imgur.Data.Error))
		}
	}
}

//Decker function, can be called from a goroutine to generate decks in parallel.
//(Don't know if concurrency is really going to be used much other then bulk testing but this is Go so why not!)
func decker(filename string) {

	//Don't crash the whole program when a bad error panics a goroutine.
	//Simply report and let the others continue.
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			_, ok = r.(error)
			if !ok {
				fmt.Print("[ERROR] ")
				fmt.Println(fmt.Errorf("file: "+filename+": %v", r))
				ct.ChangeColor(ct.Red, true, ct.None, false)
				fmt.Print("Failed ")
				ct.ResetColor()
				fmt.Println(filename + "!")
				return
			} else {
				fmt.Print("[ERROR] ")
				fmt.Println(fmt.Errorf("file: "+filename+": %v", r))
			}
		}
	}()
	
	//Leave the wait group.
	if threading {
		defer wg.Done()
	}
		
	if filepath.Ext(filename) == ".jpg" {
		upload(filename)
		return
	}

	var name string      //The name of the card.
	var info string      //Extra details to identify the card, mainly for Pokemon.
	
	var total int 		 //Number of cards in the deck.

	var game string = None //The current game as defined by the game constants.

	//output file, this is to keep track of per-file outputs when running in parallel.
	//we want to output to png.
	var output = output
	if threading || output == "" {
		output = filepath.Dir(filename)+"/"+filepath.Base(filename) + ".jpg"
	}

	var usingCache bool //Whether we have started using cached files or not.

	//temp stores the location of assigned temp directory, eg /tmp/decker-893282948 on linux.
	var temp string
	temp, err := ioutil.TempDir("", "decker")
	handle(err)
	defer func() {
		if temp != "" {
			os.RemoveAll(temp)
		}
	}()
	
	var autodetecting bool
	var possibilities []string
	var statistics = make(map[string]int)

	//Open the deck file. TODO maybe support http:// decks.
	if file, err := os.Open(filename); err == nil {

		//Read the first line and trim the space.
		reader := bufio.NewReader(file)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)

		//If possible we want to indentify the name of the Card game.
		//These names should be at the top of a deck file.
		game = plugins.Identify(line)
		
		//If there is no header, make a big deal about it.
		if game == None {
			ct.ChangeColor(ct.Red, true, ct.None, false)
			fmt.Print("Warning! ")
			ct.ResetColor()
			fmt.Println("this deck file does not have a identifyable header...\nFalling back to auto-detection.")

			//Complain to the user, they have committed a great sin.
			fmt.Println("It is STRONGLY recommended that you add a identifier at the top of the file.")
			fmt.Println("This makes it easier for people to recognise the card game...\n")
			
			autodetecting = true
		}

		//Loop through the file.
		for {
			line, err := reader.ReadString('\n') //Parse line-by-line.
			if err == io.EOF {
				break
			}
			handle(err)

			//Trim the spacing. TODO trim spacing in between words that are used for nice reading.
			line = strings.TrimSpace(line)

			//Cards are identified by having an 'x' at the beginning of the line or a number.
			//Anyother character is a comment.
			//Not many words start with x so we should be pretty safe, let's not worry about dealing with special cases.
			//This may look like a complicated if statement but don't worry about understanding it.. it works.
			//That being said, feel free to simplify if you are one of those people.

			//Compile a regular expression to test if the line is a card
			r, _ := regexp.Compile("^((\\d+x)|(x?\\d+)) +[^ \n]+")

			//Check if the line is a card 
			// (nx, n or xn followed by at least one space and then anything not space)
			if r.MatchString(line) {
				//We need to seperate the name from the number of cards.
				//This does that.
				r, _ := regexp.Compile("^((\\d+x)|(x?\\d+))")

				name = r.ReplaceAllString(line, "");
				name = strings.Join(strings.Fields(name), " ")
				
				if autodetecting {
					
					possibilities = plugins.Autodetect(name, info)
					for _, v := range possibilities {
						if _, yes := statistics[v]; yes {
							statistics[v] += 1
						} else {
							statistics[v] = 1
						}
					}
					max := 0
					id := ""
					for i, v := range statistics {
						if v > max {
							max = v
							id = i
						}
					}
					for i, v := range statistics {
						if id != i && max == v {
							goto pass 
						}
					}
					game = id
					autodetecting = false
					file.Close()
					file, err = os.Open(filename)
					handle(err)
					reader = bufio.NewReader(file)
					ct.ChangeColor(ct.Green, true, ct.None, false)
					fmt.Print("It's OK ")
					ct.ResetColor()
					fmt.Println("Game appears to be '"+game+"'")
					
					pass:
				} else {
				
					var cache = cache
					var imagename = name

					//If the imagename is different from the card name,
					if i := plugins.GetImageName(game, name); i != "" {
						imagename = i
					}
					
					//Let's check if the card we are looking for has already been downloaded.
					//Plugins may handle this by themselves.
					if _, err := os.Stat(cache + "/cards/" + game + "/" + imagename + ".jpg"); !os.IsNotExist(err) {
						if !usingCache {
							fmt.Println("using cached files for "+filename)
							usingCache = true
						}
					} else if _, err := os.Stat(filepath.Dir(filename) + "/cards/" + game + "/" + imagename + ".jpg"); !os.IsNotExist(err) {
						if !usingCache {
							fmt.Println("using cached files for "+filename)
							usingCache = true
						}
						cache = filepath.Dir(filename)
						
					} else {
						plugins.Run(game, name, info)
						if i := plugins.GetImageName(game, name); i != "" {
							imagename = i
						}
					}
					
					//If the imagename is different from the card name, we replace it now so everything works.
					name = imagename

					//Copy the card from cache to the temp directory.
					if _, err := os.Stat(temp + "/" + name + ".jpg"); os.IsNotExist(err) {
						_, err := Copy(cache+"/cards/"+game+"/"+name+".jpg", temp+"/"+name+".jpg")
						handle(err)
					}

					//Figure out how many cards there are in the deck.
					//Maximum is 99 otherwise unpredictable things will happen.
					//Should probably note this somewhere.

					//Get the count of cards by getting the xn, nx or n part and replacing the x
					count, _ := strconv.Atoi(strings.Replace(r.FindString(line), "x", "", -1));

					//Create copies of the card in the temporary directory.
					total += 1
					for i := 1; i < count; i++ {

						if _, err := os.Stat(temp + "/" + name + " " + fmt.Sprint(i+1) + ".jpg"); os.IsNotExist(err) {
							
							total += 1

							//Symbolic links don't like windows very much.. So we'll just have to copy the file multiple times.
							if runtime.GOOS == "windows" {
								Copy(cache+"/cards/"+game+"/"+name+".jpg", temp+"/"+name+" "+fmt.Sprint(i+1)+".jpg")
							} else {
								os.Symlink("./"+name+".jpg", temp+"/"+name+" "+fmt.Sprint(i+1)+".jpg")
							}

						}
					}
				}
			}
		}

		//Now we actually generate the image.
		fmt.Println("Generating image for " + filename + " to " + output + "...")

		//We use imagemagick's montage to generate the image,
		//somebody could code it in go using it's image library but I can't be bothered as imagemagick already does a perfect job.
		//Why rewrite something that already exists when you can just glue a bunch of different programs together?
		command := "montage"

		//Windows doesn't like it when you drag a deck file onto decker from a different folder.
		//Then it makes the different folder the current working directory and complains
		//when it can't find montage.exe that you packaged in the same folder.
		if runtime.GOOS == "windows" {
			command, err = filepath.Abs(os.Args[0])
			command = filepath.Dir(command)
			if err != nil {
				command = "montage"
			} else {
				command += "/montage"
			}
		}

		//Run montage. TODO maybe make these values tweakable, for now they do a fine job.
		montage := exec.Command(command, "-background", "rgb(23,20,15)", "-tile", "10x7", "-quality", "100", "-geometry", "410x586!+0+0", temp+"/*.jpg", output)
		text, err := montage.CombinedOutput()
		if err != nil {
			fmt.Print(string(text))
			handle(err)
		}
		
		//Crop the deck to a power of 2, 4096x4096 this will overwrite the file as a compressed jpeg.
		err = CropDeck(output)
		handle(err)
		
		fmt.Println("Creating Tabletop file...")
		
		//Copy to handler directory.
		Copy(filename, cache+"/decks/"+filepath.Base(filename)+".deck")
		Copy(output, cache+"/images/"+filepath.Base(filename)+".jpg")
		
		if back := plugins.GetBack(game); back != "" {
			if _, err := os.Stat( cache + "/images/"+game+".jpg"); os.IsNotExist(err) {
				response, err := client.Get(back)
				handle(err)
				imageOut, err := os.Create(cache + "/images/"+game+".jpg")
				handle(err)
				io.Copy(imageOut, response.Body)
			}
		}
		
		//Generate the Tabletop Simulator save file.
		var amount string = "100"
		fmt.Println("generated ",total," cards")
		for i:=101; i < 100+total; i++ {
			amount += ",\n        "+fmt.Sprint(i)
		} 
		
		//It is json.
		json := Template
		json = strings.Replace(json, "{{ URL1 }}", "http://"+ip_address+":20002/"+filepath.Base(filename)+".jpg", 1)
		json = strings.Replace(json, "{{ #Cards }}", amount, 1)
		json = strings.Replace(json, "{{ URL2 }}", "http://"+ip_address+":20002/"+game+".jpg", 1)
		
		//Write file to disk.
		handle(ioutil.WriteFile(chest+"/"+filepath.Base(filename)+".json", []byte(json), 0644))

		//Yay we did it!
		ct.ChangeColor(ct.Green, true, ct.None, false)
		fmt.Print("Done ")
		ct.ResetColor()
		fmt.Println(filename + "!")

	} else {
		//The file you provided doesn't seem to exist or something.
		fmt.Println(err.Error())
		//Always helps to insult the user of their spelling,
		//It makes them feel better.
		fmt.Println("Check spelling?")
		return
	}
}

//Concurrency things.
var wg sync.WaitGroup
var threading bool

//Where the cache at.
var cache string

//Where the Tabletop Chest directory is.
var chest string 
var ip_address string = "localhost"


func walker(path string, info os.FileInfo, err error) error {
	if err != nil {
		return nil
	}
	
	if info.Mode().IsDir() {
		files, err := filepath.Glob(path+"/*.deck")
		handle(err)
		for _, file := range files {
			wg.Add(1)
			go decker(file)
		}
	}
	
	return filepath.SkipDir
}

//This will serve decks to other players in Tabletop simulator.
//This should hopefully just "work"
//Not tested over the internet yet...
func host() {

	file_server := http.FileServer(http.Dir(cache+"/images/"))
	

	fmt.Println(http.ListenAndServe(":20002", 
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/?test" {
				w.Write([]byte("active"))
			}
			urlpath := r.URL.Path
			if len(urlpath) > 3 && urlpath[:3] == "/ip" {
				urlpath = urlpath[3:]
				
				if urlpath[len(urlpath)-1] == '/' {
					r.URL.Path = "/"
				} else {
					r.URL.Path = path.Base(r.URL.Path)
				}
				
				if path.Dir(urlpath)[1:] == ip_address {
					file_server.ServeHTTP(w, r)
				} else {
					proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme:"http",Host:ip_address+":20002"})
					proxy.ServeHTTP(w, r)
				}
			} else {
				file_server.ServeHTTP(w, r)
			}
	})))
}

//This will format the .json saves to IP or localhost.
func TabletopSetLocal(b bool) {
 	files, _ := ioutil.ReadDir(cache+"/images/")
    for _, f := range files {
    	file, err := os.Open(chest+"/"+f.Name()[:len(f.Name())-4]+".json")
    	if err == nil {
     	   	data, err := ioutil.ReadAll(file)
			if err == nil {
				if b {
					data = []byte(strings.Replace(string(data), ip_address, "localhost", -1))
				} else {
					data = []byte(strings.Replace(string(data), "localhost", ip_address, -1))
				}
				ioutil.WriteFile(chest+"/"+f.Name()[:len(f.Name())-4]+".json", data, 0644)
			}
        }
    }
}

func main() {

	//Figure out where we gonna put our cache.
	//If for some reason we can't write to these directories, we're screwed... BUG?
	cache = os.Getenv("HOME") + "/.cache/decker"
	chest = os.Getenv("HOME") + "/Documents/My Games/Tabletop Simulator/Saves/Chest"

	if runtime.GOOS == "windows" {
		cache = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if cache == "" {
			cache = os.Getenv("USERPROFILE")
		}
		chest = cache + "/Documents/My Games/Tabletop Simulator/Saves/Chest"
		cache += "/AppData/Roaming/decker"
	}
	
	plugins.DeckerCachePath = cache
	for _, v := range plugins.Plugins() {
		//Create a cache folder for the games.
		//Unfortunately we also get a empty "none" folder.
		//I'm not going to fix this because it could be useful in the future somehow.
		if _, err := os.Stat(cache + "/cards/" + v.Game + "/"); os.IsNotExist(err) {
			handle(os.MkdirAll(cache+"/cards/"+v.Game+"/", os.ModePerm))
		}
	}
	
	if _, err := os.Stat(cache + "/images/"); os.IsNotExist(err) {
		handle(os.MkdirAll(cache + "/images/", os.ModePerm))
	}
	if _, err := os.Stat(cache + "/decks/"); os.IsNotExist(err) {
		handle(os.MkdirAll(cache + "/decks/", os.ModePerm))
	}

	//Parse the commandline arguments.
	flag.Parse()

	//Print a very helpful usage message that everybody understands.
	if flag.Arg(0) == "" {
	
		//Grab our IP address, if able.
		ip_cache, err := os.Open(cache+"/ip")
		if err == nil {
			data, err := ioutil.ReadAll(ip_cache)
			if err == nil {
				ip_address = string(data)
			}
		}
		if ip_address == "localhost" {
			response, err := client.Get("http://myexternalip.com/raw")
			if err == nil {
				data, err := ioutil.ReadAll(response.Body)
				if err == nil {
					ip_address = strings.TrimSpace(string(data))
				}
			}
			//Cache it.
			handle(ioutil.WriteFile(cache+"/ip", []byte(ip_address), 0644))
		}
	
		go host()
		
		//Display some nice information that we are hosting files needed for Tabletop Simulator.
		fmt.Println("We are now hosting the decks so people can download them from your computer..\n")
		fmt.Println("IP Address:",ip_address)
		fmt.Print("Port forwarding:")
		_, err = client.Get("http://"+ip_address+":20002/?test")
		if err == nil {
			ct.ChangeColor(ct.Green, true, ct.None, false)
			fmt.Println(" Enabled")
			ct.ResetColor()
			TabletopSetLocal(false)
		} else {
			ct.ChangeColor(ct.Red, true, ct.None, false)
			fmt.Println(" Disabled")
			ct.ResetColor()
			fmt.Println("Please port forward 20002 to your computer.")
			TabletopSetLocal(true)
		}
		fmt.Println("\nIf the above IP address is wrong or you wish to use Decker over LAN please edit the file:")
		fmt.Println(cache+"/ip")
		fmt.Println("Change the contents to the IP/Domain that your friends connect to you by.")
		goto end
	}
	
	//Display License information.
	fmt.Println("Decker version 0.7, Copyright (C) 2014 Quentin Quaadgras")
	fmt.Println("Decker comes with ABSOLUTELY NO WARRANTY!")
	fmt.Println("This is free software, and you are welcome to redistribute it")
	fmt.Println("under certain conditions;")
	fmt.Println("visit http://www.gnu.org/licenses/gpl-2.0.txt for details.\n")

	//How many decks do we need to create sir?
	//Only one? are you sure you don't want to bulk generate decks?
	//terrible shame.
	if len(flag.Args()) > 1 {
		threading = true
		for _, v := range flag.Args() {
			if info, err := os.Stat(v); !os.IsNotExist(err) {
				if info.Mode().IsDir() {
					filepath.Walk(v, walker)
				} else {
					wg.Add(1)
					go decker(v)
				}
			}
		}
	} else {
		if info, err := os.Stat(flag.Arg(0)); !os.IsNotExist(err) {
			if info.Mode().IsDir() {
				threading = true
				filepath.Walk(flag.Arg(0), walker)
			} else {
				decker(flag.Arg(0))
			}
		}
	}

	//Wait for everybody to finish.
	wg.Wait()

	end:
	fmt.Println("Press 'Enter' to close...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}
