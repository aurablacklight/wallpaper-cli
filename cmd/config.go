package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/user/wallpaper-cli/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure default preferences",
	Long:  `Manage configuration settings for wallpaper-cli.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create default config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfig()
		path := config.GetConfigPath()
		
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file already exists at %s", path)
		}
		
		if err := cfg.Save(path); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		
		fmt.Printf("Created default config at: %s\n", path)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.GetConfigPath()
		cfg, err := config.Load(path)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		
		key := args[0]
		value := getFieldValue(cfg, key)
		if value == nil {
			return fmt.Errorf("unknown config key: %s", key)
		}
		
		fmt.Printf("%s: %v\n", key, value)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.GetConfigPath()
		cfg, err := config.Load(path)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		
		key := args[0]
		value := args[1]
		
		if err := setFieldValue(cfg, key, value); err != nil {
			return err
		}
		
		if err := cfg.Save(path); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		
		fmt.Printf("Set %s = %v\n", key, value)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all config values",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.GetConfigPath()
		cfg, err := config.Load(path)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

// getFieldValue retrieves a field value by name (case-insensitive)
func getFieldValue(cfg *config.Config, key string) interface{} {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	
	key = strings.ToLower(key)
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}
		
		if strings.ToLower(jsonTag) == key || strings.ToLower(field.Name) == key {
			return v.Field(i).Interface()
		}
	}
	
	return nil
}

// setFieldValue sets a field value by name (case-insensitive)
func setFieldValue(cfg *config.Config, key, value string) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	
	key = strings.ToLower(key)
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}
		
		if strings.ToLower(jsonTag) == key || strings.ToLower(field.Name) == key {
			fv := v.Field(i)
			
			switch fv.Kind() {
			case reflect.String:
				fv.SetString(value)
			case reflect.Bool:
				fv.SetBool(value == "true" || value == "1")
			case reflect.Int:
				var intVal int
				fmt.Sscanf(value, "%d", &intVal)
				fv.SetInt(int64(intVal))
			default:
				return fmt.Errorf("cannot set field %s (unsupported type)", key)
			}
			return nil
		}
	}
	
	return fmt.Errorf("unknown config key: %s", key)
}
