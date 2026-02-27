package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Loader interface {
	RegisterConfig(interface{}) error
	InitAndValidate() error
}

func NewConfigLoader(prefix string, opts ...Option) Loader {
	cl := &configLoader{
		v:      viper.New(),
		logger: logrus.New(),
	}
	for _, o := range opts {
		o(cl)
	}
	cl.v.SetEnvPrefix(prefix)
	cl.c = &Configuration{
		ConfigPath: "./config",
	}

	if err := cl.RegisterConfig(cl.c); err != nil {
		panic(err)
	}
	return cl
}

type configLoader struct {
	v       *viper.Viper
	configs []interface{}
	logger  *logrus.Logger
	c       *Configuration
}

func (c *configLoader) RegisterConfig(conf interface{}) error {
	c.configs = append(c.configs, conf)
	rv := reflect.ValueOf(conf)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct, but got %t", conf)
	}
	rv = rv.Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		c.logger.WithField("field", rt.Field(i).Name).Debug("Processing config struct field")
		if configTag := rt.Field(i).Tag.Get("config"); configTag != "" {
			c.logger.WithField("configTag", configTag).Debug("Binding conifig var")
			if err := c.v.BindEnv(configTag); err != nil {
				return err
			}

			if flagTag := rt.Field(i).Tag.Get("flag"); flagTag != "" {
				c.logger.WithField("flagTag", flagTag).Debug("Binding flag")
				var name, shorthand, usage string
				flagParts := strings.Split(flagTag, ",")
				name = flagParts[0]
				if len(flagParts) > 1 {
					shorthand = flagParts[1]
				}
				if len(flagParts) > 2 {
					usage = flagParts[2]
				}
				val := NewFlagValue(rv.Field(i))
				f := &pflag.Flag{
					Name:      name,
					Shorthand: shorthand,
					Usage:     usage,
					Value:     val,
					DefValue:  val.String(),
				}
				if val.Type() == "bool" {
					f.NoOptDefVal = "true"
				}
				pflag.CommandLine.AddFlag(f)
				if err := c.v.BindPFlag(configTag, f); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *configLoader) InitAndValidate() (err error) {
	pflag.Parse()
	if err = c.unmarshal(c.c); err != nil {
		return
	}
	if err = c.load(); err != nil {
		return
	}
	if c.v.GetBool("log_config_on_init") {
		c.logger.WithField("config", c.v.AllSettings()).Info("config loaded")
	}
	return
}

func (c *configLoader) load() error {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("config"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	for tag, v := range customValidators {
		if err := validate.RegisterValidation(tag, v); err != nil {
			return err
		}
	}

	configFiles, err := os.ReadDir(c.c.ConfigPath)
	if err != nil {
		return err
	}
	for _, fileInfo := range configFiles {
		if strings.Contains(fileInfo.Name(), "config.") {
			c.v.SetConfigFile(filepath.Join(c.c.ConfigPath, fileInfo.Name()))
			if mergeErr := c.v.MergeInConfig(); mergeErr != nil {
				return mergeErr
			}
		}
	}

	for _, conf := range c.configs {
		if err = c.unmarshal(conf); err == nil {
			if err = validate.Struct(conf); err == nil {
				continue
			}
		}
		return err
	}
	return nil
}

func withTagName(tn string) viper.DecoderConfigOption {
	return func(dc *mapstructure.DecoderConfig) { dc.TagName = tn }
}

func (c *configLoader) unmarshal(dest interface{}) error {
	return c.v.Unmarshal(dest, withTagName("config"))
}
