package cmd

import (
	// "fmt"
	// "os"

	goflag "flag"

	//"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// TODO: Do we really need viper?
	// "github.com/spf13/viper"
	flag "github.com/spf13/pflag"
	
    // TODO: Do we really need an external package for this?
	// TODO: How much trust do we have in this package?
	 homedir "github.com/mitchellh/go-homedir"
)

var configFile string

var rootCmd = &cobra.Command {
    Use: "reposcrapper",
    Version: "0.0.1-alpha",
}

func Execute() {
    cobra.CheckErr(rootCmd.Execute())
}

func init() {
    flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
    cobra.OnInitialize(initConfig)
}

func initConfig() {
    if configFile != "" {
        viper.SetConfigFile(configFile)
    } else {
        home, err := homedir.Dir()
        cobra.CheckErr(err)

        viper.AddConfigPath(home)
        viper.SetConfigFile(".reposcrapper2")
    }
}