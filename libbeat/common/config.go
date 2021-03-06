package common

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	//"path/filepath"
	"runtime"
	"strings"

	"packetbeat/libbeat/common/file"
	"packetbeat/libbeat/logp"
	"github.com/elastic/go-ucfg"
	"github.com/elastic/go-ucfg/cfgutil"
	cfgflag "github.com/elastic/go-ucfg/flag"
	"github.com/elastic/go-ucfg/yaml"
)

var flagStrictPerms = flag.Bool("strict.perms", true, "Strict permission checking on config files")

// IsStrictPerms returns true if strict permission checking on config files is
// enabled.
func IsStrictPerms() bool {
	if !*flagStrictPerms || os.Getenv("BEAT_STRICT_PERMS") == "false" {
		return false
	}
	return true
}

// Config object to store hierarchical configurations into.
// See https://godoc.org/github.com/elastic/go-ucfg#Config
type Config ucfg.Config

// ConfigNamespace storing at most one configuration section by name and sub-section.
type ConfigNamespace struct {
	C map[string]*Config `config:",inline"`
}

type flagOverwrite struct {
	config *ucfg.Config
	path   string
	value  string
}

var configOpts = []ucfg.Option{
	ucfg.PathSep("."),
	ucfg.ResolveEnv,
	ucfg.VarExp,
}

const (
	selectorConfig             = "config"
	selectorConfigWithPassword = "config-with-passwords"
)

var debugBlacklist = MakeStringSet(
	"password",
	"passphrase",
	"key_passphrase",
	"pass",
	"proxy_url",
	"url",
	"urls",
	"host",
	"hosts",
)

// make hasSelector and configDebugf available for unit testing
var hasSelector = logp.HasSelector
var configDebugf = logp.Debug

func NewConfig() *Config {
	return fromConfig(ucfg.New())
}

func NewConfigFrom(from interface{}) (*Config, error) {
	c, err := ucfg.NewFrom(from, configOpts...)
	return fromConfig(c), err
}

