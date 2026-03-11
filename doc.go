/*
Create MYSQL dumps in Go without the 'mysqldump' CLI as a dependancy.

# Example

This example uses the mysql driver (https://github.com/go-sql-driver/mysql) to connect to a mysql instance.

	package main

	import (
	    "fmt"
	    "os"

	    "github.com/Naumovets/go-mysqldump"
	    "github.com/go-sql-driver/mysql"
	)

	func main() {
	    // Open connection to database
	    config := mysql.NewConfig()
	    config.User = "your-user"
	    config.Passwd = "your-pw"
	    config.DBName = "your-db"
	    config.Net = "tcp"
	    config.Addr = "your-hostname:your-port"

	    // Initialise dumper with database configuration
	    dumper, err := mysqldump.Init(*config, mysqldump.TableOptions{})
	    if err != nil {
	        fmt.Println("Error initialising dumper:", err)
	        return
	    }

	    // Create output file
	    f, err := os.Create("dump.sql")
	    if err != nil {
	        fmt.Println("Error creating file:", err)
	        return
	    }
	    defer f.Close()

	    // Dump database to file
	    err = dumper.MakeDump(f)
	    if err != nil {
	        fmt.Println("Error dumping:", err)
	        return
	    }
	    fmt.Printf("File is saved to dump.sql")
	}
*/
package mysqldump
