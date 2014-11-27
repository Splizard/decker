package main

//This is a lazy quick way of saving custom decks to a file.
const Template = `
{
  "SaveName": "",
  "GameMode": "",
  "Date": "",
  "Table": "",
  "Sky": "",
  "Note": "",
  "Rules": "",
  "PlayerTurn": "",
  "ObjectStates": [
    {
      "Name": "DeckCustom",
      "Transform": {
        "posX": 0,
        "posY": 0,
        "posZ": 0,
        "rotX": 0,
        "rotY": 180.0,
        "rotZ": 180.0,
        "scaleX": 1.0,
        "scaleY": 1.0,
        "scaleZ": 1.0
      },
      "Nickname": "",
      "Description": "",
      "ColorDiffuse": {
        "r": 0.713239133,
        "g": 0.713239133,
        "b": 0.713239133
      },
      "Grid": true,
      "Locked": false,
      "SidewaysCard": false,
      "DeckIDs": [
        {{ #Cards }}
        
      ],
      "CustomDeck": {
        "1": {
          "FaceURL": "{{ URL1 }}",
          "BackURL": "{{ URL2 }}"
        }
      }
    }
  ]
}`
