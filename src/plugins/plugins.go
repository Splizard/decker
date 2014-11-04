//This package contains the API and glue which let's the plugins work.
package plugins

import "sync"

const (
	None = "none"
)

//A plugin contains it's card getting function and it's ID or Game
//and A map of ImageNames if they change.
type Plugin struct {
	Game string
	Function func(string, string, bool) string
	ImageNames map[string]string
	ImageFunction func(string) string
	Mutex *sync.Mutex
}

func init() {
	RegisterPlugin(None, func(string, string, bool) string { return None })
}

var DeckerCachePath string

var _plugins map[string]*Plugin = make(map[string]*Plugin)
var _headers map[string]string = make(map[string]string)
var _cachers map[string]func(string) string = make(map[string]func(string) string)

func RegisterPlugin(game string, function func(string, string, bool) string) {
	_plugins[game] = &Plugin{Game:game,Function:function,ImageNames:make(map[string]string), Mutex:new(sync.Mutex)}
}

func RegisterHeaders(game string, headers []string) {
	for _, v := range headers {
		_headers[v] = game
	}
}

func RegisterCacheImageName(game string, function func(string) string) {
	_cachers[game] = function
}

func Identify(line string) string {
	for i, v := range _headers {
		if i == line {
			return v
		}
	}
	return None
}

//Error handler, all bad errors will be sent here.
func Handle(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func Plugins() map[string]*Plugin {
	return _plugins
}

func SetImageName(game, name, imagename string) {
	_plugins[game].Mutex.Lock()
	_plugins[game].ImageNames[name] = imagename
	_plugins[game].Mutex.Unlock()
}

func GetImageName(game, name string) string {
	_plugins[game].Mutex.Lock()
	defer _plugins[game].Mutex.Unlock()
	key, yes := _plugins[game].ImageNames[name]
	if yes {
		return key
	}
	f, yes := _cachers[game]
	if yes {
		return f(name)
	}
	return ""
}

func Autodetect(name, input string) []string {
	identified := []string{}
	for _, v := range _plugins {
		game := v.Function(name, input, true)
		if game != None {
			identified = append(identified, game)
		}
	}
	return identified
}

func Run(game, name, input string) string {
	identified := game
	for i, v := range _plugins {
		if game == i {
			identified = v.Function(name, input, false)
		} else if game == None {
			identified = v.Function(name, input, true)
		}
		if game == None && identified != game {
			return identified
		}
	}
	return identified
}
