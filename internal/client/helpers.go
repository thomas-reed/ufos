package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/objects"
)

func (c *Client) printUFOList(list []api.UFOItem) error {
	// (Params: output, minwidth, tabwidth, padding, padchar, flags)
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	
	// 2. Print the Header
	fmt.Fprintln(w, "TYPE\tID\tPATH\tSIZE\tTAGS")
	fmt.Fprintln(w, "----\t--\t----\t----\t----")
	for _, ufo := range list {
		metadataBytes, err := crypto.Decrypt(c.MasterKey, ufo.Metadata)
		if err != nil {
			return err
		}
		var metadata objects.ObjectMetadata
		if err = json.Unmarshal(metadataBytes, &metadata); err != nil {
			return fmt.Errorf("Error unmarshalling metadata: %w", err)
		}

		typeName := "FILE"
		sizeStr := fmt.Sprintf("%d B", metadata.SizeBytes)
		if metadata.SizeBytes < 0 {
				typeName = "DIR"
				sizeStr = "-" // Folders don't have a binary size
		}

		// 3. Format tags into a single string
		tagsStr := strings.Join(metadata.Tags, ", ")
		
		// 4. Construct the full path
		fullPath := metadata.Prefix
		if !strings.HasSuffix(fullPath, "/") {
				fullPath += "/"
		}
		fullPath += metadata.Name

		// 5. Write row to the tabwriter (note the \t for tabs)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", 
				typeName, 
				ufo.ID[:8], // Just show the short UUID for brevity
				fullPath, 
				sizeStr, 
				tagsStr,
		)
		clear(metadataBytes)
		metadata = objects.ObjectMetadata{}
	}
	return w.Flush()
}
