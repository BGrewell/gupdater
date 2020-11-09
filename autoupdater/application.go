package autoupdater

type Application struct {
	Name        string `yaml:"name"`
	UpdateUrl   string `yaml:"update_url"`
	LocalDir    string `yaml:"local_dir"`
	VersionFile string `yaml:"version_file"`
	PackageName string `yaml:"package_name"`
	StopCmd     string `yaml:"stop_cmd"`
	UpdateCmd   string `yaml:"update_cmd"'`
	StartCmd    string `yaml:"start_cmd"`
}
