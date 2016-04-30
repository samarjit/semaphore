package util

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/bugsnag/bugsnag-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

var Cookie *securecookie.SecureCookie
var Migration bool
var InteractiveSetup bool
var Upgrade bool

type mySQLConfig struct {
	Hostname string `json:"host"`
	Username string `json:"user"`
	Password string `json:"pass"`
	DbName   string `json:"name"`
}

type configType struct {
	MySQL mySQLConfig `json:"mysql"`
	// Format `:port_num` eg, :3000
	Port       string `json:"port"`
	BugsnagKey string `json:"bugsnag_key"`

	// semaphore stores projects here
	TmpPath string `json:"tmp_path"`

	// cookie hashing & encryption
	CookieHash       string `json:"cookie_hash"`
	CookieEncryption string `json:"cookie_encryption"`
}

var Config *configType

func NewConfig() *configType {
	return &configType{}
}

func init() {
	flag.BoolVar(&InteractiveSetup, "setup", false, "perform interactive setup")
	flag.BoolVar(&Migration, "migrate", false, "execute migrations")
	flag.BoolVar(&Upgrade, "upgrade", false, "upgrade semaphore")
	path := flag.String("config", "", "config path")

	var pwd string
	flag.StringVar(&pwd, "hash", "", "generate hash of given password")

	var printConfig bool
	flag.BoolVar(&printConfig, "printConfig", false, "print example configuration")

	flag.Parse()

	if printConfig {
		cfg := &configType{
			MySQL: mySQLConfig{
				Hostname: "127.0.0.1:3306",
				Username: "root",
				DbName:   "semaphore",
			},
			Port:    ":3000",
			TmpPath: "/tmp/semaphore",
		}
		cfg.GenerateCookieSecrets()

		b, _ := json.MarshalIndent(cfg, "", "\t")
		fmt.Println(string(b))

		os.Exit(0)
	}

	if len(pwd) > 0 {
		password, _ := bcrypt.GenerateFromPassword([]byte(pwd), 11)
		fmt.Println("Generated password: ", string(password))

		os.Exit(0)
	}

	if path != nil && len(*path) > 0 {
		// load
		file, err := os.Open(*path)
		if err != nil {
			panic(err)
		}

		if err := json.NewDecoder(file).Decode(&Config); err != nil {
			fmt.Println("Could not decode configuration!")
			panic(err)
		}
	} else {
		configFile, err := Asset("config.json")
		if err != nil {
			fmt.Println("Cannot Find configuration.")
			os.Exit(1)
		}

		if err := json.Unmarshal(configFile, &Config); err != nil {
			fmt.Println("Could not decode configuration!")
			panic(err)
		}
	}

	if len(os.Getenv("PORT")) > 0 {
		Config.Port = ":" + os.Getenv("PORT")
	}
	if len(Config.Port) == 0 {
		Config.Port = ":3000"
	}

	if len(Config.TmpPath) == 0 {
		Config.TmpPath = "/tmp/semaphore"
	}

	var encryption []byte
	encryption = nil

	hash, _ := base64.StdEncoding.DecodeString(Config.CookieHash)
	if len(Config.CookieEncryption) > 0 {
		encryption, _ = base64.StdEncoding.DecodeString(Config.CookieEncryption)
	}

	Cookie = securecookie.New(hash, encryption)

	stage := ""
	if gin.Mode() == "release" {
		stage = "production"
	} else {
		stage = "development"
	}
	bugsnag.Configure(bugsnag.Configuration{
		APIKey:              Config.BugsnagKey,
		ReleaseStage:        stage,
		NotifyReleaseStages: []string{"production"},
		AppVersion:          Version,
		ProjectPackages:     []string{"github.com/ansible-semaphore/semaphore/**"},
	})
}

func (conf *configType) GenerateCookieSecrets() {
	hash := securecookie.GenerateRandomKey(32)
	encryption := securecookie.GenerateRandomKey(32)

	conf.CookieHash = base64.StdEncoding.EncodeToString(hash)
	conf.CookieEncryption = base64.StdEncoding.EncodeToString(encryption)
}

func (conf *configType) Scan() {
	fmt.Print(" > DB Hostname (default 127.0.0.1:3306): ")
	fmt.Scanln(&conf.MySQL.Hostname)
	if len(conf.MySQL.Hostname) == 0 {
		conf.MySQL.Hostname = "127.0.0.1:3306"
	}

	fmt.Print(" > DB User (default root): ")
	fmt.Scanln(&conf.MySQL.Username)
	if len(conf.MySQL.Username) == 0 {
		conf.MySQL.Username = "root"
	}

	fmt.Print(" > DB Password: ")
	fmt.Scanln(&conf.MySQL.Password)

	fmt.Print(" > DB Name (default semaphore): ")
	fmt.Scanln(&conf.MySQL.DbName)
	if len(conf.MySQL.DbName) == 0 {
		conf.MySQL.DbName = "semaphore"
	}

	fmt.Print(" > Playbook path: ")
	fmt.Scanln(&conf.TmpPath)

	if len(conf.TmpPath) == 0 {
		conf.TmpPath = "/tmp/semaphore"
	}
	conf.TmpPath = path.Clean(conf.TmpPath)
}
