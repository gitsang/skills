package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	MiMo     MiMoConfig     `mapstructure:"mimo"`
	FFmpeg   FFmpegConfig   `mapstructure:"ffmpeg"`
	Server   ServerConfig   `mapstructure:"server"`
	Defaults DefaultsConfig `mapstructure:"defaults"`
}

type MiMoConfig struct {
	APIKey  string        `mapstructure:"api_key"`
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type FFmpegConfig struct {
	Path        string `mapstructure:"path"`
	FFprobePath string `mapstructure:"ffprobe_path"`
}

type ServerConfig struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	UploadLimit string        `mapstructure:"upload_limit"`
	TaskTTL     time.Duration `mapstructure:"task_ttl"`
}

type DefaultsConfig struct {
	SourceLang  string `mapstructure:"source_lang"`
	TargetLang  string `mapstructure:"target_lang"`
	Voice       string `mapstructure:"voice"`
	AudioFormat string `mapstructure:"audio_format"`
	VideoCodec  string `mapstructure:"video_codec"`
	AudioCodec  string `mapstructure:"audio_codec"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("$HOME/.config/video-localization-mimo")
		v.AddConfigPath(".")
	}

	v.AutomaticEnv()
	v.SetEnvPrefix("VIDEO_LOC")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("mimo.base_url", "https://api.xiaomimimo.com/v1")
	v.SetDefault("mimo.timeout", 30*time.Second)

	v.SetDefault("ffmpeg.path", "ffmpeg")
	v.SetDefault("ffmpeg.ffprobe_path", "ffprobe")

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.upload_limit", "500MB")
	v.SetDefault("server.task_ttl", 24*time.Hour)

	v.SetDefault("defaults.source_lang", "zh")
	v.SetDefault("defaults.target_lang", "en")
	v.SetDefault("defaults.voice", "Chloe")
	v.SetDefault("defaults.audio_format", "wav")
	v.SetDefault("defaults.video_codec", "libx264")
	v.SetDefault("defaults.audio_codec", "aac")
}

func (c *Config) Validate() error {
	if c.MiMo.APIKey == "" {
		return fmt.Errorf("MiMo API Key 不能为空")
	}
	return nil
}