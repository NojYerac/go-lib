package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"source.rad.af/libs/go-lib/pkg/log"
)

type Configuration struct {
	LogConfigOnInit bool   `config:"log_config_on_init"`
	ConfigPath      string `config:"config_dir" flag:"configs,c" validate:"dir"`
}

type Loader interface {
	RegisterConfig(interface{}) error
	InitAndValidate() error
}

func WithArgs(args ...string) Option {
	return func(cl *configLoader) {
		os.Args = append([]string{os.Args[0]}, args...)
	}
}

func WithLogger(l *zerolog.Logger) Option {
	return func(cl *configLoader) {
		cl.logger = l
	}
}

type Option func(*configLoader)

func NewConfigLoader(prefix string, opts ...Option) Loader {
	cl := &configLoader{
		v:      viper.New(),
		logger: log.NewLogger(&log.Configuration{HumanFrendly: true, LogLevel: "fatal"}),
	}
	for _, o := range opts {
		o(cl)
	}
	cl.v.SetEnvPrefix(prefix)
	cl.c = &Configuration{
		ConfigPath: "./config",
	}

	if err := cl.RegisterConfig(cl.c); err != nil {
		cl.logger.Panic().Err(err).Msg("")
	}
	return cl
}

type configLoader struct {
	v       *viper.Viper
	configs []interface{}
	logger  *zerolog.Logger
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
		c.logger.Debug().
			Str("struct", rt.String()).
			Str("fieldName", rt.Field(i).Name).
			Interface("tag", rt.Field(i).Tag).
			Msg("checking config field")
		if configTag := rt.Field(i).Tag.Get("config"); configTag != "" {
			c.logger.Debug().Str("configTag", configTag).Msg("Binding conifig var")
			if err := c.v.BindEnv(configTag); err != nil {
				return err
			}

			if flagTag := rt.Field(i).Tag.Get("flag"); flagTag != "" {
				c.logger.Debug().Str("flagTag", flagTag).Msg("Binding flag")
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

func NewFlagValue(rv reflect.Value) pflag.Value {
	return &flagValue{
		rv:  rv,
		str: fmt.Sprintf("%v", rv.Interface()),
	}
}

type flagValue struct {
	rv  reflect.Value
	str string
}

func (v *flagValue) String() string {
	return v.str
}

func (v *flagValue) Set(val string) error {
	v.str = val
	return nil
}

func (v *flagValue) Type() string {
	return v.rv.Type().String()
}

func withTagName(tn string) viper.DecoderConfigOption {
	return func(dc *mapstructure.DecoderConfig) { dc.TagName = tn }
}

func (c *configLoader) unmarshal(dest interface{}) error {
	return c.v.Unmarshal(dest, withTagName("config"))
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
		c.logger.Info().Interface("config", c.v.AllSettings()).Msg("config loaded")
	}
	return
}

func (c *configLoader) load() error {
	validate := validator.New()

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
