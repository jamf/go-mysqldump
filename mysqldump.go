package mysqldump

import (
	"database/sql"
	"fmt"
	"io"

	"github.com/go-sql-driver/mysql"
)

// Init initializes a new dumper.
// config: configuration of the MySQL database that will be dumped.
// opts: options for filtering tables and configuring the dump process.
func Init(config mysql.Config, opts TableOptions) (*Data, error) {
	return &Data{
		Opts:   opts,
		Config: config,
	}, nil
}

// MakeDump performs the full database dump process.
// It opens a new connection to the database using the provided configuration,
// and streams the data to the provided io.Writer.
func (data *Data) MakeDump(out io.Writer) error {
	if data.Connection == nil {
		db, err := sql.Open("mysql", data.Config.FormatDSN())
		if err != nil {
			return fmt.Errorf("error opening database: %w", err)
		}
		defer db.Close()
		data.Connection = db
	}

	data.Out = out
	defer func() {
		data.Out = nil
	}()

	return data.dump()
}
