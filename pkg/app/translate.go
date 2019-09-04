package app

var translateTpl map[string]map[string]string
var _ = Config().Bind("translate", "", &translateTpl)

//var _, _ = toml.DecodeFile(Root()+"/configs/translate.toml", &translateTpl)

func Translate(lang, message string) string {
	if val, ok := translateTpl[message][lang]; ok {
		return val
	}
	return message
}
