/*
	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.

	Copyright (C) 2014 Quentin Quaadgras
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
	"runtime"
	"strings"
	"sync"
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

func init() {
	flag.StringVar(&output, "o", "deck.jpg", "output file")
}

//Define the current card games decker supports.
const (
	None    = "none"
)

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
			}
		}
	}()
	//Leave the wait group.
	if threading {
		defer wg.Done()
	}

	var name string      //The name of the card.
	var info string      //Extra details to identify the card, mainly for Pokemon.

	var game string = None //The current game as defined by the game constants.

	//output file, this is to keep track of per-file outputs when running in parallel.
	//we want to output to png.
	var output = output
	if threading {
		output = filepath.Base(filename) + ".jpg"
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
			if len(line) > 2 && (((line[1] == 'x' || line[2] == 'x') && (line[0] > 48 && line[0] < 58)) ||
				(line[0] > 48 && line[0] < 58) ||
				(line[0] == 'x' && line[1] > 48 && line[1] < 58)) {
				
				//We need to seperate the name from the number of cards.
				//This does that.
				if line[1] == 'x' {
					name = strings.TrimSpace(line[2:])
				} else if line[0] == 'x' || line[2] == 'x' {
					name = strings.TrimSpace(line[3:])
				} else {
					name = strings.TrimSpace(line[2:])
				}

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
				
				imagename := name
				if i := plugins.GetImageName(game, name); i != "" {
					imagename = i
				}
				
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

					//Let's check if the card we are looking for has already been downloaded.
					//Plugins may handle this by themselves.
					if _, err := os.Stat(cache + "/cards/" + game + "/" + imagename + ".jpg"); info == "" && !os.IsNotExist(err) {
						if !usingCache {
							fmt.Println("using cached files for "+filename)
							usingCache = true
						}
					} else {
						plugins.Run(game, name, info)
					}

					//If the imagename is different from the card name, we replace it now so everything works.
					if plugins.GetImageName(game, name) != "" {
						name = plugins.GetImageName(game, name)
					}

					//Copy the card from cache to the temp directory.
					if _, err := os.Stat(temp + "/" + name + ".jpg"); os.IsNotExist(err) {
						_, err := Copy(cache+"/cards/"+game+"/"+name+".jpg", temp+"/"+name+".jpg")
						handle(err)
					}

					//Figure out how many cards there are in the deck.
					//Maximum is 99 otherwise unpredictable things will happen.
					//Should probably note this somewhere.

					//More complicated code that just works.
					var tens int
					var ones int

					//For in the style of:
					//
					//	1x Card Name
					//  1  Card Name
					//
					if line[0] != 'x' {
						if line[1] > 47 && line[1] < 58 {
							tens = int(line[0] - 48)
							ones = int(line[1] - 48)
						} else {
							ones = int(line[0] - 48)
						}

						//For in the style of:
						//
						//	x1 Card Name
						//
					} else if line[0] == 'x' {
						if line[2] > 47 && line[2] < 58 {
							if line[1] > 47 && line[1] < 58 {
								tens = int(line[1] - 48)
							}
							ones = int(line[2] - 48)
						} else {
							ones = int(line[1] - 48)
						}
					}

					//Create copies of the card in the temporary directory.
					for i := 1; i < tens*10+ones; i++ {

						if _, err := os.Stat(temp + "/" + name + " " + fmt.Sprint(i+1) + ".jpg"); os.IsNotExist(err) {

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
		err := montage.Run()
		handle(err)
		
		//Crop the deck to a power of 2, 4096x4096 this will overwrite the file as a compressed jpeg.
		err = CropDeck(output)
		handle(err)

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

func main() {

	//Figure out where we gonna put our cache.
	//If for some reason we can't write to these directories, we're screwed... BUG?
	cache = os.Getenv("HOME") + "/.cache/decker"

	if runtime.GOOS == "windows" {
		cache = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if cache == "" {
			cache = os.Getenv("USERPROFILE")
		}
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

	//Parse the commandline arguments.
	flag.Parse()

	//Print a very helpful usage message that everybody understands.
	if flag.Arg(0) == "" {
		fmt.Println("usage: decker [OPTIONS] [FILE]")
		return
	}

	//How many decks do we need to create sir?
	//Only one? are you sure you don't want to bulk generate decks?
	//terrible shame.
	if len(flag.Args()) > 1 {
		threading = true
		for _, v := range flag.Args() {
			wg.Add(1)
			go decker(v)
		}
	} else {
		decker(flag.Arg(0))
	}

	//Wait for everybody to finish.
	wg.Wait()

	//On windows people don't use a command line so we better give them a chance to read any error messages :3
	if runtime.GOOS == "windows" {
		fmt.Println("Press 'Enter' to close...")
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	}
}
