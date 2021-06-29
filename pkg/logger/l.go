package logger

// Config is generic struct for logging configs.
type Config struct {
	Enabled             bool   `default:"true" split_words:"true"`
	Level               string `default:"info" split_words:"true"`
	Pretty              bool   `default:"false" split_words:"true"`
	FileRedirectEnabled bool   `default:"false" split_words:"true"`
	FileRedirectPath    string `default:"/var/log" split_words:"true"`
	FileRedirectPrefix  string `default:"detective" split_words:"true"`
}
