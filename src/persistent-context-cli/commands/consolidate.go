package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/JaimeStill/persistent-context/persistent-context-cli/pkg"
)

var (
	batchSize   int
	progressive bool
)

var consolidateCmd = &cobra.Command{
	Use:   "consolidate",
	Short: "Consolidation operations",
	Long:  `Commands for testing and monitoring memory consolidation processes.`,
}

var consolidateTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test consolidation with different strategies",
	Long: `Test memory consolidation with various batch sizes and strategies to identify
optimal performance parameters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := pkg.NewClient(viper.GetString("web_url"), viper.GetDuration("timeout"))
		
		// First, get stats to see how many memories we have
		stats, err := client.GetStats()
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}
		
		fmt.Println("System Status:")
		if statsMap, ok := stats.Stats["stats"].(map[string]interface{}); ok {
			if total, ok := statsMap["total_memories"]; ok {
				fmt.Printf("Total memories: %v\n", total)
			}
		}
		
		if progressive {
			fmt.Println("\nTesting progressive consolidation strategy...")
			// For now, we'll just trigger consolidation and measure time
			// In the future, this would test different progressive strategies
		} else {
			fmt.Printf("\nTesting consolidation with batch size: %d\n", batchSize)
			fmt.Println("Note: Batch size testing requires direct pkg access (not yet implemented)")
			fmt.Println("Currently using web service's autonomous consolidation...")
		}
		
		// Trigger consolidation and measure time
		fmt.Println("\nTriggering consolidation...")
		start := time.Now()
		
		result, err := client.TriggerConsolidation()
		if err != nil {
			duration := time.Since(start)
			return fmt.Errorf("consolidation failed after %v: %w", duration, err)
		}
		
		duration := time.Since(start)
		
		// Display results
		fmt.Printf("\nConsolidation completed in: %v\n", duration)
		fmt.Printf("Groups formed: %d\n", result.GroupsFormed)
		fmt.Printf("Groups consolidated: %d\n", result.GroupsConsolidated)
		fmt.Printf("Memories processed: %d\n", result.MemoriesProcessed)
		fmt.Printf("Total memories: %d\n", result.TotalMemories)
		
		// Analyze performance
		if result.GroupsFormed > 0 {
			avgGroupSize := float64(result.MemoriesProcessed) / float64(result.GroupsFormed)
			fmt.Printf("\nAverage group size: %.1f memories\n", avgGroupSize)
			
			if duration > 30*time.Second {
				fmt.Println("\nWARNING: Consolidation took longer than 30 seconds")
				fmt.Println("This suggests batch sizes may be too large")
			}
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(consolidateCmd)
	consolidateCmd.AddCommand(consolidateTestCmd)
	
	consolidateTestCmd.Flags().IntVar(&batchSize, "batch-size", 3, "Number of memories to consolidate per batch")
	consolidateTestCmd.Flags().BoolVar(&progressive, "progressive", false, "Use progressive consolidation strategy")
}