func MergeConfigs(cfgs ...*Config) (*Config, error) {
	config := NewConfig()
	for _, c := range cfgs {
		if err := config.Merge(c); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func NewConfigWithYAML(in []byte, source string) (*Config, error) {
	opts := append(
		[]ucfg.Option{
			ucfg.MetaData(ucfg.Meta{Source: source}),
		},
		configOpts...,
	)
	c, err := yaml.NewConfig(in, opts...)
	return fromConfig(c), err
}

func NewFlagConfig(
	set *flag.FlagSet,
	def *Config,
	name string,
	usage string,
) *Config {
	opts := append(
		[]ucfg.Option{
			ucfg.MetaData(ucfg.Meta{Source: "command line flag"}),
		},
		configOpts...,
	)

	var to *ucfg.Config
	if def != nil {
		to = def.access()
	}

	config := cfgflag.ConfigVar(set, to, name, usage, opts...)
	return fromConfig(config)
}

func NewFlagOverwrite(
	set *flag.FlagSet,
	config *Config,
	name, path, def, usage string,
) *string {
	if config == nil {
		panic("Missing configuration")
	}
	if path == "" {
		panic("empty path")
	}

	if def != "" {
		err := config.SetString(path, -1, def)
		if err != nil {
			panic(err)
		}
	}

	f := &flagOverwrite{
		config: config.access(),
		path:   path,
		value:  def,
	}

	if set == nil {
		flag.Var(f, name, usage)
	} else {
		set.Var(f, name, usage)
	}

	return &f.value
}

func LoadFile(path string) (*Config, error) {
	if IsStrictPerms() {
		if err := ownerHasExclusiveWritePerms(path); err != nil {
			return nil, err
		}
	}

	c, err := yaml.NewConfigWithFile(path, configOpts...)
	if err != nil {
		return nil, err
	}

	cfg := fromConfig(c)
	cfg.PrintDebugf("load config file '%v' =>", path)
	return cfg, err
}

func LoadFiles(paths ...string) (*Config, error) {
	merger := cfgutil.NewCollector(nil, configOpts...)
	for _, path := range paths {
		cfg, err := LoadFile(path)
		if err := merger.Add(cfg.access(), err); err != nil {
			return nil, err
		}
	}
	return fromConfig(merger.Config()), nil
}

func (c *Config) Merge(from interface{}) error {
	return c.access().Merge(from, configOpts...)
}

func (c *Config) Unpack(to interface{}) error {
	return c.access().Unpack(to, configOpts...)
}

func (c *Config) Path() string {
	return c.access().Path(".")
}

func (c *Config) PathOf(field string) string {
	return c.access().PathOf(field, ".")
}

func (c *Config) HasField(name string) bool {
	return c.access().HasField(name)
}

func (c *Config) CountField(name string) (int, error) {
	return c.access().CountField(name)
}

func (c *Config) Bool(name string, idx int) (bool, error) {
	return c.access().Bool(name, idx, configOpts...)
}

func (c *Config) String(name string, idx int) (string, error) {
	return c.access().String(name, idx, configOpts...)
}

func (c *Config) Int(name string, idx int) (int64, error) {
	return c.access().Int(name, idx, configOpts...)
}

func (c *Config) Float(name string, idx int) (float64, error) {
	return c.access().Float(name, idx, configOpts...)
}

func (c *Config) Child(name string, idx int) (*Config, error) {
	sub, err := c.access().Child(name, idx, configOpts...)
	return fromConfig(sub), err
}

func (c *Config) SetBool(name string, idx int, value bool) error {
	return c.access().SetBool(name, idx, value, configOpts...)
}

func (c *Config) SetInt(name string, idx int, value int64) error {
	return c.access().SetInt(name, idx, value, configOpts...)
}

func (c *Config) SetFloat(name string, idx int, value float64) error {
	return c.access().SetFloat(name, idx, value, configOpts...)
}

func (c *Config) SetString(name string, idx int, value string) error {
	return c.access().SetString(name, idx, value, configOpts...)
}

func (c *Config) SetChild(name string, idx int, value *Config) error {
	return c.access().SetChild(name, idx, value.access(), configOpts...)
}

func (c *Config) IsDict() bool {
	return c.access().IsDict()
}

func (c *Config) IsArray() bool {
	return c.access().IsArray()
}

func (c *Config) PrintDebugf(msg string, params ...interface{}) {
	selector := selectorConfigWithPassword
	filtered := false
	if !hasSelector(selector) {
		selector = selectorConfig
		filtered = true

		if !hasSelector(selector) {
			return
		}
	}

	debugStr := configDebugString(c, filtered)
	if debugStr != "" {
		configDebugf(selector, "%s\n%s", fmt.Sprintf(msg, params...), debugStr)
	}
}

func (c *Config) Enabled() bool {
	testEnabled := struct {
		Enabled bool `config:"enabled"`
	}{true}

	if c == nil {
		return false
	}
	if err := c.Unpack(&testEnabled); err != nil {
		// if unpacking fails, expect 'enabled' being set to default value
		return true
	}
	return testEnabled.Enabled
}

func fromConfig(in *ucfg.Config) *Config {
	return (*Config)(in)
}

func (c *Config) access() *ucfg.Config {
	return (*ucfg.Config)(c)
}

func (c *Config) GetFields() []string {
	return c.access().GetFields()
}

func (f *flagOverwrite) String() string {
	return f.value
}

func (f *flagOverwrite) Set(v string) error {
	opts := append(
		[]ucfg.Option{
			ucfg.MetaData(ucfg.Meta{Source: "command line flag"}),
		},
		configOpts...,
	)

	err := f.config.SetString(f.path, -1, v, opts...)
	if err != nil {
		return err
	}
	f.value = v
	return nil
}

func (f *flagOverwrite) Get() interface{} {
	return f.value
}

// Validate checks at most one sub-namespace being set.
func (ns *ConfigNamespace) Validate() error {
	if len(ns.C) > 1 {
		return errors.New("more then one namespace configured")
	}
	return nil
}

// Name returns the configuration sections it's name if a section has been set.
func (ns *ConfigNamespace) Name() string {
	for name := range ns.C {
		return name
	}
	return ""
}

// Config return the sub-configuration section if a section has been set.
func (ns *ConfigNamespace) Config() *Config {
	for _, cfg := range ns.C {
		return cfg
	}
	return nil
}

// IsSet returns true if a sub-configuration section has been set.
func (ns *ConfigNamespace) IsSet() bool {
	return len(ns.C) != 0
}

func configDebugString(c *Config, filterPrivate bool) string {
	var bufs []string

	if c.IsDict() {
		var content map[string]interface{}
		if err := c.Unpack(&content); err != nil {
			return fmt.Sprintf("<config error> %v", err)
		}
		if filterPrivate {
			filterDebugObject(content)
		}
		j, _ := json.MarshalIndent(content, "", "  ")
		bufs = append(bufs, string(j))
	}
	if c.IsArray() {
		var content []interface{}
		if err := c.Unpack(&content); err != nil {
			return fmt.Sprintf("<config error> %v", err)
		}
		if filterPrivate {
			filterDebugObject(content)
		}
		j, _ := json.MarshalIndent(content, "", "  ")
		bufs = append(bufs, string(j))
	}

	if len(bufs) == 0 {
		return ""
	}
	return strings.Join(bufs, "\n")
}

func filterDebugObject(c interface{}) {
	switch cfg := c.(type) {
	case map[string]interface{}:
		for k, v := range cfg {
			if debugBlacklist.Has(k) {
				if arr, ok := v.([]interface{}); ok {
					for i := range arr {
						arr[i] = "xxxxx"
					}
				} else {
					cfg[k] = "xxxxx"
				}
			} else {
				filterDebugObject(v)
			}
		}

	case []interface{}:
		for _, elem := range cfg {
			filterDebugObject(elem)
		}
	}
}

// ownerHasExclusiveWritePerms asserts that the current user or root is the
// owner of the config file and that the config file is (at most) writable by
// the owner or root (e.g. group and other cannot have write access).
func ownerHasExclusiveWritePerms(name string) error {
	if runtime.GOOS == "windows" {
		return nil
	}

	_ , err := file.Stat(name)
	if err != nil {
		return err
	}

	//euid := os.Geteuid()
	//fileUID, _ := info.UID()
	//perm := info.Mode().Perm()
	//
	//if fileUID != 0 && euid != fileUID {
	//	return fmt.Errorf(`config file ("%v") must be owned by the beat user `+
	//		`(uid=%v) or root`, name, euid)
	//}
	//
	//// Test if group or other have write permissions.
	//if perm&0022 > 0 {
	//	nameAbs, err := filepath.Abs(name)
	//	if err != nil {
	//		nameAbs = name
	//	}
	//	return fmt.Errorf(`config file ("%v") can only be writable by the `+
	//		`owner but the permissions are "%v" (to fix the permissions use: `+
	//		`'chmod go-w %v')`,
	//		name, perm, nameAbs)
	//}


	return nil
}
