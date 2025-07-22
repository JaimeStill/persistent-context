package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/JaimeStill/persistent-context/persistent-context-cli/pkg"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Service management operations",
	Long:  `Commands for checking service health, readiness, and other management operations.`,
}

var serviceHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check service health",
	Long:  `Check the health status of the persistent context web service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := pkg.NewClient(viper.GetString("web_url"), viper.GetDuration("timeout"))
		
		health, err := client.CheckHealth()
		if err != nil {
			return fmt.Errorf("failed to check health: %w", err)
		}
		
		fmt.Println("Health Status:")
		for key, value := range health {
			fmt.Printf("  %s: %v\n", key, value)
		}
		
		return nil
	},
}

var serviceReadyCmd = &cobra.Command{
	Use:   "ready",
	Short: "Check service readiness",
	Long:  `Check the readiness status of the persistent context web service.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := pkg.NewClient(viper.GetString("web_url"), viper.GetDuration("timeout"))
		
		ready, err := client.CheckReady()
		if err != nil {
			return fmt.Errorf("failed to check readiness: %w", err)
		}
		
		fmt.Println("Readiness Status:")
		for key, value := range ready {
			fmt.Printf("  %s: %v\n", key, value)
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceHealthCmd)
	serviceCmd.AddCommand(serviceReadyCmd)
}