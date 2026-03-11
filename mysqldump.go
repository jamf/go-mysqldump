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

// Close closes the output stream of the dumper if it implements io.Closer.
// It is recommended to call this to ensure files are properly closed.
func (data *Data) Close() error {
	if data.Out == nil {
		return nil
	}
	if out, ok := data.Out.(io.Closer); ok {
		return out.Close()
	}
	return nil
}

// MakeDump performs the full database dump process.
// It opens a new connection to the database using the provided configuration,
// streams the data to the provided io.Writer, and ensures the connection and out is closed.
func (data *Data) MakeDump(out io.Writer) error {
	db, err := sql.Open("mysql", data.Config.FormatDSN())
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	data.Out = out
	data.Connection = db

	defer func() {
		data.Out = nil
		data.Connection = nil
		data.Close()
	}()

	return data.Dump()
}
