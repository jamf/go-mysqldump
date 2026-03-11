# Go MySQL Dump (Fork)

Create MySQL dumps in Go without the `mysqldump` CLI as a dependency. This is a maintained fork of the original `jamf/go-mysqldump` with enhanced filtering and safety features.

**Module:** `github.com/Naumovets/go-mysqldump`  
**Go Version:** 1.21+

## Key Enhancements (Differences from Original)

- **New API**: Simplified initialization using `mysql.Config` and `TableOptions`.
- **Advanced Filtering**: Filter tables by Schema, Prefix, or Suffix directly in `TableOptions`.
- **View Support**: Correctly dumps View structures without attempting to stream data (prevents restore errors).
- **Virtual Columns**: Automatically detects and excludes `VIRTUAL GENERATED` columns.
- **Binary Data Safety**: Handles `BLOB` and `BINARY` fields by prepending the `_binary` prefix.
- **Managed Connections**: `MakeDump` handles database connection opening and closing internally based on provided config.

## Simple Example

```go
package main

import (
	"fmt"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/Naumovets/go-mysqldump"
)

func main() {
	// 1. Prepare database config
	config := mysql.Config{
		User:   "your-user",
		Passwd: "your-pw",
		DBName: "your-db",
		Addr:   "your-hostname:your-port",
		Net:    "tcp",
	}

	// 2. Set dump options
	opts := mysqldump.TableOptions{
		Schema:       "statistic", // Optional: only tables with schema 'statistic'
		TablePrefix:  "api_",    // Optional: only dump tables starting with 'api_'
		IgnoreTables: []string{"api_logs"},
		LockTables:   true,
	}

	// 3. Init dumper
	dumper, err := mysqldump.Init(config, opts)
	if err != nil {
		fmt.Println("Error initializing dumper:", err)
		return
	}

	// 4. Create output file
	f, err := os.Create("dump.sql")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer f.Close()

	// 5. Execute dump (MakeDump opens connection, performs dump and cleans up)
	if err := dumper.MakeDump(f); err != nil {
		fmt.Println("Error performing dump:", err)
		return
	}

	fmt.Println("Dump completed successfully!")
}
```

## Configuration

### TableOptions
```go
type TableOptions struct {
	Schema           string   // Target database schema (required for information_schema)
	TableSuffix      string   // Filter tables by suffix
	TablePrefix      string   // Filter tables by prefix
	IgnoreTables     []string // List of table names to exclude
	MaxAllowedPacket int      // Max packet size for INSERT statements (default: 4MB)
	LockTables       bool     // Lock tables during dump
}
```

## Testing

Run tests with:
```bash
go test -v ./...
```
