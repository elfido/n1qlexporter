package main

import (
	"fmt"
	"strings"

	"github.com/elfido/n1qlExporter/cbapi"
	"github.com/spf13/viper"
)

type configuration struct {
	clusterName string
	hosts       []string
	useHTTPS    bool
	auth        cbapi.Auth
}

func getConfigurationDefs() []configuration {
	viper.SetConfigType("json")
	viper.AutomaticEnv()
	viper.SetConfigName("settings")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading configuration file: %s\n", err.Error())
		return []configuration{}
	}
	httpuser := viper.GetString("httpuser")
	httppassword := viper.GetString("httppassword")
	useHTTPS := viper.GetBool("usehttps")
	clusters := viper.GetStringMapString("clusters")
	cfg := make([]configuration, len(clusters), len(clusters))
	ndx := 0
	for cluster := range clusters {
		hosts := strings.Split(clusters[cluster], ",")
		cfg[ndx] = configuration{
			clusterName: cluster,
			auth: cbapi.Auth{
				Username: httpuser,
				Password: httppassword,
			},
			useHTTPS: useHTTPS,
			hosts:    hosts,
		}
		ndx++
	}
	return cfg
}
