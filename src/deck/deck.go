package deck

import (
	"io"
	"bufio"
	"regexp"
	"strings"
	"strconv"
		
	"sort" //Because we are organised.
)

type Deck struct {
	Game string
	
	Cards []string
	Copies map[string]int
}

func Load(source io.Reader) (Deck, error) {
	//Read the first line and trim the space.
	reader := bufio.NewReader(source)
	line, err := reader.ReadString('\n')
	if err != nil {
		return Deck{}, err
	}
	line = strings.TrimSpace(line)
	
	var deck Deck
	deck.Game = line
	deck.Copies = make(map[string]int)
	
	var name string //Card name.
	
	//Loop through the file.
	for {
		line, err := reader.ReadString('\n') //Parse line-by-line.
		if err == io.EOF {
			break
		} else if err != nil {
			return Deck{}, err
		}

		//Trim the spacing. TODO trim spacing in between words that are used for nice reading.
		line = strings.TrimSpace(line)
		
		//Cards are identified by having an 'x' at the beginning of the line or a number.
		//Anyother character is a comment.
		//Not many words start with x so we should be pretty safe, let's not worry about dealing with special cases.

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
			
			//Get the count of cards by getting the xn, nx or n part and replacing the x
			count, _ := strconv.Atoi(strings.Replace(r.FindString(line), "x", "", -1));
			
			//Add this card to the deck.
			deck.Cards = append(deck.Cards, name)
			deck.Copies[name] += count
		}
	}
	
	//Sort our cardnames into alphabeta order.
	sort.Sort(sort.StringSlice(deck.Cards))
	
	return deck, nil
}
