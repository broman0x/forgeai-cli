package lang

var currentLang = "en"

func SetLanguage(lang string) {
	if lang == "id" || lang == "en" {
		currentLang = lang
	}
}

func GetLanguage() string {
	return currentLang
}

func T(key string) string {
	if currentLang == "id" {
		return ID[key]
	}
	return EN[key]
}
