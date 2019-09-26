package app

var translateTpl map[string]map[string]string
var _ = Config().Bind("translate", "", &translateTpl)

//var _, _ = toml.DecodeFile(Root()+"/configs/translate.toml", &translateTpl)

// Translate 增加了 i18n 模块之后不再外部可以不使用这个方法了
func Translate(lang, message string) string {
	if val, ok := translateTpl[message][lang]; ok {
		return val
	}
	return message
}
