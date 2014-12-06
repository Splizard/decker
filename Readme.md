Decker!
---------
A tool for creating trading card decks for Table Top Simulator.

At the moment, it only supports magic and pokemon decks.
Have a look at the provided example decks on how to format a deck.

### Build dependencies ###

To build decker you'll need to have the Go binaries installed.
Decker also depends on imagemagick you'll also need to have that installed.
To install these dependencies, see below.

*Linux:*

For Go:

    $ sudo apt-get install golang
or get the installer from [here](https://golang.org/dl/).

For imagemagick:

    $ sudo apt-get install golang imagemagick

*Windows:*

Download the installer from [here](https://golang.org/dl/).

You also need the imagemagick installed, see [here](http://www.imagemagick.org/script/binary-releases.php)

### Build instructions ###

To build on linux, run:

    $ make && sudo make install

Or for windows:

    > go build -o ./decker.exe ./src

### Example decks ###

Some example decks are contained in the decks folder.
Have a look at those if you are confused to how a deck file should look like.

    $ cd ./decks/magic
    $ decker *.deck
    $ cd ./decks/pokemon
    $ decker *.deck
    
This will generate the decks into images in the respective folder.
    
Usage:

    $ decker -o deck.jpg deck.txt

This will generate an image called deck.jpg.
Put this image somewhere online and it can be downloaded by Table Top Simulator!
