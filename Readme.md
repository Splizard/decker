Decker!
---------
A tool for creating trading card decks for Table Top Simulator.

At the moment, it only supports magic and pokemon decks.
Have a look at the provided example decks on how to format a deck.

### Build dependencies ###

To build Decker you'll need to have the Go binaries installed.
Decker also depends on imagemagick you'll also need to have that installed.
To install these dependencies, see below.

*Ubuntu:*

Run this command to install Go and imagemagick.

    $ sudo apt-get install golang imagemagick

*Windows:*

Download the installer from [here](https://golang.org/dl/).  
You also need the imagemagick installed, see [here](http://www.imagemagick.org/script/download.php#windows)

*Other:*
If you are running a different operating system you will need to get Go and imagemagick.

### Build instructions ###

To build on ubuntu/linux, run:

    $ make && sudo make install

Or for windows:

    > go build -o ./decker.exe ./src

### Example decks ###

Some example decks are contained in the decks folder.  
Have a look at those if you are confused to how a deck file should look like.

    $ decker ./decks/magic
    $ decker ./decks/pokemon
    $ decker ./decks/yugioh
    $ decker ./decks/custom
    $ decker ./decks/text
    
This will generate the decks into images in the respective folder.

### Usage ###
    
Commandline Usage:
	#Generate a deck
    $ decker file.deck
    
    #Run as a server
   	$ decker

GUI Usage:  
Decker has limited use as a GUI.  
Maybe you want to write a front-end for it on your platform?

When you use Decker to generate a .deck file, Decker will ouput the image with a .jpg suffix  
eg. file.deck >> file.deck.jpg

If you have Tabletop Simulator installed, it will place the deck in your chest.  
To use the decks you will either need to have Decker running in the background or 
have the image hosted online which then decker can upload it for you too.  
If you don't upload the image online then to allow people who connect to your game to use the decks you will need to port forward 20002. 

*Ubuntu:*

Decker will be made the default program to open .deck files.  
Simply double click on a .deck file to generate it.  
Run Decker from the dash before you play Tabletop Simulator and keep it running while you play.  
You can also host the image on Imgur.com right click on the generated image and open it with decker, 
this will upload the image so that when you run Tabletop you don't need Decker running.

*Windows:*

Drag decks onto decker.exe  
Run decker.exe before you play Tabletop Simulator and keep it running while you play.
You can also host the image on Imgur.com drag the generated image onto decker.exe, 
this will upload the image so that when you run Tabletop you don't need Decker running.
