package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	yaml "gopkg.in/yaml.v2"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustGetString(flags *pflag.FlagSet, flag string) string {
	s, err := flags.GetString(flag)
	checkErr(err)
	return s
}

func mustGetBool(flags *pflag.FlagSet, flag string) bool {
	b, err := flags.GetBool(flag)
	checkErr(err)
	return b
}

func mustGetUint(flags *pflag.FlagSet, flag string) uint {
	b, err := flags.GetUint(flag)
	checkErr(err)
	return b
}

func generateKey() []byte {
	k, err := settings.GenerateKey()
	checkErr(err)
	return k
}

type cobraFunc func(cmd *cobra.Command, args []string)
type pythonFunc func(cmd *cobra.Command, args []string, data pythonData)

type pythonConfig struct {
	noDB      bool
	allowNoDB bool
}

type pythonData struct {
	hadDB bool
	store *storage.Storage
}

func python(fn pythonFunc, cfg pythonConfig) cobraFunc {
	return func(cmd *cobra.Command, args []string) {
		var err error
		data := pythonData{hadDB: true}
		dbType, _ := getParamB(cmd.Flags(), "db.type")
		path, ok2 := getParamB(cmd.Flags(), "db.url")
		if dbType == "bolt" && !ok2 {
			// set type which != bolt, then clear url
			path = "bolt://filebrowser.db"
		}
		if path == "" {
			dbHost := getParam(cmd.Flags(), "db.host")
			dbPort := getParam(cmd.Flags(), "db.port")
			dbUser := getParam(cmd.Flags(), "db.user")
			dbPwd := getParam(cmd.Flags(), "db.password")
			dbName := getParam(cmd.Flags(), "db.name")
			path = dbType + "://" + dbUser + ":" + dbPwd + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName
		} else if !strings.Contains(path, "://") {
			path = dbType + "://" + path
		}
		data.store, err = storage.CreateStorage(path)
		checkErr(err)
		err = databaseInitialized(data.store)
		data.hadDB = err == nil

		if data.hadDB && cfg.noDB {
			log.Fatal(path + " already exists")
		} else if !data.hadDB && !cfg.noDB && !cfg.allowNoDB {
			log.Fatal(path + " does not exist. Please run 'filebrowser config init' first.")
		}
		fn(cmd, args, data)
	}
}

func marshal(filename string, data interface{}) error {
	fd, err := os.Create(filename)
	checkErr(err)
	defer fd.Close() //nolint: Unhandled error

	switch ext := filepath.Ext(filename); ext {
	case ".json":
		encoder := json.NewEncoder(fd)
		encoder.SetIndent("", "    ")
		return encoder.Encode(data)
	case ".yml", ".yaml": //nolint:goconst
		encoder := yaml.NewEncoder(fd)
		return encoder.Encode(data)
	default:
		return errors.New("invalid format: " + ext)
	}
}

func unmarshal(filename string, data interface{}) error {
	fd, err := os.Open(filename)
	checkErr(err)
	defer fd.Close() //nolint: Unhandled error

	switch ext := filepath.Ext(filename); ext {
	case ".json":
		return json.NewDecoder(fd).Decode(data)
	case ".yml", ".yaml":
		return yaml.NewDecoder(fd).Decode(data)
	default:
		return errors.New("invalid format: " + ext)
	}
}

func jsonYamlArg(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	switch ext := filepath.Ext(args[0]); ext {
	case ".json", ".yml", ".yaml":
		return nil
	default:
		return errors.New("invalid format: " + ext)
	}
}

func cleanUpInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range in {
		result[fmt.Sprintf("%v", k)] = cleanUpMapValue(v)
	}
	return result
}

func cleanUpInterfaceArray(in []interface{}) []interface{} {
	result := make([]interface{}, len(in))
	for i, v := range in {
		result[i] = cleanUpMapValue(v)
	}
	return result
}

func cleanUpMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanUpInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanUpInterfaceMap(v)
	default:
		return v
	}
}

// convertCmdStrToCmdArray checks if cmd string is blank (whitespace included)
// then returns empty string array, else returns the splitted word array of cmd.
// This is to ensure the result will never be []string{""}
func convertCmdStrToCmdArray(cmd string) []string {
	var cmdArray []string
	trimmedCmdStr := strings.TrimSpace(cmd)
	if trimmedCmdStr != "" {
		cmdArray = strings.Split(trimmedCmdStr, " ")
	}
	return cmdArray
}

func databaseInitialized(s *storage.Storage) error {
	version, err := s.Settings.GetVersion()
	if err != nil {
		return errors.ErrNotExist
	}
	if version != 2 {
		return errors.ErrNotExist
	}
	return nil
}
