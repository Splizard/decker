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
	 "net/http"
	"os"
	"io"
	"io/ioutil"
	"bufio"
	"strings"
	"fmt"
	"flag"
	"os/exec"
	"path/filepath"
	"sync"
	"errors"
	"runtime"
	"regexp"
	"net/url"
)

import "./ct"

func handle(err error) {
	if err != nil {
		panic(err.Error())
	}
}

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
var output string

func init() {
	flag.StringVar(&output, "o", "deck.jpg", "output file")
}

const (
	None = ""
	Magic = "magic"
	Pokemon = "pokemon"
)

var pokemonregex *regexp.Regexp
var pokemonimageregex * regexp.Regexp

func decker(filename string) {
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
                            fmt.Println(filename+"!")
                            return
                    }
            }
    }()
    if threading {
   	 	defer wg.Done()
   	}


	var name string
	var imagename string //This could different to the card name eg. Pokemon.
	var info string
	
	var client http.Client
	var temp string
	
	var game string = None
	
	var output = output
	
	if threading {
		output = filepath.Base(filename)+".jpg"
	}
	
	var usingCache bool
	
	temp, err := ioutil.TempDir("", "decker")
	handle(err)
	defer os.Remove(temp)
	

	if file, err := os.Open(filename); err == nil {
		reader := bufio.NewReader(file)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		
		//Some friendly identifiers.
		if line == "Magic: The Gathering" || line == "Magic" || line == "MTG" {
			game = Magic
		}
		if line == "Pokemon" || line == "Pokemon Trading Card Game" || line == "Pokémon Trading Card Game" {
			game = Pokemon
			pokemonregex = regexp.MustCompile(`http://pkmncards\.com/card/(.[0-9a-zA-z _\.\-:]*)/`)
			pokemonimageregex = regexp.MustCompile(`"og:image"\scontent="([0-9a-zA-z \/_\.\-:]*)`)
		}
		if game == None {
			handle(errors.New("No game found!"))
		}
		
		for {
			line, err := reader.ReadString('\n') // parse line-by-line
			if err == io.EOF {
				if len(line) == 0 {
					break
				}
			}
			handle(err)
			line = strings.TrimSpace(line)
			
			//Download images and cache.
			
			if len(line) > 2 && (((line[1] == 'x' || line[2] == 'x') && (line[0] > 48 && line[0] < 58)) || 
								(line[0] > 48 && line[1] < 58 && line[0] > 48 && line[1] < 58) ||
								(line[0] == 'x' && line[1] > 48 && line[1] < 58)) {
			
				if line[1] == 'x' {
					name = strings.TrimSpace(line[2:])
				} else if line[0] == 'x' || line[2] == 'x' {
					name = strings.TrimSpace(line[3:])
				} else {
					name = strings.TrimSpace(line[2:])
				}
				
				imagename = ""
				
				//Info is so cards can be identified easier.
				//Such is that in pokemon where lots of cards have the same name.
				info = ""
				if game == Pokemon {
					if strings.Contains(name, ",") {
						splits := strings.Split(name, ",")
						name = splits[0]
						info = strings.TrimSpace(splits[1])
					}
					if strings.Contains(name, " with ") {
						splits := strings.Split(name, " with ")
						name = splits[0]
						info =strings.TrimSpace(splits[1])
					}
					if strings.Contains(name, " that has ") {
						splits := strings.Split(name, " that has ")
						name = splits[0]
						info = strings.TrimSpace(splits[1])
					}
				}
				
				_, err := os.Stat(cache+"/cards/"+game+"/"+name+".jpg")
				
				if !os.IsNotExist(err) {
					if os.IsNotExist(err) {
						if !usingCache {
							fmt.Println("using cached files")
							usingCache = true
						}
					}
				}
				
				if os.IsNotExist(err) {
					if _, err := os.Stat(cache+"/cards/"+game+"/"); os.IsNotExist(err) {
						handle(os.MkdirAll(cache+"/cards/"+game+"/", os.ModePerm))
					}
					
					if game == Magic {
				
						println("getting", "http://mtgimage.com/card/"+name+".jpg")
						response, err := client.Get("http://mtgimage.com/card/"+name+".jpg")
						handle(err)
						if response.StatusCode == 404 {
							handle(errors.New("card name '"+ name +"' invalid!"))
						} else {
							if response.StatusCode != 200 {
								println("possible error check file! "+ name+ ", status "+response.Status)
							}
							imageOut, err := os.Create(cache+"/cards/magic/"+name+".jpg")
							handle(err)
							io.Copy(imageOut, response.Body)
							imageOut.Close()
						}
						
					} else if game == Pokemon {
						//Format url, pkmncards.com does not like an empty text:"" field.
						var search string
						if info != "" {
							search = "http://pkmncards.com/?s="+url.QueryEscape(name)+"+text%3A%22"+url.QueryEscape(info)+"%22&display=scan&sort=date"
						} else {
							search = "http://pkmncards.com/?s="+url.QueryEscape(name)+"%22&display=scan&sort=date"
						}
					
						response, err := client.Get(search)
						handle(err)
						if response.StatusCode == 404 {
							handle(errors.New("card name '"+ name +"' invalid!"))
						} else if response.StatusCode != 200 {
							println("possible error check file! "+ name+ ", status "+response.Status)
						}
						
						body, err := ioutil.ReadAll(response.Body)
						handle(err)

						card := string(pokemonregex.Find([]byte(body)))
						
						if card == "" {
							handle(errors.New("card name '"+ name +"' not found!\nCheck "+search))
						}
						
						//Get actual card.
						response, err = client.Get(card)
						body, err = ioutil.ReadAll(response.Body)
						handle(err)
						submatches := pokemonimageregex.FindStringSubmatch(string(body))
						if len(submatches) < 2 {
							handle(errors.New("No image found for card "+name+", this could be a bug !"))
						}
						image := string(submatches[1])
						
						path, err := url.Parse(image)
						handle(err)
						imagename = strings.Replace(filepath.Base(path.Path), ".jpg", "", 1)
						
						if _, err := os.Stat(cache+"/cards/"+game+"/"+imagename+".jpg"); !os.IsNotExist(err) {
							if !usingCache {
								fmt.Println("using cached files")
								usingCache = true
							}
						} else {
						
							fmt.Println("getting", image)
							response, err = client.Get(image)
							handle(err)
							if response.StatusCode == 404 {
								handle(errors.New("card name '"+ name +"' invalid!"))
							} else {
								if response.StatusCode != 200 {
									println("possible error check file! "+ name+ ", status "+response.Status)
								}
								imageOut, err := os.Create(cache+"/cards/pokemon/"+imagename+".jpg")
								handle(err)
								io.Copy(imageOut, response.Body)
								imageOut.Close()
							}
						}
					}
					
				}
				
				//Create deck.
				
				if imagename != "" {
					name = imagename
				}
				if _, err := os.Stat(temp+"/"+name+".jpg"); os.IsNotExist(err) {
					_, err := Copy(cache+"/cards/"+game+"/"+name+".jpg", temp+"/"+name+".jpg")
					handle(err)
				}
			
				
				var tens int
				var ones int
				
				// 1x Card Name
				if line[0] != 'x' {
					if (line[1] > 48 && line[1] < 58) {
						tens = int(line[0] - 48)
						ones = int(line[1] - 48)
					} else {
						ones = int(line[0] - 48)
					}
					
				// x1 Card Name	
				} else if line[0] == 'x' {
					if line[2] > 48 && line[2] < 58 {
						if (line[1] > 48 && line[1] < 58) {
							tens = int(line[1] - 48)
						}
						ones = int(line[2] - 48)
					} else {
						ones = int(line[1] - 48)
					}
				}
				for i := 1; i < tens*10+ones; i++ {
					if _, err := os.Stat(temp+"/"+name+" "+fmt.Sprint(i+1)+".jpg"); os.IsNotExist(err) {
						if runtime.GOOS == "windows" {
							Copy(cache+"/cards/magic/"+name+".jpg", temp+"/"+name+" "+fmt.Sprint(i+1)+".jpg")
						} else {
							os.Symlink("./"+name+".jpg", temp+"/"+name+" "+fmt.Sprint(i+1)+".jpg")
						}
						
					}
				}
			}
		}
		
		fmt.Println("Generating image for "+filename+"...")
		montage := exec.Command("montage", "-background", "rgb(23,20,15)", "-tile", "10x7", "-quality", "60", "-geometry", "409x585!+0+0", temp+"/*.jpg", output)
		montage.Run()
		ct.ChangeColor(ct.Green, true, ct.None, false)
		fmt.Print("Done ")
		ct.ResetColor()
		fmt.Println(filename+"!")
	} else {
		fmt.Println(err.Error())
		return
	}
}

var wg sync.WaitGroup
var threading bool

var cache string

func main() {
	cache = os.Getenv("HOME")+"/.cache/decker"
	
	if runtime.GOOS == "windows" {
		cache = os.Getenv("HOMEDRIVE")+os.Getenv("HOMEPATH")
		if cache == "" {
			cache = os.Getenv("USERPROFILE")
		}
		cache += "/AppData/Roaming/decker"
	}
	
	flag.Parse()

	if flag.Arg(0) == "" {
		fmt.Println("usage: decker [OPTIONS] [FILE]")
		return
	}
	
	if len(flag.Args()) > 1 {
		threading = true
		for _, v := range flag.Args() {
			wg.Add(1)
			go decker(v)
		}
	} else {
		decker(flag.Arg(0))
	}
	
	wg.Wait()
	if runtime.GOOS == "windows" {
		fmt.Println("Press 'Enter' to close...")
		reader := bufio.NewReader(os.Stdin)
   		reader.ReadString('\n')
	}
}
