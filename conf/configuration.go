package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/navidrome/navidrome/consts"
	"github.com/navidrome/navidrome/log"
	"github.com/spf13/viper"
)

type configOptions struct {
	ConfigFile              string
	Address                 string
	Port                    int
	MusicFolder             string
	DataFolder              string
	DbPath                  string
	LogLevel                string
	ScanInterval            time.Duration
	SessionTimeout          time.Duration
	BaseURL                 string
	UILoginBackgroundURL    string
	EnableTranscodingConfig bool
	EnableDownloads         bool
	TranscodingCacheSize    string
	ImageCacheSize          string
	AutoImportPlaylists     bool

	SearchFullString       bool
	RecentlyAddedByModTime bool
	IgnoredArticles        string
	IndexGroups            string
	ProbeCommand           string
	CoverArtPriority       string
	CoverJpegQuality       int
	UIWelcomeMessage       string
	EnableGravatar         bool
	EnableFavourites       bool
	EnableStarRating       bool
	EnableUserEditing      bool
	DefaultTheme           string
	GATrackingID           string
	EnableLogRedacting     bool
	AuthRequestLimit       int
	AuthWindowLength       time.Duration

	Scanner scannerOptions

	Agents  string
	LastFM  lastfmOptions
	Spotify spotifyOptions

	// DevFlags. These are used to enable/disable debugging and incomplete features
	DevLogSourceLine           bool
	DevAutoCreateAdminPassword string
	DevPreCacheAlbumArtwork    bool
	DevFastAccessCoverArt      bool
	DevOldCacheLayout          bool
	DevActivityPanel           bool
}

type scannerOptions struct {
	Extractor string
}

type lastfmOptions struct {
	ApiKey   string
	Secret   string
	Language string
}

type spotifyOptions struct {
	ID     string
	Secret string
}

var (
	Server = &configOptions{}
	hooks  []func()
)

func LoadFromFile(confFile string) {
	viper.SetConfigFile(confFile)
	Load()
}

func Load() {
	err := viper.Unmarshal(&Server)
	if err != nil {
		fmt.Println("Error parsing config:", err)
		os.Exit(1)
	}
	err = os.MkdirAll(Server.DataFolder, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating data path:", "path", Server.DataFolder, err)
		os.Exit(1)
	}
	Server.ConfigFile = viper.GetViper().ConfigFileUsed()
	if Server.DbPath == "" {
		Server.DbPath = filepath.Join(Server.DataFolder, consts.DefaultDbPath)
	}

	log.SetLevelString(Server.LogLevel)
	log.SetLogSourceLine(Server.DevLogSourceLine)
	log.SetRedacting(Server.EnableLogRedacting)
	log.Debug(pretty.Sprintf("Loaded configuration from '%s': %# v\n", Server.ConfigFile, Server))

	// Call init hooks
	for _, hook := range hooks {
		hook()
	}
}

// AddHook is used to register initialization code that should run as soon as the config is loaded
func AddHook(hook func()) {
	hooks = append(hooks, hook)
}

func init() {
	viper.SetDefault("musicfolder", filepath.Join(".", "music"))
	viper.SetDefault("datafolder", ".")
	viper.SetDefault("loglevel", "info")
	viper.SetDefault("address", "0.0.0.0")
	viper.SetDefault("port", 4533)
	viper.SetDefault("sessiontimeout", consts.DefaultSessionTimeout)
	viper.SetDefault("scaninterval", time.Minute)
	viper.SetDefault("baseurl", "")
	viper.SetDefault("uiloginbackgroundurl", consts.DefaultUILoginBackgroundURL)
	viper.SetDefault("enabletranscodingconfig", false)
	viper.SetDefault("transcodingcachesize", "100MB")
	viper.SetDefault("imagecachesize", "100MB")
	viper.SetDefault("autoimportplaylists", true)
	viper.SetDefault("enabledownloads", true)

	// Config options only valid for file/env configuration
	viper.SetDefault("searchfullstring", false)
	viper.SetDefault("recentlyaddedbymodtime", false)
	viper.SetDefault("ignoredarticles", "The El La Los Las Le Les Os As O A")
	viper.SetDefault("indexgroups", "A B C D E F G H I J K L M N O P Q R S T U V W X-Z(XYZ) [Unknown]([)")
	viper.SetDefault("probecommand", "ffmpeg %s -f ffmetadata")
	viper.SetDefault("coverartpriority", "embedded, cover.*, folder.*, front.*")
	viper.SetDefault("coverjpegquality", 75)
	viper.SetDefault("uiwelcomemessage", "")
	viper.SetDefault("enablegravatar", false)
	viper.SetDefault("enablefavourites", true)
	viper.SetDefault("enablestarrating", true)
	viper.SetDefault("enableuserediting", true)
	viper.SetDefault("defaulttheme", "Dark")
	viper.SetDefault("gatrackingid", "")
	viper.SetDefault("enablelogredacting", true)
	viper.SetDefault("authrequestlimit", 5)
	viper.SetDefault("authwindowlength", 20*time.Second)

	viper.SetDefault("scanner.extractor", "taglib")
	viper.SetDefault("agents", "lastfm,spotify")
	viper.SetDefault("lastfm.language", "en")
	viper.SetDefault("lastfm.apikey", "")
	viper.SetDefault("lastfm.secret", "")
	viper.SetDefault("spotify.id", "")
	viper.SetDefault("spotify.secret", "")

	// DevFlags. These are used to enable/disable debugging and incomplete features
	viper.SetDefault("devlogsourceline", false)
	viper.SetDefault("devautocreateadminpassword", "")
	viper.SetDefault("devprecachealbumartwork", false)
	viper.SetDefault("devoldcachelayout", false)
	viper.SetDefault("devFastAccessCoverArt", false)
	viper.SetDefault("devactivitypanel", true)
}

func InitConfig(cfgFile string) {
	cfgFile = getConfigFile(cfgFile)
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in local directory with name "navidrome" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName("navidrome")
	}

	_ = viper.BindEnv("port")
	viper.SetEnvPrefix("ND")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if cfgFile != "" && err != nil {
		fmt.Println("Navidrome could not open config file: ", err)
		os.Exit(1)
	}
}

func getConfigFile(cfgFile string) string {
	if cfgFile != "" {
		return cfgFile
	}
	return os.Getenv("ND_CONFIGFILE")
}
