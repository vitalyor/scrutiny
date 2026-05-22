package models

type ScanOverride struct {
	Device     string   `mapstructure:"device"`
	DeviceType []string `mapstructure:"type"`
	Ignore     bool     `mapstructure:"ignore"`
	Commands   struct {
		MetricsInfoArgs  string `mapstructure:"metrics_info_args"`
		MetricsSmartArgs string `mapstructure:"metrics_smart_args"`
		MetricsTempArgs  string `mapstructure:"metrics_temp_args"`
	} `mapstructure:"commands"`
}
