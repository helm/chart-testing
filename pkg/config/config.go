// Copyright The Helm Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"

	"github.com/helm/chart-testing/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	homeDir, _            = homedir.Dir()
	configSearchLocations = []string{"/etc/ct", path.Join(homeDir, "ct"), "."}
)

type Configuration struct {
	Remote                string        `mapstructure:"remote"`
	TargetBranch          string        `mapstructure:"target-branch"`
	BuildId               string        `mapstructure:"build-id"`
	TillerNamespace       string        `mapstructure:"tiller-namespace"`
	LintConf              string        `mapstructure:"lint-conf"`
	ChartYamlSchema       string        `mapstructure:"chart-yaml-schema"`
	ValidateMaintainers   bool          `mapstructure:"validate-maintainers"`
	CheckVersionIncrement bool          `mapstructure:"check-version-increment"`
	ProcessAllCharts      bool          `mapstructure:"all"`
	Charts                []string      `mapstructure:"charts"`
	ChartRepos            []string      `mapstructure:"chart-repos"`
	ChartDirs             []string      `mapstructure:"chart-dirs"`
	ExcludedCharts        []string      `mapstructure:"excluded-charts"`
	Timeout               time.Duration `mapstructure:"timeout"`
}

func LoadConfiguration(cfgFile string, cmd *cobra.Command, bindFlagsFunc ...func(flagSet *flag.FlagSet, viper *viper.Viper) error) (*Configuration, error) {
	v := viper.New()
	for _, bindFunc := range bindFlagsFunc {
		if err := bindFunc(cmd.Flags(), v); err != nil {
			return nil, errors.Wrap(err, "Error binding flags")
		}
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.SetEnvPrefix("CT")

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("ct")
		for _, searchLocation := range configSearchLocations {
			v.AddConfigPath(searchLocation)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if cfgFile != "" {
			// Only error out for specified config file. Ignore for default locations.
			return nil, errors.Wrap(err, "Error loading config file")
		}
	} else {
		fmt.Println("Using config file: ", v.ConfigFileUsed())
	}

	cfg := &Configuration{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, errors.Wrap(err, "Error unmarshaling configuration")
	}

	if cfg.ProcessAllCharts && len(cfg.Charts) > 0 {
		return nil, errors.New("Specifying both, '--all' and '--charts', is not allowed!")
	}

	isLint := strings.Contains(cmd.Use, "lint")
	chartYamlSchemaPath := cfg.ChartYamlSchema
	if chartYamlSchemaPath == "" {
		var err error
		cfgFile, err = findConfigFile("chart_schema.yaml")
		if err != nil && isLint {
			return nil, errors.New("'chart_schema.yaml' neither specified nor found in default locations")
		}
		cfg.ChartYamlSchema = cfgFile
	}

	lintConfPath := cfg.LintConf
	if lintConfPath == "" {
		var err error
		cfgFile, err = findConfigFile("lintconf.yaml")
		if err != nil && isLint {
			return nil, errors.New("'lintconf.yaml' neither specified nor found in default locations")
		}
		cfg.LintConf = cfgFile
	}

	printCfg(cfg)

	return cfg, nil
}

func printCfg(cfg *Configuration) {
	util.PrintDelimiterLine("-")
	fmt.Println(" Configuration")
	util.PrintDelimiterLine("-")

	e := reflect.ValueOf(cfg).Elem()
	typeOfCfg := e.Type()

	for i := 0; i < e.NumField(); i++ {
		var pattern string
		switch e.Field(i).Kind() {
		case reflect.Bool:
			pattern = "%s: %t\n"
		default:
			pattern = "%s: %s\n"
		}
		fmt.Printf(pattern, typeOfCfg.Field(i).Name, e.Field(i).Interface())
	}

	util.PrintDelimiterLine("-")
}

func findConfigFile(fileName string) (string, error) {
	for _, location := range configSearchLocations {
		filePath := path.Join(location, fileName)
		if util.FileExists(filePath) {
			return filePath, nil
		}
	}
	return "", errors.New(fmt.Sprintf("Config file not found: %s", fileName))
}
