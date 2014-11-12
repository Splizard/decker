package plugins

import "os"
import "fmt"
import "strings"
import svg "../svgo"

const Text = "text"

func init() {
	var counter int64

	RegisterHeaders(Text, []string{"Text", "Superfight"})
		
	RegisterCacheImageName(Text, func(name string) string {
		counter++
		imageOut, err := os.Create(DeckerCachePath + "/cards/text/cache-"+fmt.Sprint(counter)+".jpg")
		Handle(err)
		
		lines := strings.Split(name, " ")
		
		width, height := 480, 680
		canvas := svg.New(imageOut)
		canvas.Start(width, height)
		canvas.Rect(0, 0, width, height, "fill:white")
		for i, v := range lines {
			canvas.Text(width/2, height/(1+len(lines))+i*80, v, "word-wrap:true;text-anchor:middle;font-size:80px;fill:black")
		}
		canvas.End()
		return "cache-"+fmt.Sprint(counter)
	})

	RegisterPlugin(Text, func(name, info string, detecting bool) string {
		return None
	})
}
