package plugins

import (
	"fmt"
	"os"
	"strings"

	svg "github.com/Splizard/decker/src/svgo"
)

const Text = "text"

func init() {
	var counter int64

	RegisterHeaders(Text, []string{"Text", "Superfight"})

	RegisterCacheImageName(Text, func(name string) string {

		var size = 80

		length := func(text string) int {
			return (len(text)) * ((size / 2) + size/6)
		}

		resize_text := func(width int, text string) (string, string, bool) {
			overlap := float32(width) / float32(length(text))
			scale := overlap * float32(len(text))

			// Replace space with new line characters
			text = strings.Replace(text, "\\n", "\n", -1)

			//Check for spaces.
			for i := int(scale); i > 0; i-- {
				if text[i] == ' ' {
					scale = float32(i)
					text = text[:i] + text[i+1:]
					break
				}
			}

			if length(text[int(scale):]) > int(width) {
				return text[:int(scale)], text[int(scale):], false
			} else {
				return text[:int(scale)], text[int(scale):], true
			}
		}

		counter++
		imageOut, err := os.Create(DeckerCachePath + "/cards/text/cache-" + fmt.Sprint(counter) + ".jpg")
		Handle(err)

		width, height := 480, 680
		x, y := width/2, height/4
		canvas := svg.New(imageOut)
		canvas.Start(width, height)
		canvas.Rect(0, 0, width, height, "fill:white")

		//align := func(str string) int { return (width-len(str)*(size/2))/2 }

		for _, v := range strings.Split(name, " ") {
			for length(v) > 480 {
				size -= 1
			}
		}

		for length(name) > 480*8 {
			size -= 1
		}

		textwidth := length(name)
		if textwidth > width {
			yy := 0
			for {
				str1, str2, err := resize_text(int(width), name)
				canvas.Text(x, y+yy, str1, "alignment-baseline:central;text-anchor:middle;font-size:"+fmt.Sprint(size)+"px;fill:black;font-family:monospace")
				name = str2
				yy += size
				if err {
					canvas.Text(x, y+yy, str2, "alignment-baseline:central;text-anchor:middle;font-size:"+fmt.Sprint(size)+"px;fill:black;font-family:monospace")
					break
				}
			}

		} else {
			canvas.Text(x, y, name, "alignment-baseline:central;text-anchor:middle;font-size:"+fmt.Sprint(size)+"px;fill:black;font-family:monospace")
		}

		canvas.End()
		return "cache-" + fmt.Sprint(counter)
	})

	RegisterPlugin(Text, func(name, info string, detecting bool) string {
		return None
	})
}
