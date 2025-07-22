package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/JaimeStill/persistent-context/persistent-context-cli/pkg"
)

var memoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Memory operations",
	Long:  `Commands for inspecting and managing memories in the persistent context system.`,
}

var memoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all memories",
	Long:  `Display a list of all memories in the system with their IDs and timestamps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := pkg.NewClient(viper.GetString("web_url"), viper.GetDuration("timeout"))
		
		memories, err := client.GetMemories(100) // Default limit
		if err != nil {
			return fmt.Errorf("failed to list memories: %w", err)
		}
		
		if len(memories) == 0 {
			fmt.Println("No memories found")
			return nil
		}
		
		// Create a tabwriter for aligned output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTYPE\tCREATED\tCONTENT PREVIEW")
		fmt.Fprintln(w, "---\t----\t-------\t--------------")
		
		for _, mem := range memories {
			created := mem.CreatedAt.Format("2006-01-02 15:04:05")
			preview := mem.Content
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", mem.ID, mem.Type, created, preview)
		}
		w.Flush()
		
		fmt.Printf("\nTotal memories: %d\n", len(memories))
		return nil
	},
}

var memoryShowCmd = &cobra.Command{
	Use:   "show <memory-id>",
	Short: "Show memory details",
	Long:  `Display detailed information about a specific memory including its content and associations.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		memoryID := args[0]
		client := pkg.NewClient(viper.GetString("web_url"), viper.GetDuration("timeout"))
		
		memory, err := client.GetMemory(memoryID)
		if err != nil {
			return fmt.Errorf("failed to get memory: %w", err)
		}
		
		fmt.Printf("Memory ID: %s\n", memory.ID)
		fmt.Printf("Type: %s\n", memory.Type)
		fmt.Printf("Created: %s\n", memory.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Last Accessed: %s\n", memory.AccessedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("\nContent:\n%s\n", memory.Content)
		
		// Display scoring information
		fmt.Printf("\nScoring:\n")
		fmt.Printf("  Base Importance: %.2f\n", memory.Score.BaseImportance)
		fmt.Printf("  Decay Factor: %.2f\n", memory.Score.DecayFactor)
		fmt.Printf("  Access Frequency: %d\n", memory.Score.AccessFrequency)
		fmt.Printf("  Composite Score: %.2f\n", memory.Score.CompositeScore)
		
		// Display associations if any
		if len(memory.AssociationIDs) > 0 {
			fmt.Printf("\nAssociations: %d\n", len(memory.AssociationIDs))
			for _, assocID := range memory.AssociationIDs {
				fmt.Printf("  - %s\n", assocID)
			}
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(memoryCmd)
	memoryCmd.AddCommand(memoryListCmd)
	memoryCmd.AddCommand(memoryShowCmd)
}