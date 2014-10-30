package main

import (
	"errors"
	"os"
	"image"
	"image/jpeg"
	_ "image/png"
)
import "./cutter"

//Added for decker to solve the "black cards bug".
func CropDeck(deck string) error {
	fi, err := os.Open(deck)
	if err != nil {
		return errors.New("Cannot open file '"+deck+"':"+err.Error())
	}

	img, _, err := image.Decode(fi)
	if err != nil {
		return errors.New("Cannot decode image at '"+deck+"':"+err.Error())
	}

	cImg, err := cutter.Crop(img, cutter.Config{
		Height:  4096,                  // height in pixel or Y ratio(see Ratio Option below)
		Width:   4096,                  // width in pixel or X ratio
		Mode:    cutter.TopLeft,        // Accepted Mode: TopLeft, Centered
		Anchor:  image.Point{0, 0}, // Position of the top left point
		Options: 0,                     // Accepted Option: Ratio
	})
	if err != nil {
		return errors.New("Cannot crop image:"+err.Error())
	}
	
	fi.Close()
	fo, err := os.Create(deck)
	if err != nil {
		return errors.New("Cannot modify file '"+deck+"':"+err.Error())
	}
	defer fo.Close()

	err = jpeg.Encode(fo, cImg, &jpeg.Options{Quality:60})
	if err != nil {
		return err
	}
	//fmt.Println("Image cropped to power of 2", deck)
	return nil
}
